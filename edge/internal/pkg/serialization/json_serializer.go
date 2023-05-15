package serialization

import (
	jsoniter "github.com/json-iterator/go"
)

type JSONSerializer struct {
	Prefix string
	Indent string
	json   jsoniter.API
}

func DefaultJSONSerializer() *JSONSerializer {
	return &JSONSerializer{
		Prefix: "",
		Indent: "",
		json:   jsoniter.ConfigCompatibleWithStandardLibrary,
	}
}

func (s *JSONSerializer) Serialize(v interface{}) ([]byte, error) {
	return s.json.MarshalIndent(v, s.Prefix, s.Indent)
}

func (s *JSONSerializer) Deserialize(bytes []byte, v interface{}) error {
	return s.json.Unmarshal(bytes, v)
}
