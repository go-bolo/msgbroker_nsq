package msgbroker_nsq

import (
	"fmt"
	"testing"

	"github.com/go-bolo/msgbroker"
)

// Mock implementation of MessageHandler for testing
type mockMessageHandler struct {
	HandledMessages []struct {
		QueueName string
		Message   msgbroker.Message
	}
	ShouldError bool
}

func (m *mockMessageHandler) HandleMessage(queueName string, message msgbroker.Message) error {
	m.HandledMessages = append(m.HandledMessages, struct {
		QueueName string
		Message   msgbroker.Message
	}{
		QueueName: queueName,
		Message:   message,
	})

	if m.ShouldError {
		return fmt.Errorf("mock error")
	}

	return nil
}

func TestNSQQueue_GetName(t *testing.T) {
	tests := []struct {
		name         string
		queueName    string
		expectedName string
	}{
		{
			name:         "should return queue name when set",
			queueName:    "test-queue",
			expectedName: "test-queue",
		},
		{
			name:         "should return empty string when name is empty",
			queueName:    "",
			expectedName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := &NSQQueue{
				Name: tt.queueName,
			}

			result := queue.GetName()

			if result != tt.expectedName {
				t.Errorf("Expected name %s, got %s", tt.expectedName, result)
			}
		})
	}
}

func TestNSQQueue_SetName(t *testing.T) {
	tests := []struct {
		name      string
		queueName string
	}{
		{
			name:      "should set queue name successfully",
			queueName: "new-queue",
		},
		{
			name:      "should set empty queue name successfully",
			queueName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := &NSQQueue{}

			err := queue.SetName(tt.queueName)

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if queue.Name != tt.queueName {
				t.Errorf("Expected name %s, got %s", tt.queueName, queue.Name)
			}
		})
	}
}

func TestNSQQueue_GetHandler(t *testing.T) {
	mockHandler := &mockMessageHandler{}

	tests := []struct {
		name            string
		handler         msgbroker.MessageHandler
		expectedHandler msgbroker.MessageHandler
	}{
		{
			name:            "should return handler when set",
			handler:         mockHandler,
			expectedHandler: mockHandler,
		},
		{
			name:            "should return nil when handler is nil",
			handler:         nil,
			expectedHandler: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := &NSQQueue{
				Handler: tt.handler,
			}

			result := queue.GetHandler()

			if result != tt.expectedHandler {
				t.Errorf("Expected handler %v, got %v", tt.expectedHandler, result)
			}
		})
	}
}

func TestNSQQueue_SetHandler(t *testing.T) {
	mockHandler := &mockMessageHandler{}

	tests := []struct {
		name    string
		handler msgbroker.MessageHandler
	}{
		{
			name:    "should set handler successfully",
			handler: mockHandler,
		},
		{
			name:    "should set nil handler successfully",
			handler: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := &NSQQueue{}

			err := queue.SetHandler(tt.handler)

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if queue.Handler != tt.handler {
				t.Errorf("Expected handler %v, got %v", tt.handler, queue.Handler)
			}
		})
	}
}

func TestNSQQueue_Creation(t *testing.T) {
	mockHandler := &mockMessageHandler{}
	queueName := "test-queue"

	queue := &NSQQueue{
		Name:    queueName,
		Handler: mockHandler,
	}

	if queue.Name != queueName {
		t.Errorf("Expected name %s, got %s", queueName, queue.Name)
	}

	if queue.Handler != mockHandler {
		t.Errorf("Expected handler %v, got %v", mockHandler, queue.Handler)
	}
}

func TestNSQQueue_EmptyCreation(t *testing.T) {
	queue := &NSQQueue{}

	if queue.Name != "" {
		t.Errorf("Expected empty name, got %s", queue.Name)
	}

	if queue.Handler != nil {
		t.Errorf("Expected nil handler, got %v", queue.Handler)
	}
}

func TestNSQQueue_InterfaceCompliance(t *testing.T) {
	// This test ensures that NSQQueue implements the msgbroker.Queue interface
	var _ msgbroker.Queue = &NSQQueue{}
}

func TestNSQQueueHandler_EmptyStruct(t *testing.T) {
	// Test that NSQQueueHandler can be instantiated
	handler := &NSQQueueHandler{}

	if handler == nil {
		t.Error("Expected NSQQueueHandler to be instantiated, got nil")
	}
}
