package models

import (
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
)

type Item struct {
	ID         string      `gorm:"<-:create;type:char(128);unique;primaryKey;index;not null;default:null;" json:"id"`
	Name       string      `gorm:"index;unique;not null;default:null" json:"name"`
	Type       string      `gorm:"not null;default:null" json:"type"`
	CustomerID null.String `gorm:"default null" json:"customer_id"`
	EdgeMac    null.String `gorm:"type:char(17);default null;" json:"edge_mac"`
	Attributes []Attribute `gorm:"-" json:"attributes,omitempty"`

	SyncModel
}

func (Item) TableName() string {
	return ItemsTableName
}

//=============================================================================
// Hooks
//=============================================================================

func (i *Item) AfterFind(tx *gorm.DB) error {
	return nil
}

func (i *Item) BeforeCreate(tx *gorm.DB) error {
	return i.updateVersion(i)
}

func (i *Item) AfterCreate(tx *gorm.DB) error {
	return nil
}

func (i *Item) BeforeSave(tx *gorm.DB) error {
	return i.updateVersion(i)
}

func (i *Item) AfterSave(tx *gorm.DB) error {
	return nil
}

func (i *Item) BeforeUpdate(tx *gorm.DB) error {
	return i.updateVersion(i)
}

func (i *Item) AfterUpdate(tx *gorm.DB) error {
	return nil
}

func (i *Item) BeforeDelete(tx *gorm.DB) error {
	return nil
}

func (i *Item) AfterDelete(tx *gorm.DB) error {
	return nil
}
