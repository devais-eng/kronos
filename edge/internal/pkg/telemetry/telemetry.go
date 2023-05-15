package telemetry

import (
	"strings"
	"sync"
	"time"

	"devais.it/kronos/internal/pkg/db"
	"devais.it/kronos/internal/pkg/services"
	"devais.it/kronos/internal/pkg/util"
	"github.com/distatus/battery"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
)

type BatteryData struct {
	Level        int    `json:"level"`
	Charging     *bool  `json:"charging"`
	StatusString string `json:"status_string"`
}

type Data struct {
	ApplicationUptime     uint64        `json:"application_uptime"`
	SystemUptime          uint64        `json:"system_uptime"`
	IsInDocker            bool          `json:"is_in_docker"`
	Batteries             []BatteryData `json:"batteries"`
	TimestampUTC          uint64        `json:"timestamp_utc"`
	TimestampLocal        uint64        `json:"timestamp_local"`
	LastSyncTs            uint64        `json:"last_sync_ts"`
	LastReceivedMessageTs uint64        `json:"last_received_message_ts"`
	DBFileSize            int64         `json:"db_file_size"`
	ItemsCount            int64         `json:"items_count"`
	AttributesCount       int64         `json:"attributes_count"`
	RelationsCount        int64         `json:"relations_count"`
}

type telemetryState struct {
	sync.RWMutex
	lastSyncTs            uint64
	lastReceivedMessageTs uint64
}

var (
	state = telemetryState{}
)

func SetLastSyncTs() {
	state.Lock()
	defer state.Unlock()
	state.lastSyncTs = util.TimestampMs()
}

func GetLastSyncTs() uint64 {
	state.RLock()
	defer state.RUnlock()
	return state.lastSyncTs
}

func SetLastMessageReceivedTs() {
	state.Lock()
	defer state.Unlock()
	state.lastReceivedMessageTs = util.TimestampMs()
}

func GetLastMessageReceivedTs() uint64 {
	state.RLock()
	defer state.RUnlock()
	return state.lastReceivedMessageTs
}

func Get() (*Data, error) {
	state.RLock()
	defer state.RUnlock()

	systemUptimeSecs, err := util.GetSystemUptimeSeconds()
	if err != nil {
		return nil, eris.Wrap(err, "failed to query for system uptime")
	}
	systemUptimeMs := systemUptimeSecs * 1000

	dbFileSize, err := db.Size()
	if err != nil {
		return nil, eris.Wrap(err, "failed to query for DB file size")
	}

	itemsCount, err := services.GetItemsCount()
	if err != nil {
		return nil, eris.Wrap(err, "failed to get items count")
	}

	attributesCount, err := services.GetAttributesCount()
	if err != nil {
		return nil, eris.Wrap(err, "failed to get attributes count")
	}

	relationsCount, err := services.GetRelationsCount()
	if err != nil {
		return nil, eris.Wrap(err, "failed to get relations count")
	}

	timestampUtc := uint64(time.Now().UTC().UnixNano() / 1_000_000)
	timestampLocal := uint64(time.Now().Local().UnixNano() / 1_000_000)

	data := &Data{
		ApplicationUptime:     util.GetUptimeMs(),
		SystemUptime:          systemUptimeMs,
		IsInDocker:            util.IsInDocker(),
		Batteries:             getBatteries(),
		TimestampUTC:          timestampUtc,
		TimestampLocal:        timestampLocal,
		LastSyncTs:            state.lastSyncTs,
		LastReceivedMessageTs: state.lastReceivedMessageTs,
		DBFileSize:            dbFileSize,
		ItemsCount:            itemsCount,
		AttributesCount:       attributesCount,
		RelationsCount:        relationsCount,
	}

	return data, nil
}

func convertBatteryToData(bat *battery.Battery, batErr error) (*BatteryData, error) {
	switch err := batErr.(type) {
	case battery.ErrPartial:
		if err.Current != nil {
			return nil, eris.Wrap(err.Current, "failed to read battery current capacity")
		}

		if err.Full != nil {
			return nil, eris.Wrap(err.Full, "failed to read battery full capacity")
		}

		level := util.ClampInt(int(bat.Current/bat.Full*100), 0, 100)

		data := &BatteryData{
			Level:        level,
			StatusString: strings.ToLower(bat.State.String()),
		}

		// Set charging to null if we weren't able to read battery status
		if err.State != nil || bat.State == battery.Unknown {
			data.Charging = nil
		} else {
			charging := bat.State == battery.Charging
			data.Charging = &charging
		}

		return data, nil
	default:
		return nil, err
	}
}

func getBatteries() []BatteryData {
	var batteriesData []BatteryData

	bats, err := battery.GetAll()

	switch e := err.(type) {
	case battery.ErrFatal:
		log.Warnf(
			"Failed to read batteries information: %s",
			eris.ToString(err, false),
		)
	case battery.Errors:
		if len(e) == len(bats) {
			batteriesData = make([]BatteryData, 0, len(bats))
			for i, bat := range bats {
				batErr := e[i]
				data, err := convertBatteryToData(bat, batErr)
				if err == nil {
					batteriesData = append(batteriesData, *data)
				} else {
					log.Warnf("Failed to create battery data: %s", eris.ToString(err, false))
				}
			}
		} else {
			log.Warn("Possible battery library bug: batteries count and errors count differs")
		}
	}

	return batteriesData
}
