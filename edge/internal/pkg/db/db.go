package db

import (
	"devais.it/kronos/internal/pkg/config"
	"devais.it/kronos/internal/pkg/logging"
	"devais.it/kronos/internal/pkg/util"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	db *gorm.DB

	ErrForeignKeysDisabled = eris.New("Foreign keys are disabled")
	ErrInvalidPagination   = eris.New("Invalid pagination")
	ErrMissingID           = eris.New("Missing ID field")
)

func OpenDB(conf *config.DBConfig) error {
	var err error

	gormConfig := &gorm.Config{
		SkipDefaultTransaction: conf.SkipDefaultTransaction,
		CreateBatchSize:        conf.CreateBatchSize,
		Logger: &logging.GormLogger{
			SlowQueriesThreshold: conf.SlowQueriesThreshold,
		},
		// Set now function as UTC
		NowFunc: util.NowFuncUTC,
	}

	if conf.UseLocaltime {
		gormConfig.NowFunc = util.NowFuncLocal
	}

	db, err = gorm.Open(sqlite.Open(conf.URL), gormConfig)
	if err != nil {
		return eris.Wrap(err, "failed to init db")
	}

	errRef := &err

	defer func() {
		// Close opened DB instance if an error occurred
		if *errRef != nil {
			err := Close()
			if err != nil {
				log.Errorf("Faled to close DB: %v", err)
			}
			db = nil
		}
	}()

	err = configureSqlite(db, conf)
	if err != nil {
		return eris.Wrap(err, "failed to configure SQLite db")
	}

	err = runMigrations(db, conf)
	if err != nil {
		return eris.Wrap(err, "migrations failed")
	}

	return nil
}

func DB() *gorm.DB {
	return db
}

func Close() error {
	sqlDb, err := db.DB()
	if err != nil {
		return eris.Wrap(err, "failed to get sql.DB")
	}
	return sqlDb.Close()
}

// Size returns the database size
func Size() (int64, error) {
	query := `SELECT page_count * page_size FROM pragma_page_count(), pragma_page_size()`

	var size int64
	tx := db.Raw(query).First(&size)
	if tx.Error != nil {
		return 0, tx.Error
	}
	return size, nil
}

func IsSoftDeleteEnabled() bool {
	return viper.GetBool("db.softDelete")
}

//=============================================================================
// CRUD operations
//
// Read operations will be run directly on the shared database instance
// while operations that modify data requires a transaction object.
//=============================================================================

func isIDNil(id interface{}) bool {
	return id == nil || id == "" || id == 0
}

func CheckID(id interface{}) error {
	if isIDNil(id) {
		return ErrMissingID
	}
	return nil
}

func Create(tx *gorm.DB, model interface{}) error {
	err := tx.Create(model).Error
	if err != nil {
		return eris.Wrap(err, "create failed")
	}
	return nil
}

func GetByID(id interface{}, dest interface{}) error {
	if isIDNil(id) {
		return ErrMissingID
	}
	tx := db.Where("id = ?", id).First(dest)
	if tx.Error != nil {
		return eris.Wrap(tx.Error, "get by ID query failed")
	}
	return nil
}

// Paginate performs pagination on a transaction.
// If page and pageSize arguments are 0, default pagination is applied.
func Paginate(tx *gorm.DB, page, pageSize int) (*gorm.DB, error) {
	offset := 0
	limit := 0

	if page < 0 || pageSize < 0 {
		return nil, ErrInvalidPagination
	}

	if page == 0 {
		page = 1
	}

	if pageSize == 0 {
		pageSize = viper.GetInt("db.paginationSize")
		if pageSize <= 0 {
			pageSize = config.DefaultDBPaginationSize
		}
	}

	offset = (page - 1) * pageSize
	limit = pageSize

	return tx.Offset(offset).Limit(limit), nil
}

// GetAll returns all the record of a table applying some pagination first
func GetAll(dest interface{}, page, pageSize int) error {
	tx, err := Paginate(db, page, pageSize)
	if err != nil {
		return eris.Wrap(err, "pagination failed")
	}
	err = tx.Find(dest).Error
	if err != nil {
		return eris.Wrap(err, "get all query failed")
	}
	return nil
}

// Update updates a record with a map containing changes to apply.
// Each map entry should have the column name as the key and the
// updated value as the value.
func Update(tx *gorm.DB, model interface{}, patch map[string]interface{}) error {
	if id, ok := patch["id"]; ok {
		if isIDNil(id) {
			return ErrMissingID
		}
		updateTx := tx.
			Model(model).
			Where("id = ?", id).
			Updates(patch)
		if updateTx.Error != nil {
			return eris.Wrap(updateTx.Error, "update failed")
		}
		if updateTx.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	}
	return ErrMissingID
}

// checkDeleteResult checks if a delete transaction is failed.
// A delete operation is considered failed when a transaction
// error occurred or the number of rows deleted is 0 (i.e. no records were found)
func checkDeleteResult(tx *gorm.DB) error {
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetHardDeleteTx returns a transaction to use for delete
// operations which will perform hard deletes when soft delete is
// enabled.
func GetHardDeleteTx(tx *gorm.DB) *gorm.DB {
	if IsSoftDeleteEnabled() {
		return tx.Unscoped()
	}
	return tx
}

// Delete deletes a record
func Delete(tx *gorm.DB, model interface{}) error {
	err := checkDeleteResult(tx.Delete(model))
	if err != nil {
		return eris.Wrap(err, "delete transaction failed")
	}
	return nil
}

// DeleteByID deletes an entity by ID, where the ID can be a column
// of any type named 'id'.
func DeleteByID(tx *gorm.DB, id interface{}, model interface{}) error {
	if isIDNil(id) {
		return ErrMissingID
	}

	err := checkDeleteResult(tx.Where("id = ?", id).Delete(model))
	if err != nil {
		return eris.Wrap(err, "delete by ID transaction failed")
	}
	return nil
}

// Count returns the number of records of the given model
func Count(model interface{}) (count int64, err error) {
	tx := db.Model(model).Count(&count)
	if tx.Error != nil {
		err = eris.Wrap(tx.Error, "count query failed")
	}
	return
}

// CloseRows closes rows returned from a query
func CloseRows(tx *gorm.DB) error {
	rows, err := tx.Rows()
	if err != nil {
		return eris.Wrap(err, "failed to get transaction rows")
	}
	err = rows.Close()
	if err != nil {
		return eris.Wrap(err, "failed to close transaction rows")
	}
	return nil
}

// IntegrityCheck performs an integrity check on the SQLite database.
// An error is returned when the query fails, otherwise integrity errors
// are returned as a slice.
func IntegrityCheck() ([]string, error) {
	var result []string
	tx := db.Raw("PRAGMA integrity_check").Find(&result)
	if tx.Error != nil {
		return nil, eris.Wrap(tx.Error, "integrity check failed")
	}
	return result, nil
}

//=============================================================================
