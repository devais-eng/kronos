package dbus

import (
	"devais.it/kronos/internal/pkg/config"
	"devais.it/kronos/internal/pkg/db"
	"devais.it/kronos/internal/pkg/db/models"
	"devais.it/kronos/internal/pkg/serialization"
	"devais.it/kronos/internal/pkg/util"
	"github.com/godbus/dbus/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"testing"
)

func newItem() *models.Item {
	id := uuid.NewString()
	return &models.Item{
		ID:   id,
		Name: "TestItem-" + id,
		Type: "TestItem",
	}
}

type DBusTestSuite struct {
	db.SuiteBase
	serializationConf config.SerializationConfig
	dbusConf          config.DBusConfig
	serializer        serialization.Serializer
	deserializer      serialization.Deserializer
	testServer        *Server
	clientConn        *dbus.Conn
}

func (s *DBusTestSuite) SetupSuite() {
	s.SuiteBase.SetupSuite()

	assert := s.Require()

	serializationConf := config.DefaultSerializationConfig()
	serializationConf.Type = serialization.TypeJSON

	conf := config.DefaultDBusConfig()
	conf.Enabled = true
	conf.Serialization = serializationConf

	var err error

	s.dbusConf = conf
	s.serializationConf = serializationConf
	s.serializer, s.deserializer, err = serializationConf.NewSerializer()
	assert.NoError(err)

	s.testServer, err = NewServer(&conf)
	assert.NoError(err)

	err = s.testServer.Start()
	assert.NoError(err)

	if conf.UseSystemBus {
		s.clientConn, err = dbus.ConnectSystemBus()
	} else {
		s.clientConn, err = dbus.ConnectSessionBus()
	}

	assert.NoError(err)

	s.T().Log("DBus server started")
}

func (s *DBusTestSuite) TearDownSuite() {
	s.SuiteBase.TearDownSuite()
	assert := s.Require()

	err := s.clientConn.Close()
	assert.NoError(err)

	err = s.testServer.Stop()
	assert.NoError(err)

	s.T().Log("DBus server stopped")
}

func (s *DBusTestSuite) CallMethod(iface, method string, retvalues interface{}, args ...interface{}) {
	assert := s.Require()

	err := s.clientConn.
		Object(s.dbusConf.InterfaceName, dbus.ObjectPath(s.dbusConf.PathName)).
		Call(iface+"."+method, 0, args...).
		Store(retvalues)

	assert.NoError(err)
}

func (s *DBusTestSuite) GetCount(iface string) int64 {
	var count int64
	s.CallMethod(iface, "Count", &count)
	return count
}

func (s *DBusTestSuite) AssertCount(iface string, expected int) {
	assert := s.Require()
	count := s.GetCount(iface)
	assert.Equal(int64(expected), count)
}

func (s *DBusTestSuite) GetAll(iface string, page, pageSize int, res interface{}) {
	var resMsg messageType
	s.CallMethod(iface, "GetAll", &resMsg, page, pageSize)

	assert := s.Require()
	err := s.deserializer.Deserialize([]byte(resMsg), res)
	assert.NoError(err)
}

func (s *DBusTestSuite) Create(iface, method string, req interface{}, res interface{}) {
	assert := s.Require()

	serBytes, err := s.serializer.Serialize(req)
	assert.NoError(err)

	var resMsg messageType

	s.CallMethod(iface, method, &resMsg, messageType(serBytes))

	if res != nil {
		err = s.deserializer.Deserialize([]byte(resMsg), res)
		assert.NoError(err)
	}
}

func (s *DBusTestSuite) Get(iface, method, id string, res interface{}) {
	assert := s.Require()
	var resMsg messageType
	s.CallMethod(iface, method, &resMsg, id)
	err := s.deserializer.Deserialize([]byte(resMsg), res)
	assert.NoError(err)
}

func (s *DBusTestSuite) Update(iface string, patch map[string]interface{}) {
	assert := s.Require()

	patchBytes, err := s.serializer.Serialize(patch)
	assert.NoError(err)

	var retMsg messageType

	s.CallMethod(iface, "Update", &retMsg, messageType(patchBytes))
}

func (s *DBusTestSuite) Delete(iface, method, id string) {
	var res messageType
	s.CallMethod(iface, method, &res, id)
}

func (s *DBusTestSuite) TestItems() {
	assert := s.Require()

	iface := s.dbusConf.ItemsInterfaceName

	var items []models.Item
	s.GetAll(iface, 1, 10, &items)

	assert.Empty(items)

	s.AssertCount(iface, 0)

	// Create
	item := newItem()

	s.Create(iface, "Create", item, nil)

	s.AssertCount(iface, 1)
	s.GetAll(iface, 1, 10, &items)
	assert.Len(items, 1)

	var createdItem models.Item
	s.Get(iface, "GetByID", item.ID, &createdItem)

	assert.Equal(item.ID, createdItem.ID)
	assert.Equal(item.Name, createdItem.Name)
	assert.Equal(item.Type, createdItem.Type)

	newName := "TestItemNewName"
	s.Update(iface, map[string]interface{}{"id": item.ID, "name": newName})

	s.Get(iface, "GetByID", item.ID, &createdItem)
	assert.Equal(newName, createdItem.Name)

	// Delete created item
	s.Delete(iface, "DeleteByID", item.ID)
	s.AssertCount(iface, 0)

	items = []models.Item{*newItem(), *newItem()}

	s.Create(iface, "CreateBatch", items, nil)
	s.AssertCount(iface, len(items))

	// Delete 1 item
	s.Delete(iface, "HardDeleteByID", items[0].ID)
	s.AssertCount(iface, len(items)-1)
}

func TestDBusServer(t *testing.T) {
	// Skip tests if running inside a Docker container as DBus
	// is not supported
	if util.IsInDocker() {
		return
	}

	suite.Run(t, new(DBusTestSuite))
}
