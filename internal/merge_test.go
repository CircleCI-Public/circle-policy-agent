package internal

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMerge(t *testing.T) {
	cases := []struct {
		Name   string
		A      string
		B      string
		Result string
	}{
		{
			Name:   "simple map",
			A:      `{"hello": "world"}`,
			B:      `{"jane":"doe"}`,
			Result: `{"hello":"world","jane":"doe"}`,
		},
		{
			Name:   "map with collisions",
			A:      `{"key":{"hello":"world","sub":{}}}`,
			B:      `{"key":{"jane":"doe","sub":{"k":"v"}}}`,
			Result: `{"key":{"hello":"world","jane":"doe","sub":{"k":"v"}}}`,
		},
		{
			Name:   "nil removes keys from map",
			A:      `{"hello":"world"}`,
			B:      `{"hello":null}`,
			Result: "{}",
		},
		{
			Name:   "slice",
			A:      "[1,2,3,4]",
			B:      "[1,2,1]",
			Result: "[1,2,1,4]",
		},
		{
			Name:   "null removes element from slice",
			A:      "[1,2,3]",
			B:      "[null, null]",
			Result: "[3]",
		},
		{
			Name:   "slice of maps",
			A:      `[{"hello":"world"},5]`,
			B:      `[{"hello":"john"}]`,
			Result: `[{"hello":"john"},5]`,
		},
		{
			Name:   "B overrides A",
			A:      "3",
			B:      "true",
			Result: "true",
		},
		{
			Name:   "can squash any value with null",
			A:      "5",
			B:      "null",
			Result: "null",
		},
		{
			Name:   "null gets overwritten",
			A:      "null",
			B:      "true",
			Result: "true",
		},
		{
			Name: "can overwrite map with a slice",
			A: `{
					"key": {"inner":"map"},
					"value": 42
				}`,
			B: `{"key":[{}]}`,
			Result: `{ 
						"key": [{}],
						"value": 42
					 }`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var a, b any
			require.NoError(t, json.Unmarshal([]byte(tc.A), &a))
			require.NoError(t, json.Unmarshal([]byte(tc.B), &b))

			merged := Merge(a, b)

			actual, err := json.Marshal(merged)
			require.NoError(t, err)

			require.JSONEq(t, tc.Result, string(actual))
		})
	}
}
