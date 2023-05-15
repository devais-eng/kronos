// this package contains methods to probe for application's health

package health

import (
	"devais.it/kronos/internal/pkg/db"
	"devais.it/kronos/internal/pkg/logging"
	"devais.it/kronos/internal/pkg/sync"
	"time"
)

const (
	timeout = 5 * time.Second
)

type CheckResult struct {
	Error          error       `json:"error"`
	AdditionalInfo interface{} `json:"additional_info"`
}

func Check() (result *CheckResult) {
	result = &CheckResult{}

	fksEnabled, err := db.CheckForeignKeysEnabled(db.DB())
	if err != nil {
		result.Error = err
		return
	}

	if !fksEnabled {
		result.Error = db.ErrForeignKeysDisabled
		result.AdditionalInfo = "Foreign keys are not enabled"
		return
	}

	// SQLite integrity check
	integrityResult, err := db.IntegrityCheck()
	if err != nil || len(integrityResult) > 0 {
		result.Error = err
		result.AdditionalInfo = integrityResult
		return
	}

	err = sync.PingWorker(timeout)
	if err != nil {
		logging.Error(err, "Failed to ping sync worker")
		result.Error = err
		return
	}

	return
}
