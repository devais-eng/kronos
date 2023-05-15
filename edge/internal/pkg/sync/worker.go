package sync

import (
	"devais.it/kronos/internal/pkg/telemetry"
	"sync"
	"time"

	"devais.it/kronos/internal/pkg/sync/messages"

	"devais.it/kronos/internal/pkg/config"
	"devais.it/kronos/internal/pkg/db"
	"devais.it/kronos/internal/pkg/db/models"
	"devais.it/kronos/internal/pkg/logging"
	"devais.it/kronos/internal/pkg/services"
	"devais.it/kronos/internal/pkg/util"
	"github.com/cenkalti/backoff/v4"
	"github.com/getsentry/sentry-go"
	"github.com/looplab/fsm"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	// FSM states
	stateConnecting  = "connecting"
	stateSubscribing = "subscribing"
	statePubVersions = "pubVersions"
	stateDequeueing  = "dequeueing"
	stateStopped     = "stopped"

	// FSM events
	eventConnected         = "connected"
	eventSubscribed        = "subscribed"
	eventVersionsPublished = "versionsPublished"
	eventDisconnected      = "disconnected"
	eventStop              = "stop"
)

var (
	allStates = []string{
		stateConnecting,
		stateSubscribing,
		statePubVersions,
		stateDequeueing,
		stateStopped,
	}

	fsmEvents = fsm.Events{
		{Name: eventConnected, Src: []string{stateConnecting}, Dst: stateSubscribing},
		{Name: eventSubscribed, Src: []string{stateSubscribing}, Dst: statePubVersions},
		{Name: eventVersionsPublished, Src: []string{statePubVersions}, Dst: stateDequeueing},
		{Name: eventDisconnected, Src: allStates, Dst: stateConnecting},
		{Name: eventStop, Src: allStates, Dst: stateStopped},
	}
)

var (
	pingChan = make(chan string, 1)

	pingRequestMsg  = "ping"
	pingResponseMsg = "pong"
)

type Worker struct {
	conf *config.SyncConfig
	fsm  *fsm.FSM

	// Mutex for concurrent callbacks
	cbMutex sync.Mutex

	eventCond    util.ChanCond
	stoppedCond  util.ChanCond
	client       Client
	timeMutex    sync.Mutex
	lastSyncTime time.Time

	syncCallbacks []SyncCallback
	syncCbMutex   sync.RWMutex

	// Prometheus collectors
	cyclesCounter       prometheus.Counter
	errorsCounter       prometheus.Counter
	msgsReceivedCounter prometheus.Counter
	panicsCounter       prometheus.Counter
	pubEventsCounter    prometheus.Counter
}

func NewWorker(conf *config.SyncConfig) (*Worker, error) {
	var client Client

	if conf.ClientType == config.SyncClientMQTT {
		mqttClient, err := NewMQTTClient(conf)
		if err != nil {
			return nil, eris.Wrap(err, "Failed to create sync worker client")
		}
		client = mqttClient
	} else {
		return nil, eris.Errorf("Unknown sync client type: %v", conf.ClientType)
	}

	worker := &Worker{
		conf:   conf,
		fsm:    fsm.NewFSM(stateConnecting, fsmEvents, fsm.Callbacks{}),
		client: client,
		// Metrics
		cyclesCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "kronos_worker_cycles_total",
			Help: "The total number of worker cycles",
		}),
		errorsCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "kronos_worker_errors_total",
			Help: "The number of worker errors",
		}),
		msgsReceivedCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "kronos_worker_messages_received_total",
			Help: "The number of received messages",
		}),
		panicsCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "kronos_worker_panics_total",
			Help: "The number of worker panics",
		}),
		pubEventsCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "kronos_published_events_total",
			Help: "The number of published events",
		}),
	}

	return worker, nil
}

func (w *Worker) Collectors() []prometheus.Collector {
	return []prometheus.Collector{
		w.cyclesCounter,
		w.errorsCounter,
		w.msgsReceivedCounter,
		w.panicsCounter,
		w.pubEventsCounter,
	}
}

func (w *Worker) RefreshMetrics() error {
	return nil
}

func (w *Worker) AddSyncCallback(cb SyncCallback) {
	w.syncCbMutex.Lock()
	defer w.syncCbMutex.Unlock()

	w.syncCallbacks = append(w.syncCallbacks, cb)
}

func (w *Worker) SignalSyncEvent(message messages.Sync) {
	w.syncCbMutex.RLock()
	defer w.syncCbMutex.RUnlock()

	for _, cb := range w.syncCallbacks {
		cb(message)
	}
}

func (w *Worker) Start() error {
	c := w.client

	c.SetConnectionCallback(w.onConnected)
	c.SetDisconnectionCallback(w.onDisconnected)
	c.SetSyncCallback(w.onSyncMessage)
	c.SetCommandCallback(w.onServerCommandMessage)

	go w.workerRoutine()
	log.Info("Sync worker started")
	return nil
}

func (w *Worker) Stop() error {
	w.stoppedCond.InitBuffered(1)

	w.fsmEvent(eventStop)

	if err := w.client.Disconnect(); err != nil {
		return err
	}

	select {
	case <-w.stoppedCond.Wait():
		log.Info("Sync worker stopped")
	case <-time.NewTicker(w.conf.StopTimeout).C:
		log.Warn("Sync worker stop timed out")
	}

	return nil
}

func PingWorker(timeout time.Duration) error {
	select {
	case <-time.NewTicker(timeout).C:
		return ErrTimeout
	case pingChan <- pingRequestMsg:
		select {
		case <-time.NewTicker(timeout).C:
			return ErrTimeout
		case reply := <-pingChan:
			if reply == pingResponseMsg {
				log.Trace("Sync worker ping success")
				return nil
			}
			return eris.New("Invalid ping response")
		}
	}
}

func (w *Worker) signalEvent() {
	// Send event signal
	w.eventCond.Signal()
}

func (w *Worker) fsmEvent(event string, args ...interface{}) {
	if err := w.fsm.Event(event, args...); err != nil {
		logging.Error(err, "Failed to process sync worker FSM event")
	}

	w.signalEvent()
}

func (w *Worker) onConnected() {
	// Lock to avoid concurrent changes between fsm.Current and event firing
	w.cbMutex.Lock()
	defer w.cbMutex.Unlock()

	state := w.fsm.Current()

	if state == stateConnecting {
		log.Info("Sync worker connected")
		w.fsmEvent(eventConnected)
	} else {
		log.Debugf("Connected while in state %s", state)
	}
}

func (w *Worker) onDisconnected(err error) {
	w.cbMutex.Lock()
	defer w.cbMutex.Unlock()

	state := w.fsm.Current()
	if state != stateConnecting {
		if err != nil {
			log.Error("Sync worker disconnected, error: ", err)
		}
		w.fsmEvent(eventDisconnected)
	} else {
		log.Debugf("Disconnected while in state %s", state)
	}
}

func (w *Worker) dequeueEvents() error {
	// Dequeue events in a transaction
	return db.DB().Transaction(func(tx *gorm.DB) error {
		count, err := services.TryDequeueEvents(tx, w.conf.MaxEvents, w.publishEvents)
		if err != nil {
			return err
		}
		log.Infof("%d events dequeued", count)

		// Metrics
		w.pubEventsCounter.Add(float64(count))

		return nil
	})
}

func (w *Worker) onSyncMessage(message messages.Sync) {
	defer func() {
		if err := recover(); err != nil {
			w.handlePanicRecovered(err)
		}
	}()

	telemetry.SetLastMessageReceivedTs()

	if err := w.handleSyncMessage(message); err != nil {
		logging.Error(err, "Failed to handle sync message")
	} else {
		log.Debug("Sync message handled")

		// Update last sync telemetry
		telemetry.SetLastSyncTs()

		// Check delta time between last sync message to handle bursts
		// This is to avoid dequeueing events while a lot of sync messages
		// are being received, locking the database
		w.timeMutex.Lock()
		defer w.timeMutex.Unlock()

		if time.Since(w.lastSyncTime) >= w.conf.Backoff.InitialInterval {
			// Signal to immediately dequeue events
			log.Debug("Sync handled, signaling event...")
			w.signalEvent()
		}

		w.lastSyncTime = time.Now()

		// Signal event to registered callbacks
		w.SignalSyncEvent(message)
	}

	// Metrics
	w.msgsReceivedCounter.Inc()
}

func (w *Worker) onServerCommandMessage(message *messages.ServerCommand) {
	defer func() {
		if err := recover(); err != nil {
			w.handlePanicRecovered(err)
		}
	}()

	// Update message received telemetry
	telemetry.SetLastMessageReceivedTs()

	// Metrics
	w.msgsReceivedCounter.Inc()

	log.Debugf("Received server command: %s - %s", message.UUID, message.CommandType)

	var err error
	response := &messages.CommandResponse{
		UUID: message.UUID,
	}

	switch message.CommandType {
	case messages.CommandGetVersion:
		response.Body, err = w.handleVersionCommand(message.EntityType, message.EntityID)
	case messages.CommandGetAllVersions:
		err = w.publishVersions()
	case messages.CommandGetEntity:
		response.Body, err = w.handleGetEntityCommand(message.EntityType, message.EntityID)
	case messages.CommandGetTelemetry:
		response.Body, err = w.handleGetTelemetryCommand()
	default:
		err = eris.Errorf("Unknown command: '%s'", message.CommandType)
	}

	if err != nil {
		response.Success = false
		response.Error = eris.ToString(err, false)
	} else {
		response.Success = true
	}

	err = w.client.PublishCommandResponse(response)
	if err != nil {
		logging.Error(err, "Failed to publish command response")
	}
}

func (w *Worker) connect() error {
	log.Debug("Sync worker connecting...")
	return w.client.Connect()
}

func (w *Worker) subscribe() error {
	log.Debug("Sync worker subscribing...")
	return w.client.Subscribe()
}

func (w *Worker) publishVersions() error {
	log.Debug("Sync worker publishing versions...")
	return w.client.PublishVersions()
}

func (w *Worker) publishEvents(events []models.Event) error {
	eventMessages := make([]messages.Event, 0, len(events))

	for _, event := range events {
		var eventBody map[string]interface{}

		err := event.UnmarshalBody(&eventBody)
		if err != nil {
			return eris.Wrap(err, "failed to unmarshal event body")
		}

		eventMsg := messages.Event{
			ID:          event.ID,
			EntityType:  event.EntityType,
			EntityID:    event.EntityID,
			TriggeredBy: event.TriggeredBy,
			TxUUID:      event.TxUUID,
			TxType:      event.EventType,
			TxLen:       event.TxLen,
			TxIndex:     event.TxIndex,
			Timestamp:   event.Timestamp,
			Body:        eventBody,
		}

		eventMessages = append(eventMessages, eventMsg)
	}

	return w.client.PublishEvents(eventMessages)
}

func (w *Worker) doWork(ticker *time.Ticker, backOff *backoff.ExponentialBackOff) bool {
	state := w.fsm.Current()

	log.Debugf("Sync worker state: %s", state)

	if state == stateStopped {
		log.Info("Sync worker routine stopped")
		return false
	}

	var err error

	switch state {
	case stateConnecting:
		err = w.connect()
		if err == nil {
			w.onConnected()
		}
	case stateSubscribing:
		err = w.subscribe()
		if err == nil {
			log.Info("Sync worker subscribed")
			w.fsmEvent(eventSubscribed)
		}
	case statePubVersions:
		if w.conf.PublishVersions {
			err = w.publishVersions()
			if err == nil {
				log.Info("Sync worker published versions")
				w.fsmEvent(eventVersionsPublished)
			}
		} else {
			// Skip version publishing if not enabled by configuration
			err = nil
			w.fsmEvent(eventVersionsPublished)
		}
	case stateDequeueing:
		err = w.dequeueEvents()
		if err == nil {
			log.Debug("Sync worker dequeued events")
		}
	default:
		log.Errorf("Unhandled sync worker FSM state: %s", state)
	}

	if err != nil {
		if eris.Is(err, gorm.ErrRecordNotFound) {
			ticker.Reset(backOff.InitialInterval)
		} else {
			if eris.Is(err, ErrNotConnected) {
				log.Error("Sync worker not connected")
			} else {
				logging.Error(err, "Sync worker error")
			}
			nextInterval := backOff.NextBackOff()
			if nextInterval == backOff.Stop {
				log.Trace("Max backoff interval reached")
				nextInterval = backOff.MaxInterval
			}

			ticker.Reset(nextInterval)

			// Metrics
			w.errorsCounter.Inc()
		}
	} else {
		// No error, immediately try to handle next state
		backOff.Reset()
		ticker.Reset(w.conf.MinSleepTime)
		log.Trace("Sync worker no error")
	}

	return true
}

func (w *Worker) workerRoutine() {
	// Panic handling
	defer func() {
		if err := recover(); err != nil {
			w.handlePanicRecovered(err)
			// Restart panicked goroutine
			go w.workerRoutine()
		} else {
			w.stoppedCond.Signal()
		}
	}()

	backOff := w.conf.Backoff.NewBackoff()
	ticker := time.NewTicker(backOff.InitialInterval)

	if !w.doWork(ticker, backOff) {
		return
	}

	for {
		select {
		case <-w.eventCond.Wait():
			if !w.doWork(ticker, backOff) {
				return
			}
		case <-ticker.C:
			if !w.doWork(ticker, backOff) {
				return
			}
		case req := <-pingChan:
			if req == pingRequestMsg {
				// Reply to ping
				select {
				case pingChan <- pingResponseMsg:
				default:
				}
			}
		}

		// Stats
		w.cyclesCounter.Add(1)
	}
}

func (w *Worker) handlePanicRecovered(err interface{}) {
	log.Error("Sync worker recovered from panic. Error: ", err)

	if w.conf.Sentry.Enabled {
		// Notify panic to Sentry
		hub := sentry.CurrentHub()
		eventID := hub.Recover(err)
		if eventID != nil {
			if w.conf.Sentry.WaitForDelivery {
				hub.Flush(w.conf.Sentry.DeliveryTimeout)
			}
			log.Infof("Panic error sent to Sentry. Event ID: '%s'", *eventID)
		}
	}

	// Metrics
	w.panicsCounter.Inc()
}
