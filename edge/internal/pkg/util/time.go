package util

import (
	"github.com/rotisserie/eris"
	"syscall"
	"time"
)

var (
	startTimeMs uint64
)

func init() {
	// Set application startup timestamp
	startTimeMs = TimestampMs()
}

func NowFuncLocal() time.Time {
	return time.Now().Local()
}

func NowFuncUTC() time.Time {
	return time.Now().UTC()
}

// NowFunc is the default timestamp function of the application
func NowFunc() time.Time {
	return NowFuncUTC()
}

// TimestampMs returns the timestamp in milliseconds using
// NowFunc as the time source
func TimestampMs() uint64 {
	nano := NowFunc().UnixNano()
	return uint64(nano / 1_000_000)
}

// GetUptimeMs returns the application uptime in milliseconds
func GetUptimeMs() uint64 {
	return TimestampMs() - startTimeMs
}

// GetSystemUptimeSeconds returns the system uptime in seconds
func GetSystemUptimeSeconds() (uint64, error) {
	sysInfo := &syscall.Sysinfo_t{}
	err := syscall.Sysinfo(sysInfo)
	if err != nil {
		return 0, eris.Wrap(err, "failed to get system uptime")
	}
	return uint64(sysInfo.Uptime), nil
}
