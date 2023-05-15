package db

import (
	"devais.it/kronos/internal/pkg/constants"
	"devais.it/kronos/internal/pkg/db/models"
)

func CalcTxLength(model interface{}) int {
	var txLen int

	switch v := model.(type) {
	case models.Item:
		txLen = 1 + CalcTxLength(v.Attributes)
	case *models.Item:
		txLen = 1 + CalcTxLength(v.Attributes)
	case models.Attribute:
		txLen = 1
	case *models.Attribute:
		txLen = 1
	case models.Relation:
		txLen = 1
	case *models.Relation:
		txLen = 1
	case []models.Item:
		txLen = len(v)
		for _, item := range v {
			txLen += CalcTxLength(item.Attributes)
		}
	case []models.Attribute:
		txLen = len(v)
	case []models.Relation:
		txLen = len(v)
	case map[string]interface{}:
		txLen = 1

		if attributes, ok := v[constants.AttributesField]; ok {
			if attrsList, ok := attributes.([]map[string]interface{}); ok {
				txLen += CalcTxLength(attrsList)
			} else if attrsList, ok := attributes.([]interface{}); ok {
				txLen += CalcTxLength(attrsList)
			}
		}
	case []map[string]interface{}:
		for i := 0; i < len(v); i++ {
			txLen += CalcTxLength(v[i])
		}
	case []interface{}:
		for i := 0; i < len(v); i++ {
			if jsonMap, ok := v[i].(map[string]interface{}); ok {
				txLen += CalcTxLength(jsonMap)
			}
		}
	default:
		txLen = 0
	}

	return txLen
}
