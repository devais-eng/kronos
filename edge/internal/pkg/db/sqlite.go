package db

import (
	"devais.it/kronos/internal/pkg/config"
	"devais.it/kronos/internal/pkg/util"
	"fmt"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"strings"
	"time"
)

func CheckForeignKeysEnabled(db *gorm.DB) (bool, error) {
	var foreignKeysEnabled bool
	tx := db.Raw("PRAGMA foreign_keys").First(&foreignKeysEnabled)
	if tx.Error != nil {
		return false, eris.Wrap(tx.Error, "failed to query for foreign keys status")
	}

	return foreignKeysEnabled, nil
}

func EnableForeignKeys(db *gorm.DB) error {
	tx := db.Exec("PRAGMA foreign_keys = ON")
	if tx.Error != nil {
		return eris.Wrap(tx.Error, "failed to enable foreign keys")
	}
	return nil
}

// configureSqlite configures the SQLite database enabling foreign keys
// and options from the config.DBConfig configuration structure.
func configureSqlite(db *gorm.DB, conf *config.DBConfig) error {
	var tx *gorm.DB

	// Retry enabling foreign keys
	err := util.Retry(10, func() error {
		// Enable foreign keys
		err := EnableForeignKeys(db)
		if err != nil {
			return err
		}

		// Assert foreign keys are enabled
		var foreignKeysEnabled bool
		foreignKeysEnabled, err = CheckForeignKeysEnabled(db)
		if err != nil {
			return err
		}

		if foreignKeysEnabled {
			log.Debug("Foreign keys enabled")
			return nil
		} else {
			return ErrForeignKeysDisabled
		}
	})

	if err != nil {
		return err
	}

	if conf.WALEnabled {
		// Enable Write-Ahead Log
		tx = db.Exec("PRAGMA journal_mode = 'WAL'")
		if tx.Error != nil {
			return eris.Wrap(tx.Error, "failed to set SQLite WAL mode")
		}
	}

	if conf.MemTempStoreEnabled {
		tx = db.Exec("PRAGMA temp_store = 2")
		if tx.Error != nil {
			return eris.Wrap(tx.Error, "failed to set SQLite temp_store")
		}
	}

	if conf.CacheSize > 0 {
		cacheSize := -1 * int64(conf.CacheSize/1024)
		tx = db.Exec(fmt.Sprintf("PRAGMA cache_size = %d", cacheSize))
		if tx.Error != nil {
			return eris.Wrapf(tx.Error, "failed to set SQLite cache size to %d", cacheSize)
		}
	}

	if conf.SynchronousFull || conf.WALEnabled {
		mode := 1
		if conf.SynchronousFull {
			mode = 2
		}

		tx = db.Exec(fmt.Sprintf("PRAGMA synchronous = %d", mode))
		if tx.Error != nil {
			return eris.Wrap(tx.Error, "failed to set SQLite synchronous flag to NORMAL")
		}
	}

	//if conf.LockExclusive {
	//	tx = db.Exec("PRAGMA main.locking_mode = EXCLUSIVE")
	//	if tx.Error != nil {
	//		return eris.Wrap(tx.Error, "failed to set SQLite locking mode")
	//	}
	//}

	if conf.BusyTimeout > 0 {
		tx = db.Exec(fmt.Sprintf("PRAGMA busy_timeout = %d", conf.BusyTimeout.Milliseconds()))
		if tx.Error != nil {
			return eris.Wrap(tx.Error, "failed to set SQLite busy timeout")
		}
	}

	if log.IsLevelEnabled(log.DebugLevel) {
		// Query current settings
		var sqliteVersion string
		var curJournalMode string
		var curTempStore int
		var curCacheSize int64
		var curSynchronous int
		var curLockingMode string
		var curBusyTimeout int64

		tx = db.
			Raw("SELECT sqlite_version()").Scan(&sqliteVersion).
			Raw("PRAGMA journal_mode").Scan(&curJournalMode).
			Raw("PRAGMA temp_store").Scan(&curTempStore).
			Raw("PRAGMA cache_size").Scan(&curCacheSize).
			Raw("PRAGMA synchronous").Scan(&curSynchronous).
			Raw("PRAGMA main.locking_mode").Scan(&curLockingMode).
			Raw("PRAGMA busy_timeout").Scan(&curBusyTimeout)
		if tx.Error != nil {
			return eris.Wrap(tx.Error, "failed to query for SQLite settings")
		}

		fields := log.Fields{
			"sqlite_version":    sqliteVersion,
			"journal_mode":      curJournalMode,
			"temp_store":        curTempStore,
			"cache_size":        curCacheSize,
			"synchronous":       curSynchronous,
			"main.locking_mode": curLockingMode,
			"busy_timeout":      time.Duration(curBusyTimeout) * time.Millisecond,
		}

		log.WithFields(fields).Debug("SQLite current settings")
	}

	if log.IsLevelEnabled(log.TraceLevel) {
		var compileOptions []string
		tx = db.Raw("PRAGMA compile_options;").Scan(&compileOptions)
		if tx.Error != nil {
			return eris.Wrap(tx.Error, "failed to query for SQLite compile options")
		}
		log.Debug("SQLite compile options:\n", strings.Join(compileOptions, "\n"))
	}

	return nil
}
