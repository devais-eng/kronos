package serialization

import (
	"github.com/fxamacker/cbor/v2"
)

type CborSerializer struct{}

func DefaultCborSerializer() *CborSerializer {
	return &CborSerializer{}
}

func (s *CborSerializer) Serialize(v interface{}) ([]byte, error) {
	return cbor.Marshal(v)
}

func (s *CborSerializer) Deserialize(bytes []byte, v interface{}) error {
	return cbor.Unmarshal(bytes, v)
}
