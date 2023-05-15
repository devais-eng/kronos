package db

import (
	"devais.it/kronos/internal/pkg/config"
	"devais.it/kronos/internal/pkg/db/models"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

func runMigrations(db *gorm.DB, conf *config.DBConfig) (err error) {
	// Find out if we should auto migrate database models
	shouldMigrate := conf.AlwaysAutoMigrate

	if !shouldMigrate {
		tableNames := models.GetTableNames()
		for _, tableName := range tableNames {
			if !db.Migrator().HasTable(tableName) {
				shouldMigrate = true
				break
			}
		}
	}

	if shouldMigrate {
		err = db.AutoMigrate(models.GetAllModels()...)
		if err != nil {
			return eris.Wrap(err, "failed to auto-migrate")
		}
	}

	// Always run manual migrations
	err = runManualMigrations(db.Migrator())
	if err != nil {
		return eris.Wrap(err, "failed to run manual migrations")
	}

	return
}

func runManualMigrations(migrator gorm.Migrator) error {
	// Add manual migrations here
	return nil
}
