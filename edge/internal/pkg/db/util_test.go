package db

import (
	"devais.it/kronos/internal/pkg/db/models"
	"encoding/json"
	"github.com/stretchr/testify/suite"
	"testing"
)

type UtilSuite struct {
	suite.Suite
}

func (s *UtilSuite) TestCalcTxLength() {
	assert := s.Require()

	items := []models.Item{
		{
			ID: "TestItem00",
			Attributes: []models.Attribute{
				{
					ID: "TestAttribute1",
				},
				{
					ID: "TestAttribute2",
				},
			},
		},
	}

	txLen := CalcTxLength(items)
	assert.Equal(3, txLen)

	txLen = CalcTxLength(models.Item{})
	assert.Equal(1, txLen)

	txLen = CalcTxLength(models.Attribute{})
	assert.Equal(1, txLen)

	attributes := []models.Attribute{
		{
		},
		{
		},
	}

	txLen = CalcTxLength(attributes)
	assert.Equal(2, txLen)

	itemsBytes, err := json.Marshal(items)
	assert.NoError(err)
	var itemsJSON []map[string]interface{}
	err = json.Unmarshal(itemsBytes, &itemsJSON)
	assert.NoError(err)

	txLen = CalcTxLength(itemsJSON)
	assert.Equal(3, txLen)
}

func TestUtil(t *testing.T) {
	suite.Run(t, new(UtilSuite))
}
