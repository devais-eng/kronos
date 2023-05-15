package types

type EntityType string

const (
	EntityTypeItem      EntityType = "ITEM"
	EntityTypeAttribute EntityType = "ATTRIBUTE"
	EntityTypeRelation  EntityType = "RELATION"
)
