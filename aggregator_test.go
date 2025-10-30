package bravka

import (
	"encoding/json"
	"reflect"
	"testing"

	"go.uber.org/zap"
)

//nolint:gocogint // normal for tests
func TestDefaultAggregator_Aggregate(t *testing.T) {
	logger := zap.NewNop()
	agg := &defaultAggregator{log: logger}

	tests := []struct {
		name                string
		mode                string
		allowPartialResults bool
		responses           [][]byte
		mergeResult         map[string]interface{}
		arrayResult         []map[string]interface{}
	}{
		{
			name:                "merge strategy success case",
			mode:                "merge",
			allowPartialResults: false,
			responses: [][]byte{
				[]byte(`{"a": 1, "b": 2}`),
				[]byte(`{"b": 3, "c": 4}`),
			},
			mergeResult: map[string]interface{}{
				"a": float64(1),
				"b": float64(3),
				"c": float64(4),
			},
		},
		{
			name:                "merge strategy success case partial results",
			mode:                "merge",
			allowPartialResults: true,
			responses: [][]byte{
				[]byte(`{"a": 1, "b": 2}`),
				[]byte(`non-json string "c": 3`),
			},
			mergeResult: map[string]interface{}{
				"a": float64(1),
				"b": float64(2),
			},
		},
		{
			name:                "merge strategy fail case partial results",
			mode:                "merge",
			allowPartialResults: false,
			responses: [][]byte{
				[]byte(`{"a": 1, "b": 2}`),
				[]byte(`non-json string "c": 3`),
			},
			mergeResult: nil,
		},
		{
			name:                "array strategy success case",
			mode:                "array",
			allowPartialResults: false,
			responses: [][]byte{
				[]byte(`{"a": 1}`),
				[]byte(`{"b": 2}`),
				[]byte(`{"c": 3}`),
			},
			arrayResult: []map[string]interface{}{
				{"a": float64(1)},
				{"b": float64(2)},
				{"c": float64(3)},
			},
		},
		{
			name:                "unknown strategy",
			mode:                "unknown",
			allowPartialResults: false,
			responses: [][]byte{
				[]byte(`{"x":1}`),
				[]byte(`{"y":2}`),
			},
			mergeResult: nil,
			arrayResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBytes := agg.aggregate(tt.responses, tt.mode, tt.allowPartialResults)

			switch tt.mode {
			case strategyMerge:
				var got map[string]interface{}

				if err := json.Unmarshal(gotBytes, &got); err != nil && tt.allowPartialResults {
					t.Fatalf("failed to unmarshal result: %v", err)
				}

				if !reflect.DeepEqual(got, tt.mergeResult) {
					t.Errorf("got %+v, mergeResult %+v", got, tt.mergeResult)
				}
			case strategyArray:
				var got []map[string]interface{}
				if err := json.Unmarshal(gotBytes, &got); err != nil {
					t.Fatalf("failed to unmarshal array: %v", err)
				}
				if !reflect.DeepEqual(got, tt.arrayResult) {
					t.Errorf("got %+v, mergeResult %+v", got, tt.arrayResult)
				}
			default:
				if gotBytes != nil {
					t.Errorf("got %v, mergeResult nil", gotBytes)
				}
			}
		})
	}
}
