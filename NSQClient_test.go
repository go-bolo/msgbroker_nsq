package msgbroker_nsq

import (
	"testing"

	"github.com/go-bolo/msgbroker"
)

func TestNewNSQClient(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *NSQClientCfg
		expectedConcurrency int
	}{
		{
			name: "should set default concurrency to 0 when not provided",
			cfg: &NSQClientCfg{
				AutoCreateTopic: true,
			},
			expectedConcurrency: 0,
		},
		{
			name: "should set concurrency when provided",
			cfg: &NSQClientCfg{
				AutoCreateTopic: true,
				Concurrency:     5,
			},
			expectedConcurrency: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientInterface := NewNSQClient(tt.cfg)
			client, ok := clientInterface.(*NSQClient)
			
			if !ok {
				t.Fatalf("Expected client to be of type *NSQClient")
			}

			if client.Concurrency != tt.expectedConcurrency {
				t.Errorf("Expected concurrency %d, got %d", tt.expectedConcurrency, client.Concurrency)
			}
		})
	}
}

func TestNSQClient_InterfaceCompliance(t *testing.T) {
	// This test ensures that NSQClient implements the msgbroker.Client interface
	var _ msgbroker.Client = &NSQClient{}
}
