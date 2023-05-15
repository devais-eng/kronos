package config

import (
	"devais.it/kronos/internal/pkg/types"
	"devais.it/kronos/internal/pkg/util"
	"time"
)

const (
	defaultDBFile                 = "kronos.db"
	defaultDBSlowQueriesThreshold = 5 * time.Second
	DefaultDBPaginationSize       = 20
	defaultDBCreateBatchSize      = 100
)

type DBConfig struct {
	// URL is the database connection string
	URL string

	// SlowQueriesThreshold is the period after which a query should be
	// considered slow.
	// Queries tagged as 'slow' are automatically logged with warning level.
	SlowQueriesThreshold time.Duration

	// VersionAlgorithm sets the algorithm used to compute
	// records version.
	VersionAlgorithm util.VersionAlgorithm

	// AlwaysAutoMigrate determines if auto migrations should be
	// executed everytime the database is opened
	AlwaysAutoMigrate bool

	// PaginationSize is the default database pagination size.
	// If a method which supports pagination is called with a
	// pagination size of 0, then the value of PaginationSize
	// will be used instead.
	PaginationSize int

	// SoftDelete sets whether soft delete is enabled or not.
	// If enabled, instead of erasing the entire row, records are
	// deleted setting the DeletedAt column to current time.
	// Soft deleted records are not findable with normal queries.
	SoftDelete bool

	// Single create, update and delete operations are performed in transactions by default
	// to ensure database data integrity
	// You can disable this behaviour by setting SkipDefaultTransaction to true
	SkipDefaultTransaction bool

	// UseLocaltime if set to true, database timestamps will be in localtime.
	// Not recommended
	UseLocaltime bool

	// If > 0, each batch operation will be split into sub-batches of size
	// CreateBatchSize. This is to overcome DBMs limitations.
	CreateBatchSize int

	// WALEnabled determines if SQLite Write-ahead Logging is enabled (recommended)
	// https://sqlite.org/pragma.html#pragma_journal_mode
	WALEnabled bool

	// MemTempStoreEnabled determines if SQLite temporary store should be
	// kept in memory.
	// https://www.sqlite.org/pragma.html#pragma_temp_store
	MemTempStoreEnabled bool

	// CacheSize sets the SQLite cache size.
	// https://www.sqlite.org/pragma.html#pragma_cache_size
	CacheSize types.FileSize

	// SynchronousFull will force the SQLite synchronous flag to be FULL (2).
	// If false and WAL is enabled, synchronous flag will be set to NORMAL (1)
	// https://www.sqlite.org/pragma.html#pragma_synchronous
	SynchronousFull bool

	// LockExclusive sets SQLite locking mode to EXCLUSIVE
	// https://sqlite.org/pragma.html#pragma_locking_mode
	//LockExclusive bool

	// BusyTimeout sets SQLite busy timeout
	// Setting multiples of seconds is recommended in order to avoid
	// problems on platforms where usleep is not available (HAVE_USLEEP=0).
	// https://www.sqlite.org/pragma.html#pragma_busy_timeout
	BusyTimeout time.Duration
}

// DefaultDBConfig creates a new database configuration structure
// filed with default parameters
func DefaultDBConfig() DBConfig {
	return DBConfig{
		URL:                    defaultDBFile,
		SlowQueriesThreshold:   defaultDBSlowQueriesThreshold,
		VersionAlgorithm:       util.VersionAlgorithmSha1,
		AlwaysAutoMigrate:      false,
		PaginationSize:         DefaultDBPaginationSize,
		SoftDelete:             false,
		SkipDefaultTransaction: false,
		UseLocaltime:           false,
		CreateBatchSize:        defaultDBCreateBatchSize,
		WALEnabled:             false,
		MemTempStoreEnabled:    false,
		CacheSize:              0,
		SynchronousFull:        false,
		//LockExclusive:        false,
		BusyTimeout: 0,
	}
}
