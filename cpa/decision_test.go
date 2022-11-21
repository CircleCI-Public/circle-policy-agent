package cpa

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJsonOutput(t *testing.T) {
	t.Run("omitempty", func(t *testing.T) {
		d := Decision{
			Status:       "",
			Cause:        "",
			EnabledRules: nil,
			HardFailures: nil,
			SoftFailures: nil,
		}

		bytes, err := json.Marshal(d)
		if err != nil {
			t.Fatalf("failed to marshal decision: %v", err)
		}

		require.Equal(t, `{"status":""}`, string(bytes))
	})

	t.Run("key names", func(t *testing.T) {
		d := Decision{
			Status:       "",
			Cause:        "cause",
			EnabledRules: []string{""},
			HardFailures: []Violation{{}},
			SoftFailures: []Violation{{}},
		}

		bytes, err := json.Marshal(d)
		if err != nil {
			t.Fatalf("failed to marshal json: %v", err)
		}

		var m map[string]interface{}
		if err := json.Unmarshal(bytes, &m); err != nil {
			t.Fatalf("failed to unmarshal json: %v", err)
		}

		keys := make([]string, 0, len(m))
		for key := range m {
			keys = append(keys, key)
		}
		sort.StringSlice(keys).Sort()

		require.EqualValues(
			t,
			[]string{"cause", "enabled_rules", "hard_failures", "soft_failures", "status"},
			keys,
		)
	})
}
