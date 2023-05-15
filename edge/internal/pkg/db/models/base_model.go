package models

// BaseModel contains fields common to all entities
type BaseModel struct {
	CreatedAt       uint64 `gorm:"autoCreateTime:milli" json:"created_at"`
	ModifiedAt      uint64 `gorm:"autoUpdateTime:milli" json:"modified_at"`
	CreatedBy       string `gorm:"type:char(20);not null;default:null" json:"created_by"`
	ModifiedBy      string `gorm:"type:char(20);not null;default:null" json:"modified_by"`
	SourceTimestamp uint64 `gorm:"default 0" json:"source_timestamp"`
	// Uncomment the field below to enable soft delete
	//DeletedAt gorm.DeletedAt `gorm:"index" json:",omitempty"`
}
