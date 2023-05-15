package http

import (
	"bytes"
	"devais.it/kronos/internal/pkg/config"
	"devais.it/kronos/internal/pkg/constants"
	"devais.it/kronos/internal/pkg/db"
	"devais.it/kronos/internal/pkg/db/models"
	"devais.it/kronos/internal/pkg/services"
	"encoding/json"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type HTTPSuite struct {
	db.SuiteBase
	handler    http.Handler
	testServer *httptest.Server
	url        string
}

func (s *HTTPSuite) SetupSuite() {
	s.SuiteBase.SetupSuite()

	conf := config.DefaultHTTPConfig()
	conf.Enabled = true
	conf.DebugMode = false
	conf.ReplyCreatedData = true

	s.handler = NewServer(&conf).engine
	s.testServer = httptest.NewServer(s.handler)
	s.url = s.testServer.URL

	s.T().Logf("Test server listening on %s", s.url)
}

func (s *HTTPSuite) TearDownSuite() {
	s.SuiteBase.TearDownSuite()
	s.testServer.Close()
}

func (s *HTTPSuite) GetString(path string) string {
	assert := s.Require()

	resp, err := http.Get(s.url + path)
	assert.NoError(err)
	assert.Equal(http.StatusOK, resp.StatusCode)

	defer func() {
		assert.NoError(resp.Body.Close())
	}()

	data, err := io.ReadAll(resp.Body)
	assert.NoError(err)

	return string(data)
}

func (s *HTTPSuite) GetJSON(path string, v interface{}) {
	assert := s.Require()

	resp, err := http.Get(s.url + path)
	assert.NoError(err)

	defer func() {
		assert.NoError(resp.Body.Close())
	}()

	data, err := io.ReadAll(resp.Body)
	assert.NoError(err)

	err = json.Unmarshal(data, v)
	assert.NoError(err)
}

func (s *HTTPSuite) PostJSON(path string, reqModel, resModel interface{}) {
	assert := s.Require()

	body, err := json.Marshal(reqModel)
	assert.NoError(err)

	resp, err := http.Post(s.url+path, "application/json", bytes.NewBuffer(body))
	assert.NoError(err)
	assert.Contains([]int{http.StatusOK, http.StatusCreated}, resp.StatusCode)

	defer func() {
		assert.NoError(resp.Body.Close())
	}()

	body, err = io.ReadAll(resp.Body)
	assert.NoError(err)

	err = json.Unmarshal(body, resModel)
	assert.NoError(err)
}

func (s *HTTPSuite) PutJSON(path string, reqModel, resModel interface{}) {
	assert := s.Require()

	body, err := json.Marshal(reqModel)
	assert.NoError(err)

	client := &http.Client{}

	req, err := http.NewRequest(http.MethodPut, s.url+path, bytes.NewBuffer(body))
	assert.NoError(err)

	res, err := client.Do(req)
	assert.NoError(err)

	defer func() {
		assert.NoError(res.Body.Close())
	}()

	assert.Equal(http.StatusOK, res.StatusCode)

	body, err = io.ReadAll(res.Body)
	assert.NoError(err)

	err = json.Unmarshal(body, resModel)
	assert.NoError(err)
}

func (s *HTTPSuite) Delete(path string) *http.Response {
	assert := s.Require()

	client := &http.Client{}

	req, err := http.NewRequest(http.MethodDelete, s.url+path, nil)
	assert.NoError(err)

	resp, err := client.Do(req)
	assert.NoError(err)

	return resp
}

func (s *HTTPSuite) TestPing() {
	assert := s.Require()

	resp := s.GetString("/ping")

	assert.Equal("pong", resp)
}

func (s *HTTPSuite) TestHealth() {
	assert := s.Require()

	resp := s.GetString("/health")
	assert.Equal("ok", resp)
}

func (s *HTTPSuite) TestCreateItems() {
	assert := s.Require()

	start := uint64(time.Now().UnixNano() / 1_000_000)
	var end uint64

	checkItem := func(expected models.Item, actual models.Item) {
		assert.Equal(expected.ID, actual.ID)
		assert.Equal(expected.Name, actual.Name)
		assert.Equal(expected.Type, actual.Type)

		assert.Equal(constants.ModifiedByHTTPAPIName, actual.CreatedBy)
		assert.Equal(constants.ModifiedByHTTPAPIName, actual.ModifiedBy)

		assert.GreaterOrEqual(actual.CreatedAt, start)
		assert.GreaterOrEqual(actual.ModifiedAt, start)

		assert.LessOrEqual(actual.CreatedAt, end)
		assert.LessOrEqual(actual.ModifiedAt, end)
	}

	fakeItems := []models.Item{
		{
			ID:   "FakeItem00-ID",
			Name: "FakeItem00",
			Type: "FakeItem",
		},
		{
			ID:   "FakeItem01-ID",
			Name: "FakeItem01",
			Type: "FakeItem",
		},
		{
			ID:   "FakeItem02-ID",
			Name: "FakeItem02",
			Type: "FakeItem",
		},
	}

	var createdItems []models.Item

	s.PostJSON("/items", fakeItems, &createdItems)

	end = uint64(time.Now().UnixNano() / 1_000_000)

	assert.NotNil(createdItems)
	assert.Len(createdItems, len(fakeItems))

	for i := 0; i < len(createdItems); i++ {
		checkItem(fakeItems[i], createdItems[i])

		item, err := services.GetItemByID(fakeItems[i].ID)
		assert.NoError(err)

		checkItem(fakeItems[i], *item)
	}

	var count map[string]int

	s.GetJSON("/items/count", &count)

	assert.NotEmpty(count)
	assert.Contains(count, "count")
	assert.Equal(len(fakeItems), count["count"])
}

func (s *HTTPSuite) TestUpdateItems() {
	assert := s.Require()

	fakeItem := models.Item{
		ID:   "FakeItem00-ID",
		Name: "FakeItem00",
		Type: "FakeItem",
	}

	var createdItems []models.Item

	s.PostJSON("/items", []models.Item{fakeItem}, &createdItems)

	assert.Len(createdItems, 1)

	newName := "FakeItem00NewName"
	patch := map[string]interface{}{"name": newName}
	updatedItem := &models.Item{}

	s.PutJSON("/item/"+fakeItem.ID, patch, updatedItem)

	assert.Equal(newName, updatedItem.Name)

	updatedItem, err := services.GetItemByID(fakeItem.ID)
	assert.NoError(err)
	assert.Equal(newName, updatedItem.Name)

	s.GetJSON("/item/"+fakeItem.ID, updatedItem)
	assert.Equal(newName, updatedItem.Name)
}

func (s *HTTPSuite) TestDeleteItems() {
	assert := s.Require()

	fakeItem := models.Item{
		ID:   "FakeItem00-ID",
		Name: "FakeItem00",
		Type: "FakeItem",
	}

	resp := s.Delete("/item/" + fakeItem.ID)
	assert.Equal(http.StatusNotFound, resp.StatusCode)
	assert.NoError(resp.Body.Close())

	var createdItems []models.Item

	s.PostJSON("/items", []models.Item{fakeItem}, &createdItems)

	assert.Len(createdItems, 1)

	resp = s.Delete("/item/" + fakeItem.ID)
	assert.Equal(http.StatusOK, resp.StatusCode)

	resp, err := http.Get(s.url + "/item/" + fakeItem.ID)
	assert.NoError(err)
	assert.NoError(resp.Body.Close())

	assert.Equal(http.StatusNotFound, resp.StatusCode)
}

func TestHTTPServer(t *testing.T) {
	suite.Run(t, new(HTTPSuite))
}
