package logging

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
	"time"
)

const (
	gormTraceFormat = "%s [%s]"
	gormTag         = "[gorm] "
)

type GormLogger struct {
	SlowQueriesThreshold time.Duration
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

func (l *GormLogger) Info(ctx context.Context, s string, i ...interface{}) {
	log.WithContext(ctx).Info(gormTag, fmt.Sprintf(s, i...))
}

func (l *GormLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	log.WithContext(ctx).Warn(gormTag, fmt.Sprintf(s, i...))
}

func (l *GormLogger) Error(ctx context.Context, s string, i ...interface{}) {
	log.WithContext(ctx).Error(gormTag, fmt.Sprintf(s, i...))
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if !log.IsLevelEnabled(log.TraceLevel) {
		return
	}

	fields := log.Fields{}
	elapsed := time.Since(begin)
	sql, rows := fc()
	fields["rows"] = rows

	if err != nil {
		fields[log.ErrorKey] = err
		log.WithContext(ctx).WithFields(fields).Error(gormTag, fmt.Sprintf(gormTraceFormat, sql, elapsed))
	} else if l.SlowQueriesThreshold != 0 && elapsed > l.SlowQueriesThreshold {
		log.WithContext(ctx).WithFields(fields).Warn(gormTag, fmt.Sprintf(gormTraceFormat, sql, elapsed))
	} else {
		log.WithContext(ctx).Debug(gormTag, fmt.Sprintf(gormTraceFormat, sql, elapsed))
	}
}
