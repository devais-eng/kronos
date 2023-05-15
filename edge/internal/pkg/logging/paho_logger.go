package logging

import (
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	pahoTag = "[paho] "
)

type PahoLogger struct {
	level log.Level
}

func (l PahoLogger) Println(v ...interface{}) {
	log.WithTime(time.Now()).Log(l.level, pahoTag, fmt.Sprint(v...))
}

func (l PahoLogger) Printf(format string, v ...interface{}) {
	log.WithTime(time.Now()).Log(l.level, pahoTag, fmt.Sprintf(format, v...))
}

func InitPahoLogger() {
	MQTT.ERROR = PahoLogger{level: log.ErrorLevel}
	MQTT.CRITICAL = PahoLogger{level: log.FatalLevel}
	MQTT.WARN = PahoLogger{level: log.WarnLevel}
	MQTT.DEBUG = PahoLogger{level: log.DebugLevel}
}
