package services

import (
	"devais.it/kronos/internal/pkg/db"
	"devais.it/kronos/internal/pkg/db/models"
	"devais.it/kronos/internal/pkg/types"
	"devais.it/kronos/internal/pkg/util"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func GetLastEvent() (*models.Event, error) {
	event := &models.Event{}
	tx := db.DB().Limit(1).Order("id DESC, timestamp").First(event)
	if tx.Error != nil {
		return nil, eris.Wrap(tx.Error, "failed to get last event")
	}
	return event, nil
}

func GetLastEvents(tx *gorm.DB, limit int) ([]models.Event, error) {
	var events []models.Event
	tx = tx.Limit(limit).Order("id DESC, timestamp").Find(&events)
	if tx.Error != nil {
		return nil, eris.Wrap(tx.Error, "failed to get last events")
	}
	return events, nil
}

func GetFirstEvent() (*models.Event, error) {
	event := &models.Event{}
	tx := db.DB().Limit(1).Order("id ASC, timestamp").First(event)
	if tx.Error != nil {
		return nil, eris.Wrap(tx.Error, "failed to get first event")
	}
	return event, nil
}

func GetFirstEvents(tx *gorm.DB, limit int) ([]models.Event, error) {
	var events []models.Event
	tx = tx.Limit(limit).Order("id ASC, timestamp").Find(&events)
	if tx.Error != nil {
		return nil, eris.Wrap(tx.Error, "failed to get first events")
	}
	return events, nil
}

func GetEvent(
	eventType types.EventType,
	entityType types.EntityType,
	entityID string) (*models.Event, error) {
	event := &models.Event{}
	tx := db.DB().Where(
		"event_type = ? AND entity_type = ? AND entity_id = ?",
		eventType,
		entityType,
		entityID,
	).First(event)
	if tx.Error != nil {
		return nil, eris.Wrap(tx.Error, "failed to get event")
	}
	return event, nil
}

func UpdateEvent(tx *gorm.DB, patch map[string]interface{}) error {
	err := db.Update(tx, &models.Event{}, patch)
	if err != nil {
		return eris.Wrap(err, "failed to update event")
	}
	return nil
}

func DeleteEvent(tx *gorm.DB, event *models.Event) error {
	err := tx.Delete(event).Error
	if err != nil {
		return eris.Wrap(err, "failed to delete event")
	}
	return nil
}

func DeleteEventsByType(
	tx *gorm.DB,
	entityType types.EntityType,
	entityID string,
	triggeredBy string,
	types []types.EventType) error {
	tx = tx.
		Where(
			"entity_type = ? AND entity_id = ? AND triggered_by = ? AND event_type IN ?",
			entityType,
			entityID,
			triggeredBy,
			types,
		).
		Delete(&models.Event{})
	if tx.Error != nil {
		return eris.Wrap(tx.Error, "failed to delete events by type")
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func GetEventsCount() (count int64, err error) {
	count, err = db.Count(&models.Event{})
	if err != nil {
		err = eris.Wrap(err, "failed to get events count")
	}
	return
}

func patchEventBody(body, patch string) (string, error) {
	var err error
	var bodyMap map[string]interface{}
	var patchMap map[string]interface{}

	err = json.Unmarshal([]byte(body), &bodyMap)
	if err != nil {
		return "", eris.Wrap(err, "failed to unmarshal event body JSON")
	}

	err = json.Unmarshal([]byte(patch), &patchMap)
	if err != nil {
		return "", eris.Wrap(err, "failed to unmarshal event patch JSON")
	}

	// Update body map
	for k, v := range patchMap {
		bodyMap[k] = v
	}

	res, err := json.Marshal(bodyMap)
	if err != nil {
		return "", eris.Wrap(err, "failed to marshal event body JSON")
	}

	return string(res), nil
}

func PublishEvent(
	ctx *db.TxContext,
	eventType types.EventType,
	entityType types.EntityType,
	entityID string,
	triggeredBy string,
	bodyModel interface{}) error {
	timestampMs := util.TimestampMs()

	bodyBytes, err := json.Marshal(bodyModel)
	if err != nil {
		return eris.Wrap(err, "failed to marshal event body")
	}

	body := string(bodyBytes)

	if eventType == types.EventEntityDeleted {
		// If event is entity deleted, remove previous creation
		err = DeleteEventsByType(
			ctx.Tx,
			entityType,
			entityID,
			triggeredBy,
			[]types.EventType{
				types.EventEntityCreated,
			},
		)

		if eris.Is(err, gorm.ErrRecordNotFound) {
			log.Debugf(
				"Previous '%s' or '%s' not found for '%s' event ['%s', '%s']",
				types.EventEntityCreated,
				types.EventEntityUpdated,
				eventType,
				entityType,
				entityID,
			)
		} else if err != nil {
			return eris.Wrap(err, "failed to delete old events")
		} else {
			return nil
		}
	} else if eventType == types.EventEntityUpdated {
		// Try to update existing create event
		lastEvent, err := GetEvent(types.EventEntityCreated, entityType, entityID)
		if err != nil || lastEvent == nil {
			lastEvent, err = GetEvent(types.EventEntityUpdated, entityType, entityID)
		}

		// Check if event is triggered by the same entity
		if err == nil && lastEvent.TriggeredBy == triggeredBy {
			newBody, err := patchEventBody(lastEvent.Body, body)
			if err != nil {
				return eris.Wrap(err, "failed to patch event body")
			}

			eventPatch := map[string]interface{}{
				"id":        lastEvent.ID,
				"timestamp": timestampMs,
				"body":      newBody,
			}

			err = UpdateEvent(ctx.Tx, eventPatch)
			if err != nil {
				return eris.Wrap(err, "failed to update event")
			}

			return nil
		}
	}

	event := &models.Event{
		EventType:   eventType,
		EntityID:    entityID,
		EntityType:  entityType,
		TriggeredBy: triggeredBy,
		TxUUID:      ctx.TxUUID,
		TxLen:       ctx.TxLen,
		TxIndex:     int(ctx.TxIndex),
		Timestamp:   timestampMs,
		Body:        body,
	}

	err = db.Create(ctx.Tx, event)
	if err != nil {
		return eris.Wrap(err, "failed to publish event")
	}

	log.Debugf(
		"Event %s, [%s, %s] published to queue",
		eventType,
		entityType,
		entityID,
	)

	// Increment transaction index
	ctx.IncTxIndex()

	return nil
}

func TryDequeueEvent(tx *gorm.DB, fn func(event *models.Event) error) error {
	firstEvent, err := GetFirstEvent()

	if eris.Is(err, gorm.ErrRecordNotFound) {
		return gorm.ErrRecordNotFound
	}

	if err != nil {
		return eris.Wrap(err, "failed to get first event")
	}

	err = fn(firstEvent)
	if err != nil {
		return eris.Wrap(err, "failed to process first event")
	}

	err = DeleteEvent(tx, firstEvent)
	if err != nil {
		return eris.Wrap(err, "failed to dequeue first event")
	}

	log.Debugf(
		"Event %d, %s, [%s, %s] dequeued",
		firstEvent.ID,
		firstEvent.EventType,
		firstEvent.EntityType,
		firstEvent.EntityID,
	)

	return nil
}

func TryDequeueEvents(tx *gorm.DB, limit int, fn func(events []models.Event) error) (int, error) {
	events, err := GetFirstEvents(tx, limit)
	if err != nil {
		return 0, err
	}
	eventsCount := len(events)

	if eventsCount == 0 {
		return 0, gorm.ErrRecordNotFound
	}

	err = fn(events)
	if err != nil {
		return 0, err
	}

	err = tx.Delete(events).Error
	if err != nil {
		return 0, err
	}

	return eventsCount, nil
}
