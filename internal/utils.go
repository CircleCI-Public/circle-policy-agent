package internal

import (
	"encoding/json"
	"fmt"
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

func ConvertYAMLMapKeyTypes(x any, path ...string) any {
	switch x := x.(type) {
	case map[interface{}]interface{}:
		result := make(map[string]interface{}, len(x))
		for k, v := range x {
			str, ok := k.(string)
			if !ok {
				str = fmt.Sprintf("%v", k)
			}
			result[str] = ConvertYAMLMapKeyTypes(v, append(path, str)...)
		}
		return result
	case map[string]any:
		for k, v := range x {
			x[k] = ConvertYAMLMapKeyTypes(v, append(path, k)...)
		}
		return x
	case []interface{}:
		for i := range x {
			x[i] = ConvertYAMLMapKeyTypes(x[i], append(path, fmt.Sprintf("%d", i))...)
		}
		return x
	default:
		return x
	}
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Must2[T any](value T, err error) T {
	Must(err)
	return value
}
