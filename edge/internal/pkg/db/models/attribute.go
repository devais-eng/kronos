package models

import (
	"gorm.io/gorm"
)

type Attribute struct {
	ID        string `gorm:"<-:create;type:char(128);unique;primaryKey;index;not null;default:null" json:"id"`
	Name      string `gorm:"not null;default:null;uniqueIndex:idx_item_id_name" json:"name"`
	Type      string `gorm:"index;not null;default:null" json:"type"`
	Value     string `gorm:"default:null;" json:"value"`
	ValueType string `gorm:"default:null" json:"value_type"`

	ItemID string `gorm:"<-:create;char(128);not null;default:null;uniqueIndex:idx_item_id_name" json:"item_id"`
	Item   *Item  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`

	SyncModel
}

//=============================================================================
// Hooks
//=============================================================================

func (Attribute) TableName() string {
	return AttributesTableName
}

func (a *Attribute) AfterFind(*gorm.DB) error {
	return nil
}

func (a *Attribute) BeforeCreate(*gorm.DB) error {
	return a.updateVersion(a)
}

func (a *Attribute) AfterCreate(*gorm.DB) error {
	return nil
}

func (a *Attribute) BeforeSave(*gorm.DB) error {
	return a.updateVersion(a)
}

func (a *Attribute) AfterSave(*gorm.DB) error {
	return nil
}

func (a *Attribute) BeforeUpdate(*gorm.DB) error {
	return a.updateVersion(a)
}

func (a *Attribute) AfterUpdate(*gorm.DB) error {
	return nil
}

func (a *Attribute) BeforeDelete(*gorm.DB) error {
	return nil
}

func (a *Attribute) AfterDelete(*gorm.DB) error {
	return nil
}
