package sync

import (
	"bytes"
	"devais.it/kronos/internal/pkg/logging"
	"github.com/dgraph-io/badger"
	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
)

// BadgerStore implements a Paho messages store (mqtt.Store)
// using Badger as the key/value backend.
// Badger is fast, reliable and fault tolerant, but its memory usage
// is quite high (~80Mib)
type BadgerStore struct {
	path string
	db   *badger.DB
}

func NewBadgerStore(path string) *BadgerStore {
	return &BadgerStore{
		path: path,
		db:   nil,
	}
}

func (s *BadgerStore) Open() {
	opts := badger.DefaultOptions(s.path)
	opts.SyncWrites = true
	opts.Logger = logging.BadgerLogger{}

	db, err := badger.Open(opts)
	if err != nil {
		logging.Panic(err, "failed to open badger DB")
	}
	s.db = db

	log.Info("Badger store opened. Path: ", s.path)
}

func (s *BadgerStore) Close() {
	if err := s.db.Close(); err != nil {
		logging.Panic(err, "failed to close badger DB")
	} else {
		log.Info("Badger store closed")
	}
}

func (s *BadgerStore) Put(key string, message packets.ControlPacket) {
	err := s.db.Update(func(txn *badger.Txn) error {
		var b bytes.Buffer
		if err := message.Write(&b); err != nil {
			return err
		}
		return txn.Set([]byte(key), b.Bytes())
	})

	if err != nil {
		logging.Error(err, "Failed to put message to badger")
	}
}

func (s *BadgerStore) Get(key string) packets.ControlPacket {
	var msg packets.ControlPacket

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if eris.Is(err, badger.ErrKeyNotFound) {
				log.Errorf("Item '%s' not found", key)
			}
			return err
		}
		return item.Value(func(val []byte) error {
			msg, err = packets.ReadPacket(bytes.NewReader(val))
			return err
		})
	})

	if err != nil {
		logging.Error(err, "Failed to get item from badger ", key)
		return nil
	}

	return msg
}

func (s *BadgerStore) All() []string {
	const capacity = 10

	keys := make([]string, 0, capacity)

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = capacity
		opts.PrefetchValues = false

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			keys = append(keys, string(item.Key()))
		}

		return nil
	})

	if err != nil {
		logging.Error(err, "Failed to get keys from badger")
		return nil
	}

	log.Debug("Keys: ", keys)

	return keys
}

func (s *BadgerStore) Del(key string) {
	err := s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})

	if err != nil {
		logging.Error(err, "Failed to delete key from badger")
	}
}

func (s *BadgerStore) Reset() {
	err := s.db.DropAll()
	if err != nil {
		logging.Error(err, "Failed to drop badger database")
	}
	log.Info("Badger database dropped")
}
