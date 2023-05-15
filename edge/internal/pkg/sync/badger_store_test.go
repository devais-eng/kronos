package sync

import (
	"fmt"
	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"path"
	"testing"
)

type BadgerStoreSuite struct {
	suite.Suite
	storePath string
	store     *BadgerStore
}

func (s *BadgerStoreSuite) SetupSuite() {
	logrus.SetOutput(ioutil.Discard)

	s.storePath = path.Join(s.T().TempDir(), "badger-test")
	s.store = NewBadgerStore(s.storePath)
	s.store.Open()
}

func (s *BadgerStoreSuite) TearDownSuite() {
	s.store.Close()
}

func (s *BadgerStoreSuite) SetupTest() {
	s.store.Reset()
}

func (s *BadgerStoreSuite) TestInsert() {
	assert := s.Require()

	key := "key1"
	msg := packets.NewControlPacket(packets.Publish)

	assert.Nil(s.store.Get(key))

	s.store.Del(key)

	s.store.Put(key, msg)

	storedMsg := s.store.Get(key)
	assert.NotNil(storedMsg)

	assert.Equal(msg.String(), storedMsg.String())
}

func (s *BadgerStoreSuite) TestDelete() {
	assert := s.Require()

	containsKey := func(key string) bool {
		keys := s.store.All()
		for _, k := range keys {
			if k == key {
				return true
			}
		}
		return false
	}

	key := "key2"
	msg := packets.NewControlPacket(packets.Connack)

	s.store.Put(key, msg)

	assert.True(containsKey(key))

	// Delete
	s.store.Del(key)

	assert.False(containsKey(key))
}

func (s *BadgerStoreSuite) TestReset() {
	assert := s.Require()

	itemsCount := 20
	msg := packets.NewControlPacket(packets.Publish)

	for i := 0; i < itemsCount; i++ {
		s.store.Put(fmt.Sprintf("key-%d", i), msg)
	}

	keys := s.store.All()

	assert.Equal(itemsCount, len(keys))

	s.store.Reset()

	keys = s.store.All()

	assert.Len(keys, 0)
}

func TestBadgerStore(t *testing.T) {
	suite.Run(t, new(BadgerStoreSuite))
}
