package msgbroker_nsq

import (
	"testing"
)

func TestNSQMessage_GetData(t *testing.T) {
	testMessage := []byte("test message")
	emptyMessage := []byte{}

	tests := []struct {
		name     string
		data     *[]byte
		expected *[]byte
	}{
		{
			name:     "should return data when data is set",
			data:     &testMessage,
			expected: &testMessage,
		},
		{
			name:     "should return nil when data is nil",
			data:     nil,
			expected: nil,
		},
		{
			name:     "should return empty slice when data is empty",
			data:     &emptyMessage,
			expected: &emptyMessage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &NSQMessage{
				Data: tt.data,
			}

			result := msg.GetData()

			if tt.expected == nil && result != nil {
				t.Errorf("Expected nil, got %v", result)
				return
			}

			if tt.expected != nil && result == nil {
				t.Errorf("Expected %v, got nil", tt.expected)
				return
			}

			if tt.expected != nil && result != nil {
				if len(*tt.expected) != len(*result) {
					t.Errorf("Expected data length %d, got %d", len(*tt.expected), len(*result))
					return
				}

				for i, b := range *tt.expected {
					if (*result)[i] != b {
						t.Errorf("Expected data %v, got %v", *tt.expected, *result)
						break
					}
				}
			}
		})
	}
}

func TestNSQMessage_Creation(t *testing.T) {
	testData := []byte("hello world")

	msg := &NSQMessage{
		Data: &testData,
	}

	if msg.Data == nil {
		t.Error("Expected Data to be set, got nil")
	}

	if string(*msg.Data) != "hello world" {
		t.Errorf("Expected 'hello world', got %s", string(*msg.Data))
	}
}

func TestNSQMessage_EmptyCreation(t *testing.T) {
	msg := &NSQMessage{}

	if msg.Data != nil {
		t.Error("Expected Data to be nil for empty creation, got non-nil")
	}

	result := msg.GetData()
	if result != nil {
		t.Error("Expected GetData() to return nil for empty message, got non-nil")
	}
}
