package main

import "testing"

type Payload struct {
	key   string
	value string
}

type SyncTest0 struct {
	name     string
	payloads []Payload
	getParam string
	expected string
}

func TestKVStore(testing *testing.T) {
	tests := []SyncTest0{
		{
			name: "sets exactly one key and one value exactly",
			payloads: []Payload{
				{key: "id:10", value: "{\"userId\": 10, \"name\":\"John\"}"},
			},
			getParam: "id:10",
			expected: "{\"userId\": 10, \"name\":\"John\"}",
		},
		{
			name: "sets exactly one key but overwrites it multiple times",
			payloads: []Payload{
				{key: "id:10", value: "{\"userId\": 10, \"name\":\"John\"}"},
				{key: "id:10", value: "{\"userId\": 10, \"name\":\"John Doe\"}"},
				{key: "id:10", value: "{\"userId\": 10, \"name\":\"Johnny Doe\"}"},
			},
			getParam: "id:10",
			expected: "{\"userId\": 10, \"name\":\"Johnny Doe\"}",
		},
	}

	store, err := NewStore()
	if err != nil {
		testing.Log("Couldn't setup store")
		return
	}
	for i, test := range tests {
		testing.Logf("Test %d - name '%s'", i, test.name)
		for _, payload := range test.payloads {
			store.PutKey(payload.key, payload.value)
		}
		result, err := store.GetKey(test.getParam)
		if err != nil {
			testing.Logf("Test failed with error: %s", err)
			testing.FailNow()
		}
		if test.expected != result {
			testing.Logf("Test failed. Result %s v/s Expected %s", result, test.expected)
			testing.Fail()
		}
	}
}
