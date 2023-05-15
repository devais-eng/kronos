package config

import (
	"devais.it/kronos/internal/pkg/serialization"
	"github.com/rotisserie/eris"
)

type SerializationConfig struct {
	// Type is the type of serialization to use.
	Type serialization.Type

	// JSONPrefix is the prefix string to use for JSON serialization
	JSONPrefix string

	// JSONIdent is the ident string to use for JSON serialization
	JSONIdent string
}

func DefaultSerializationConfig() SerializationConfig {
	return SerializationConfig{
		Type:       serialization.TypeJSON,
		JSONPrefix: "",
		JSONIdent:  "",
	}
}

// NewSerializer constructs a new serializer/deserializer couple using information
// from a SerializationConfig structure
func (c *SerializationConfig) NewSerializer() (serialization.Serializer, serialization.Deserializer, error) {
	if c.Type == serialization.TypeJSON {
		jsonSerializer := serialization.DefaultJSONSerializer()
		jsonSerializer.Prefix = c.JSONPrefix
		jsonSerializer.Indent = c.JSONIdent
		return jsonSerializer, jsonSerializer, nil
	} else if c.Type == serialization.TypeCBOR {
		cborSerializer := serialization.DefaultCborSerializer()
		return cborSerializer, cborSerializer, nil
	}

	return nil, nil, eris.Errorf("Invalid serialization type: '%v'", c.Type)
}
