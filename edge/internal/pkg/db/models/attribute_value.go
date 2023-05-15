package models

// AttributeValue is a structure representing the value of an attribute.
// It is returned by getValue queries
type AttributeValue struct {
	Value     string `json:"value"`
	ValueType string `json:"value_type"`
}
