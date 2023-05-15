package logging

import (
	"io"
	"log/syslog"
	"os"

	"devais.it/kronos/internal/pkg/config"
	"devais.it/kronos/internal/pkg/constants"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	logrus_syslog "github.com/sirupsen/logrus/hooks/syslog"
	"gopkg.in/natefinch/lumberjack.v2"
)

func lumberjackLogger(c *config.LoggingFileConfig) io.Writer {
	// Lumberjack file size parameter should be expressed in Megabytes
	maxSizeMb := int(c.MaxSize / 1024 / 1024)

	return &lumberjack.Logger{
		Filename:   c.Filename,
		MaxSize:    maxSizeMb,
		MaxAge:     c.MaxAgeDays,
		MaxBackups: c.MaxBackups,
		LocalTime:  c.UseLocalTime,
		Compress:   c.Compress,
	}
}

func Setup(conf *config.LoggingConfig) error {
	if !conf.Enabled {
		// Disable output
		log.SetOutput(io.Discard)
		return nil
	}

	var formatter log.Formatter

	if conf.FormatAsJSON {
		formatter = &log.JSONFormatter{
			PrettyPrint: conf.PrettyPrint,
		}
	} else {
		formatter = &log.TextFormatter{
			FullTimestamp:    true,
			ForceColors:      conf.ForceColors,
			ForceQuote:       false,
			DisableQuote:     true,
			QuoteEmptyFields: true,
		}
	}

	log.SetFormatter(formatter)
	log.SetLevel(conf.Level)
	log.SetReportCaller(conf.ReportCaller)

	var outputs []io.Writer

	if conf.ToStderr {
		outputs = append(outputs, os.Stderr)
	} else {
		outputs = append(outputs, os.Stdout)
	}

	if conf.UseSyslog {
		hook, err := logrus_syslog.NewSyslogHook("", "", syslog.LOG_INFO, constants.AppName)
		if err != nil {
			return eris.Wrap(err, "Failed to get syslog hook")
		}
		log.AddHook(hook)
		log.Info("Logging hooked to syslog")
	}

	if conf.File.Enabled && conf.File.Filename != "" {
		outputs = append(outputs, lumberjackLogger(&conf.File))
	}

	log.SetOutput(io.MultiWriter(outputs...))

	return nil
}

func Error(err error, args ...interface{}) {
	fields := log.Fields{}
	fields[log.ErrorKey] = eris.ToJSON(err, true)
	log.WithFields(fields).Error(args...)
}

func Panic(err error, args ...interface{}) {
	fields := log.Fields{}
	fields[log.ErrorKey] = eris.ToJSON(err, true)
	log.WithFields(fields).Panic(args...)
}
