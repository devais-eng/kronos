package serialization

import (
	"fmt"
	"github.com/rotisserie/eris"
	"strings"
)

type Type int

const (
	TypeJSON Type = iota
	TypeCBOR
)

func (s Type) MarshalText() ([]byte, error) {
	switch s {
	case TypeJSON:
		return []byte("JSON"), nil
	case TypeCBOR:
		return []byte("CBOR"), nil
	default:
		return nil, eris.New("Invalid serialization type")
	}
}

func (s Type) String() string {
	text, err := s.MarshalText()
	if err != nil {
		return "Unknown"
	}
	return string(text)
}

func TypeFromString(str string) (Type, error) {
	switch strings.ToLower(str) {
	case "json":
		return TypeJSON, nil
	case "cbor":
		return TypeCBOR, nil
	default:
		return -1, eris.Errorf("Unknown serialization type: '%s'", str)
	}
}

type Serializer interface {
	Serialize(v interface{}) ([]byte, error)
}

type Deserializer interface {
	Deserialize(bytes []byte, v interface{}) error
}

func DefaultSerializer(serializationType Type) (Serializer, Deserializer) {
	switch serializationType {
	case TypeJSON:
		s := DefaultJSONSerializer()
		return s, s
	case TypeCBOR:
		s := DefaultCborSerializer()
		return s, s
	}

	panic(fmt.Sprintf("Invalid serialization type '%v'", serializationType))
}
