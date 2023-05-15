package models

import (
	"devais.it/kronos/internal/pkg/constants"
	"devais.it/kronos/internal/pkg/util"
	"github.com/spf13/viper"
	"gopkg.in/guregu/null.v4"
)

type SyncModel struct {
	BaseModel
	SyncPolicy  null.String `gorm:"default:null" json:"sync_policy,omitempty"`
	Version     string      `gorm:"type:char(40);not null;default:null" json:"version"`
	SyncVersion null.String `gorm:"type:char(40);default:null" json:"sync_version"`
}

func (s *SyncModel) updateVersion(model interface{}) error {
	if s.Version == "" {
		algo := util.VersionAlgorithm(viper.GetInt("db.versionAlgo"))
		version, err := util.GenerateVersionChecksum(model, algo)
		if err != nil {
			return err
		}
		s.Version = version
	}

	if s.SyncVersion.IsZero() && s.ModifiedBy == constants.ModifiedBySyncName {
		// Set Sync version
		s.SyncVersion = null.StringFrom(s.Version)
	}

	return nil
}
