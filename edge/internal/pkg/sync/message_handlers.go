package sync

import (
	"devais.it/kronos/internal/pkg/constants"
	"devais.it/kronos/internal/pkg/db"
	"devais.it/kronos/internal/pkg/db/models"
	"devais.it/kronos/internal/pkg/services"
	"devais.it/kronos/internal/pkg/sync/messages"
	"devais.it/kronos/internal/pkg/telemetry"
	"devais.it/kronos/internal/pkg/types"
	"devais.it/kronos/internal/pkg/util"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
)

var (
	ErrInvalidEntityType = eris.New("Invalid entity type")
	ErrInvalidAction     = eris.New("Invalid action")
)

func (w *Worker) handleVersionCommand(entityType types.EntityType, entityID string) (map[string]interface{}, error) {
	if entityType == types.EntityTypeItem {
		version, err := services.GetItemVersion(entityID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"version": version,
		}, nil
	} else if entityType == types.EntityTypeAttribute {
		version, err := services.GetAttributeVersion(entityID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"version": version,
		}, nil
	} else {
		return nil, eris.Errorf("Unknown entity type: '%s'", string(entityType))
	}
}

func (w *Worker) handleGetEntityCommand(entityType types.EntityType, entityID string) (map[string]interface{}, error) {
	if entityType == types.EntityTypeItem {
		item, err := services.GetItemByID(entityID)
		if err != nil {
			return nil, err
		}
		return util.StructToJSONMap(item)
	} else if entityType == types.EntityTypeAttribute {
		attribute, err := services.GetAttributeByID(entityID)
		if err != nil {
			return nil, err
		}
		return util.StructToJSONMap(attribute)
	} else {
		return nil, eris.Errorf("Unknown entity type: '%s'", string(entityType))
	}
}

func (w *Worker) handleGetTelemetryCommand() (map[string]interface{}, error) {
	telData, err := telemetry.Get()
	if err != nil {
		return nil, eris.Wrap(err, "failed to get telemetry data for command response")
	}
	return util.StructToJSONMap(telData)
}

/*
func (w *Worker) checkForeignKeys() error {
	fksEnabled, err := db.CheckForeignKeysEnabled(db.DB())
	if err != nil {
		log.Error("Failed to query for foreign keys status")
		return err
	}
	if !fksEnabled {
		log.Error("Foreign keys are disabled, trying to enable them...")
		err = db.EnableForeignKeys(db.DB())
		if err != nil {
			log.Error("Failed to enable foreign keys")
			return err
		}
	}
	return nil
}
*/

func (w *Worker) handleSyncMessage(message messages.Sync) error {
	// TODO: Understand why this LOC is necessary.
	// Apparently without doing this, cascade deletes won't work. And calling
	// w.checkForeignKeys won't make it work either. If we query SQLite to check
	// whether foreign keys are enabled or not, it will say: "of course they are",
	// but cascade deletes will still work randomly. The only thing that fixes the problem
	// is calling this very exact code.
	// This might be an ORM bug, or some sort of UB inside the C binding.
	tx := db.DB().Exec("PRAGMA foreign_keys = ON")
	if tx.Error != nil {
		return eris.Wrap(tx.Error, "failed to enable foreign keys")
	}
	// End of workaround

	// Optimization for single entries
	if len(message) == 1 {
		err := db.GetHardDeleteTx(db.DB()).Transaction(func(tx *gorm.DB) error {
			ctx := &db.TxContext{Tx: tx}
			return syncEntry(ctx, &message[0])
		})

		if err != nil {
			return eris.Wrap(err, "Failed to handle sync message")
		}

		return nil
	}

	entriesMap := make(map[types.EntityType]map[messages.SyncAction][]*messages.SyncEntry)

	for i := 0; i < len(message); i++ {
		entry := &message[i]

		actions, ok := entriesMap[entry.EntityType]
		if !ok {
			actions = make(map[messages.SyncAction][]*messages.SyncEntry)
			entriesMap[entry.EntityType] = actions
		}

		entries, ok := actions[entry.Action]
		if !ok {
			entries = make([]*messages.SyncEntry, 0, 2)
			actions[entry.Action] = entries
		}

		actions[entry.Action] = append(entries, entry)
	}

	typesOrder := []types.EntityType{
		types.EntityTypeItem,
		types.EntityTypeAttribute,
		types.EntityTypeRelation,
	}

	actionsOrder := []messages.SyncAction{
		messages.SyncActionCreate,
		messages.SyncActionUpdate,
		messages.SyncActionDelete,
	}

	err := db.GetHardDeleteTx(db.DB()).Transaction(func(tx *gorm.DB) error {
		txLen := 0
		for _, entry := range message {
			txLen += db.CalcTxLength(entry.Payload)
		}

		//defer db.CloseRows(tx)
		ctx := &db.TxContext{
			TxUUID: uuid.NewString(),
			TxLen:  txLen,
			Tx:     tx,
		}

		for _, entityType := range typesOrder {
			for _, action := range actionsOrder {
				entries := entriesMap[entityType][action]

				for _, entry := range entries {
					err := syncEntry(ctx, entry)

					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		return eris.Wrap(err, "Failed to handle sync message")
	}

	return nil
}

func syncEntry(ctx *db.TxContext, entry *messages.SyncEntry) error {
	mb := constants.ModifiedBySyncName

	// Declare models
	item := &models.Item{}
	attribute := &models.Attribute{}
	relation := &models.Relation{}

	var model interface{}
	var syncPolicy null.String

	var err error

	switch entry.EntityType {
	case types.EntityTypeItem:
		item.ID = entry.EntityID
		model = item
		syncPolicy, err = services.GetItemSyncPolicy(item.ID)
	case types.EntityTypeAttribute:
		attribute.ID = entry.EntityID
		model = attribute
		syncPolicy, err = services.GetAttributesSyncPolicy(attribute.ID)
	case types.EntityTypeRelation:
		err = relation.SetCompositeID(entry.EntityID)
		if err != nil {
			return err
		}
		model = relation
		syncPolicy, err = services.GetRelationSyncPolicy(relation.ParentID, relation.ChildID)
	default:
		return ErrInvalidEntityType
	}

	// Check the result of GetSyncPolicy methods
	if err != nil && !eris.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	log.Debugf("Entity '%s' sync policy is '%s'", entry.EntityID, syncPolicy.ValueOrZero())

	if syncPolicy.ValueOrZero() == constants.SyncPolicyDontSync {
		// Skip synchronization of this entry
		log.Debugf(
			"Entity '%s' sync policy is '%s', skipping synchronization...",
			entry.EntityID,
			constants.SyncPolicyDontSync,
		)
		return nil
	}

	if entry.Action != messages.SyncActionDelete && entry.Version != "" {
		var version string
		// Fetch version and compare it with the incoming one
		versionQuery := ctx.Tx.Model(model).Select("version")
		if entry.EntityType == types.EntityTypeRelation {
			versionQuery = versionQuery.Where("parent_id = ? AND child_id = ?", relation.ParentID, relation.ChildID)
		} else {
			versionQuery = versionQuery.Where("id = ?", entry.EntityID)
		}
		err = versionQuery.First(&version).Error
		if err != nil && !eris.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if version == entry.Version {
			// Local entity already synchronized to requested version. Skip synchronization
			return nil
		}
	}

	if entry.Action == messages.SyncActionCreate ||
		entry.Action == messages.SyncActionUpdate {
		// NOTE: Maybe it would be better to move this assignment
		// down to database hooks in order to have all version checks
		// in the same file.
		if entry.Payload != nil {
			entry.Payload["version"] = entry.Version
			entry.Payload["sync_version"] = entry.Version
		}
	}

	if entry.Action == messages.SyncActionCreate {
		err = createEntry(ctx, entry)
		if err != nil {
			return err
		}
	} else if entry.Action == messages.SyncActionUpdate {
		if entry.EntityType == types.EntityTypeRelation {
			_, err = services.GetRelation(relation.ParentID, relation.ChildID)
		} else {
			entry.Payload["id"] = entry.EntityID

			if entry.EntityType == types.EntityTypeItem {
				err = services.UpdateItemTx(ctx, entry.Payload, mb)
			} else if entry.EntityType == types.EntityTypeAttribute {
				err = services.UpdateAttributeTx(ctx, entry.Payload, mb)
			} else {
				err = ErrInvalidEntityType
			}
		}

		if eris.Is(err, gorm.ErrRecordNotFound) {
			// Create the entity
			err = createEntry(ctx, entry)
		}

		if err != nil {
			return err
		}
	} else if entry.Action == messages.SyncActionDelete {
		if entry.EntityType == types.EntityTypeRelation {
			err = services.DeleteRelationTx(ctx, relation.ParentID, relation.ChildID, mb)
		} else {
			err = services.DeleteByIDTx(ctx, entry.EntityID, model, mb)
		}
		if err != nil {
			if eris.Is(err, gorm.ErrRecordNotFound) {
				log.Debugf("%s '%s' does not exist.", entry.EntityType, entry.EntityID)
				return nil
			}
			return err
		}
	} else {
		return ErrInvalidAction
	}

	return nil
}

func createEntry(ctx *db.TxContext, entry *messages.SyncEntry) error {
	mb := constants.ModifiedBySyncName

	// Declare models
	item := &models.Item{}
	attribute := &models.Attribute{}
	relation := &models.Relation{}

	var model interface{}

	switch entry.EntityType {
	case types.EntityTypeItem:
		model = item
	case types.EntityTypeAttribute:
		model = attribute
	case types.EntityTypeRelation:
		model = relation
	default:
		return ErrInvalidEntityType
	}

	err := util.JSONToStruct(entry.Payload, model)
	if err != nil {
		return err
	}

	switch entry.EntityType {
	case types.EntityTypeItem:
		if item.ID == "" {
			item.ID = entry.EntityID
		}
		item.SyncVersion = null.StringFrom(entry.Version)
		err = services.BatchCreateItemsTx(ctx, []models.Item{*item}, mb)
	case types.EntityTypeAttribute:
		attr := model.(*models.Attribute)
		if attr.ID == "" {
			attr.ID = entry.EntityID
		}
		attr.SyncVersion = null.StringFrom(entry.Version)
		err = services.BatchCreateAttributesTx(ctx, []models.Attribute{*attr}, mb)
	case types.EntityTypeRelation:
		rel := model.(*models.Relation)
		if rel.ParentID == "" || rel.ChildID == "" {
			err = rel.SetCompositeID(entry.EntityID)
			if err != nil {
				return err
			}
		}
		rel.Version = entry.Version
		rel.SyncVersion = null.StringFrom(entry.Version)
		err = services.BatchCreateRelationsTx(ctx, []models.Relation{*rel}, mb)
	default:
		err = ErrInvalidEntityType
	}

	if err != nil {
		return err
	}

	return nil
}
