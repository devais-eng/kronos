package util

import (
	jsoniter "github.com/json-iterator/go"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

func StructToJSONMap(v interface{}) (map[string]interface{}, error) {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	resMap := map[string]interface{}{}
	err = json.Unmarshal(jsonBytes, &resMap)
	if err != nil {
		return nil, err
	}
	return resMap, nil
}

func JSONToStruct(j interface{}, v interface{}) error {
	jsonBytes, err := json.Marshal(j)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonBytes, v)
}
