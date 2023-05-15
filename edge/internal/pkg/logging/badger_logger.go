package logging

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
)

type BadgerLogger struct{}

func (BadgerLogger) log(level log.Level, format string, args ...interface{}) {
	str := strings.TrimSuffix(fmt.Sprintf(format, args...), "\n")
	log.WithFields(nil).Log(level, "[badger] ", str)
}

func (l BadgerLogger) Errorf(format string, args ...interface{}) {
	l.log(log.ErrorLevel, format, args...)
}

func (l BadgerLogger) Warningf(format string, args ...interface{}) {
	l.log(log.WarnLevel, format, args...)
}

func (l BadgerLogger) Infof(format string, args ...interface{}) {
	l.log(log.InfoLevel, format, args...)
}

func (l BadgerLogger) Debugf(format string, args ...interface{}) {
	l.log(log.DebugLevel, format, args...)
}
