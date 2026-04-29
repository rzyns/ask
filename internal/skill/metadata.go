package skill

import (
	"encoding/json"
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"
)

// MetadataMap is the SKILL.md frontmatter metadata object.
//
// It intentionally models JSON-like data instead of arbitrary YAML values:
// strings, booleans, numbers, null, arrays, and string-keyed objects. YAML-only
// constructs such as aliases, timestamps, binary blobs, and non-string map keys
// are rejected during parsing so callers do not need to handle arbitrary Go
// values.
type MetadataMap map[string]MetadataValue

// MetadataKind identifies the concrete JSON-like value held by MetadataValue.
type MetadataKind string

const (
	MetadataNull   MetadataKind = "null"
	MetadataString MetadataKind = "string"
	MetadataBool   MetadataKind = "bool"
	MetadataNumber MetadataKind = "number"
	MetadataArray  MetadataKind = "array"
	MetadataObject MetadataKind = "object"
)

// MetadataValue is a single JSON-like metadata value.
type MetadataValue struct {
	kind   MetadataKind
	str    string
	boolv  bool
	number string
	array  []MetadataValue
	object MetadataMap
}

// Kind returns the concrete metadata value kind.
func (v MetadataValue) Kind() MetadataKind { return v.kind }

// String returns the value as a string when Kind() == MetadataString.
func (v MetadataValue) String() (string, bool) {
	if v.kind != MetadataString {
		return "", false
	}
	return v.str, true
}

// Bool returns the value as a bool when Kind() == MetadataBool.
func (v MetadataValue) Bool() (bool, bool) {
	if v.kind != MetadataBool {
		return false, false
	}
	return v.boolv, true
}

// Number returns the original YAML numeric literal when Kind() == MetadataNumber.
func (v MetadataValue) Number() (string, bool) {
	if v.kind != MetadataNumber {
		return "", false
	}
	return v.number, true
}

// Array returns the array elements when Kind() == MetadataArray.
func (v MetadataValue) Array() ([]MetadataValue, bool) {
	if v.kind != MetadataArray {
		return nil, false
	}
	return v.array, true
}

// Object returns the object members when Kind() == MetadataObject.
func (v MetadataValue) Object() (MetadataMap, bool) {
	if v.kind != MetadataObject {
		return nil, false
	}
	return v.object, true
}

// IsNull reports whether the value is YAML/JSON null.
func (v MetadataValue) IsNull() bool { return v.kind == MetadataNull }

// String returns a string metadata member by key.
func (m MetadataMap) String(key string) (string, bool) {
	if m == nil {
		return "", false
	}
	return m[key].String()
}

// UnmarshalYAML decodes a metadata object from a YAML node, rejecting YAML-only
// constructs and non-string object keys.
func (m *MetadataMap) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode && node.ShortTag() == "!!null" {
		*m = nil
		return nil
	}
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("metadata must be an object")
	}
	object, err := metadataMapFromYAMLNode(node)
	if err != nil {
		return err
	}
	*m = object
	return nil
}

// MarshalJSON preserves JSON-like metadata for API/JSON consumers.
func (v MetadataValue) MarshalJSON() ([]byte, error) {
	switch v.kind {
	case MetadataNull, "":
		return []byte("null"), nil
	case MetadataString:
		return json.Marshal(v.str)
	case MetadataBool:
		return json.Marshal(v.boolv)
	case MetadataNumber:
		return []byte(v.number), nil
	case MetadataArray:
		return json.Marshal(v.array)
	case MetadataObject:
		return json.Marshal(v.object)
	default:
		return nil, fmt.Errorf("unknown metadata kind %q", v.kind)
	}
}

func metadataValueFromYAMLNode(node *yaml.Node) (MetadataValue, error) {
	if node.Kind == yaml.AliasNode {
		return MetadataValue{}, fmt.Errorf("metadata aliases are not supported")
	}

	switch node.Kind {
	case yaml.ScalarNode:
		switch node.ShortTag() {
		case "!!null":
			return MetadataValue{kind: MetadataNull}, nil
		case "!!str":
			return MetadataValue{kind: MetadataString, str: node.Value}, nil
		case "!!bool":
			value, err := strconv.ParseBool(node.Value)
			if err != nil {
				return MetadataValue{}, fmt.Errorf("invalid metadata bool %q", node.Value)
			}
			return MetadataValue{kind: MetadataBool, boolv: value}, nil
		case "!!int", "!!float":
			if !json.Valid([]byte(node.Value)) {
				return MetadataValue{}, fmt.Errorf("metadata number %q is not JSON-compatible", node.Value)
			}
			return MetadataValue{kind: MetadataNumber, number: node.Value}, nil
		default:
			return MetadataValue{}, fmt.Errorf("unsupported metadata scalar type %s", node.ShortTag())
		}
	case yaml.SequenceNode:
		items := make([]MetadataValue, 0, len(node.Content))
		for i, child := range node.Content {
			value, err := metadataValueFromYAMLNode(child)
			if err != nil {
				return MetadataValue{}, fmt.Errorf("metadata array item %d: %w", i, err)
			}
			items = append(items, value)
		}
		return MetadataValue{kind: MetadataArray, array: items}, nil
	case yaml.MappingNode:
		object, err := metadataMapFromYAMLNode(node)
		if err != nil {
			return MetadataValue{}, err
		}
		return MetadataValue{kind: MetadataObject, object: object}, nil
	default:
		return MetadataValue{}, fmt.Errorf("unsupported metadata YAML node kind %d", node.Kind)
	}
}

func metadataMapFromYAMLNode(node *yaml.Node) (MetadataMap, error) {
	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("metadata object must be a map")
	}
	object := make(MetadataMap, len(node.Content)/2)
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		if keyNode.Kind != yaml.ScalarNode || keyNode.ShortTag() != "!!str" {
			return nil, fmt.Errorf("metadata object keys must be strings")
		}
		value, err := metadataValueFromYAMLNode(valueNode)
		if err != nil {
			return nil, fmt.Errorf("metadata.%s: %w", keyNode.Value, err)
		}
		object[keyNode.Value] = value
	}
	return object, nil
}
