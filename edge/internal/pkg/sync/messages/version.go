package messages

import (
	"devais.it/kronos/internal/pkg/db/models"
	"devais.it/kronos/internal/pkg/types"
)

type EntityVersions []models.EntityVersion

type Versions struct {
	Timestamp uint64                              `json:"timestamp"`
	Versions  map[types.EntityType]EntityVersions `json:"versions"`
}
