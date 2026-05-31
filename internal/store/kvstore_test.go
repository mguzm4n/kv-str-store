package store_test

import (
	"testing"

	"github.com/mguzm4n/kv-str-store/internal/store"
)

type Payload struct {
	key   string
	value string
}

type SyncTest0 struct {
	name        string
	payloads    []Payload
	expected    map[string]string
	expectError bool
}

func TestKVStore(t *testing.T) {
	tests := []SyncTest0{
		{
			name: "sets exactly one key and one value exactly",
			payloads: []Payload{
				{key: "id:10", value: "{\"userId\": 10, \"name\":\"John\"}"},
			},
			expected: map[string]string{
				"id:10": "{\"userId\": 10, \"name\":\"John\"}",
			},
		},
		{
			name: "sets exactly one key but overwrites it multiple times",
			payloads: []Payload{
				{key: "id:10", value: "{\"userId\": 10, \"name\":\"John\"}"},
				{key: "id:10", value: "{\"userId\": 10, \"name\":\"John Doe\"}"},
				{key: "id:10", value: "{\"userId\": 10, \"name\":\"Johnny Doe\"}"},
			},
			expected: map[string]string{
				"id:10": "{\"userId\": 10, \"name\":\"Johnny Doe\"}",
			},
		},
		{
			name: "handles multiple distinct keys correctly",
			payloads: []Payload{
				{key: "A", value: "valueA"},
				{key: "B", value: "valueB"},
				{key: "C", value: "valueC"},
			},
			expected: map[string]string{
				"A": "valueA",
				"B": "valueB",
				"C": "valueC",
			},
		},
		{
			name: "getting a non-existent key returns error",
			payloads: []Payload{
				{key: "A", value: "valueA"},
			},
			expected:    map[string]string{"Z": ""},
			expectError: true,
		},
		{
			name: "empty values are stored and retrieved correctly",
			payloads: []Payload{
				{key: "empty_val_key", value: ""},
			},
			expected: map[string]string{"empty_val_key": ""},
		},
	}

	// sequential reads
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			store, _ := store.NewStore(dir)
			defer store.CloseSegments()
			for _, payload := range test.payloads {
				err := store.PutKey(payload.key, payload.value)
				if err != nil {
					t.Fatalf("Failed to put key %s: %v", payload.key, err)
				}
			}

			for getParam, expectedVal := range test.expected {
				result, err := store.GetKey(getParam)

				if test.expectError {
					if err == nil {
						t.Errorf("Expected an error for key %s, but got none", getParam)
					}
					continue
				}

				if err != nil {
					t.Errorf("Unexpected error getting key %s: %v", getParam, err)
				}

				if result != expectedVal {
					t.Errorf("Result %s v/s Expected %s", result, expectedVal)
				}
			}
		})
	}

}
