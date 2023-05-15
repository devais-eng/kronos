// package models contains database GORM models

package models

import (
	"devais.it/kronos/internal/pkg/types"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"gopkg.in/guregu/null.v4"
	"reflect"
)

const (
	ItemsTableName      = "items"
	AttributesTableName = "attributes"
	RelationsTableName  = "relations"
	EventsTableName     = "events_queue"
)

// GetAllModels returns an empty list of all database models
func GetAllModels() []interface{} {
	return []interface{}{
		&Item{},
		&Attribute{},
		&Relation{},
		&Event{},
	}
}

// GetTableNames returns a list of all database tables' names
func GetTableNames() []string {
	return []string{
		ItemsTableName,
		AttributesTableName,
		RelationsTableName,
		EventsTableName,
	}
}

func GetTableName(entityType types.EntityType) (string, error) {
	switch entityType {
	case types.EntityTypeItem:
		return ItemsTableName, nil
	case types.EntityTypeAttribute:
		return AttributesTableName, nil
	case types.EntityTypeRelation:
		return RelationsTableName, nil
	default:
		return "", eris.Errorf("Unknown entity type: '%s'", entityType)
	}
}

// GetEntityType returns the types.EntityType of a given entity
func GetEntityType(entity interface{}) types.EntityType {
	switch entity.(type) {
	case Item:
		return types.EntityTypeItem
	case *Item:
		return types.EntityTypeItem
	case []Item:
		return types.EntityTypeItem
	case Attribute:
		return types.EntityTypeAttribute
	case *Attribute:
		return types.EntityTypeAttribute
	case []Attribute:
		return types.EntityTypeAttribute
	case Relation:
		return types.EntityTypeRelation
	case *Relation:
		return types.EntityTypeRelation
	case []Relation:
		return types.EntityTypeRelation
	default:
		log.Panicf("Unknown entity type: %v", reflect.TypeOf(entity))
		return ""
	}
}

func GetEntityID(entity interface{}) string {
	switch e := entity.(type) {
	case Item:
		return e.ID
	case *Item:
		return e.ID
	case Attribute:
		return e.ID
	case *Attribute:
		return e.ID
	case Relation:
		return e.ParentID
	case *Relation:
		return e.ParentID
	default:
		log.Panicf("Unknown entity type: %v", reflect.TypeOf(entity))
		return ""
	}
}

func GetEntitySyncPolicy(entity interface{}) (null.String, error) {
	switch e := entity.(type) {
	case Item:
		return e.SyncPolicy, nil
	case *Item:
		return e.SyncPolicy, nil
	case Attribute:
		return e.SyncPolicy, nil
	case *Attribute:
		return e.SyncPolicy, nil
	case Relation:
		return e.SyncPolicy, nil
	case *Relation:
		return e.SyncPolicy, nil
	case map[string]interface{}:
		if policy, ok := e["sync_policy"]; ok {
			switch p := policy.(type) {
			case string:
				return null.StringFrom(p), nil
			default:
			}
		}
		return null.String{}, nil
	default:
		return null.String{}, eris.Errorf("Unknown entity type: %v", reflect.TypeOf(entity))
	}
}
