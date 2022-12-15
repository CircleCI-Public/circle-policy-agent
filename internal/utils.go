package internal

import (
	"encoding/json"
	"fmt"
	"strings"
)

func ToRawInterface(value any) (any, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var result any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// This function has been stolen from the OPA codebase, ast/parser.go:2210 for
// converting yaml decoded types to types that OPA can handle
func ConvertYAMLMapKeyTypes(x interface{}, path ...string) (interface{}, error) {
	var err error
	switch x := x.(type) {
	case map[interface{}]interface{}:
		result := make(map[string]interface{}, len(x))
		for k, v := range x {
			str, ok := k.(string)
			if !ok {
				return nil, fmt.Errorf("invalid map key type(s): %v", strings.Join(path, "/"))
			}
			result[str], err = ConvertYAMLMapKeyTypes(v, append(path, str)...)
			if err != nil {
				return nil, err
			}
		}
		return result, nil
	case []interface{}:
		for i := range x {
			x[i], err = ConvertYAMLMapKeyTypes(x[i], append(path, fmt.Sprintf("%d", i))...)
			if err != nil {
				return nil, err
			}
		}
		return x, nil
	default:
		return x, nil
	}
}

func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}
