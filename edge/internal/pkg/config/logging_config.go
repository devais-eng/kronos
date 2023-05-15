package config

import (
	"devais.it/kronos/internal/pkg/types"
	"github.com/sirupsen/logrus"
)

const (
	defaultLogRedactPasswords = true
	defaultLogFileMaxSize     = 10 * 1024 * 1024 // 10 Megabytes
	defaultLogFileMaxBackups  = 3
	defaultLogFileMaxAgeDays  = 14
)

type LoggingFileConfig struct {
	// Enabled determines if file logging should be enabled.
	Enabled bool

	// Filename is the file to write logs to. Backup logs will be retained in
	// the same directory.
	Filename string

	// MaxSize is the maximum size of the log file before it gets rotated.
	MaxSize types.FileSize

	// MaxBackups is the maximum number of old log files to retain.
	MaxBackups int

	// MaxAgeDays is the maximum number of days to retain old log files.
	MaxAgeDays int

	// UseLocalTime determines if the time used for formatting the timestamps in
	// backup files is the computer's local time. If false, UTC is used.
	UseLocalTime bool

	// Compress determines if the rotated log files should be compressed
	// using gzip.
	Compress bool
}

type LoggingConfig struct {
	// Enabled determines if logging should be enabled.
	Enabled bool

	// ToStderr determines if log output should be directed to
	// standard error. If false, standard output is used instead.
	ToStderr bool

	// UseSyslog determines if logs should be directed to system syslog
	UseSyslog bool

	// Level is the logger level.
	Level logrus.Level

	// ReportCaller sets whether the standard logger will include the calling
	// method as a field.
	ReportCaller bool

	// FormatAsJSON determines if the logging output should be formatted as
	// parsable JSON.
	FormatAsJSON bool

	// PrettyPrint determines whether the JSON output should be pretty
	// printed or not.
	PrettyPrint bool

	// ForceColors disables checking for a TTY before outputting colors.
	// This will force all output to be colored.
	ForceColors bool

	// RedactPasswords determines whether passwords should be hidden
	// from output or not.
	RedactPasswords bool

	// File is the logging file configuration
	File LoggingFileConfig
}

// DefaultLoggingConfig creates a new logging configuration structure
// filled with default options
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Enabled:         true,
		ToStderr:        true,
		UseSyslog:       false,
		Level:           logrus.InfoLevel,
		ReportCaller:    false,
		FormatAsJSON:    false,
		PrettyPrint:     false,
		RedactPasswords: defaultLogRedactPasswords,
		File:            DefaultLoggingFileConfig(),
	}
}

// DefaultLoggingFileConfig creates a new logging file configuration structure
// filed with default parameters
func DefaultLoggingFileConfig() LoggingFileConfig {
	return LoggingFileConfig{
		Enabled:      false,
		Filename:     "",
		MaxSize:      defaultLogFileMaxSize,
		MaxBackups:   defaultLogFileMaxBackups,
		MaxAgeDays:   defaultLogFileMaxAgeDays,
		UseLocalTime: false,
		Compress:     false,
	}
}
