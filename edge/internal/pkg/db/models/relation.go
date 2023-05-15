package models

import (
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"strings"
)

const (
	relationCompositeIDSeparator = "->"
)

type Relation struct {
	ParentID string `gorm:"<-:create;type:char(128);primaryKey" json:"parent_id"`
	Parent   *Item  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`

	ChildID string `gorm:"<-:create;type:char(128);primaryKey" json:"child_id"`
	Child   *Item  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`

	SyncModel
}

func (Relation) TableName() string {
	return RelationsTableName
}

// CompositeID get the composite relation ID.
func (r *Relation) CompositeID() string {
	return r.ParentID + relationCompositeIDSeparator + r.ChildID
}

// SetCompositeID sets the relation ParentID and ChildID from a composite ID
func (r *Relation) SetCompositeID(id string) error {
	parts := strings.SplitN(id, relationCompositeIDSeparator, 2)
	if len(parts) != 2 {
		return eris.Errorf("failed to parse relation composite ID: '%s'", id)
	}

	r.ParentID = parts[0]
	r.ChildID = parts[1]
	return nil
}

//=============================================================================
// Hooks
//=============================================================================

func (r *Relation) AfterFind(*gorm.DB) error {
	return nil
}

func (r *Relation) BeforeCreate(*gorm.DB) error {
	return r.updateVersion(r)
}

func (r *Relation) AfterCreate(*gorm.DB) error {
	return nil
}

func (r *Relation) BeforeSave(*gorm.DB) error {
	return r.updateVersion(r)
}

func (r *Relation) AfterSave(*gorm.DB) error {
	return nil
}

func (r *Relation) BeforeUpdate(*gorm.DB) error {
	return r.updateVersion(r)
}

func (r *Relation) AfterUpdate(*gorm.DB) error {
	return nil
}

func (r *Relation) BeforeDelete(*gorm.DB) error {
	return nil
}

func (r *Relation) AfterDelete(*gorm.DB) error {
	return nil
}
