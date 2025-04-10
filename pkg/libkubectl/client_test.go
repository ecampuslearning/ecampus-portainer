package libkubectl

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name             string
		libKubectlAccess ClientAccess
		namespace        string
		kubeconfig       string
		insecure         bool
		wantErr          bool
		errContains      string
	}{
		{
			name:             "valid client with token and server",
			libKubectlAccess: ClientAccess{Token: "test-token", ServerUrl: "https://localhost:6443"},
			namespace:        "default",
			insecure:         true,
			wantErr:          false,
		},
		{
			name:       "valid client with kubeconfig",
			kubeconfig: "/path/to/kubeconfig",
			namespace:  "test-namespace",
			insecure:   false,
			wantErr:    false,
		},
		{
			name:        "missing both token/server and kubeconfig",
			namespace:   "default",
			insecure:    false,
			wantErr:     true,
			errContains: "must provide either a kubeconfig path or a server and token",
		},
		{
			name:             "missing token with server",
			libKubectlAccess: ClientAccess{ServerUrl: "https://localhost:6443"},
			namespace:        "default",
			insecure:         false,
			wantErr:          true,
			errContains:      "must provide either a kubeconfig path or a server and token",
		},
		{
			name:             "missing server with token",
			libKubectlAccess: ClientAccess{Token: "test-token"},
			namespace:        "default",
			insecure:         false,
			wantErr:          true,
			errContains:      "must provide either a kubeconfig path or a server and token",
		},
		{
			name:             "empty namespace is valid",
			libKubectlAccess: ClientAccess{Token: "test-token", ServerUrl: "https://localhost:6443"},
			namespace:        "",
			insecure:         false,
			wantErr:          false,
		},
		{
			name:             "insecure true with valid credentials",
			libKubectlAccess: ClientAccess{Token: "test-token", ServerUrl: "https://localhost:6443"},
			namespace:        "default",
			insecure:         true,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(&tt.libKubectlAccess, tt.namespace, tt.kubeconfig, tt.insecure)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				if got := err.Error(); got != tt.errContains {
					t.Errorf("NewClient() error = %v, want error containing %v", got, tt.errContains)
				}
				return
			}

			if !tt.wantErr {
				if client == nil {
					t.Error("NewClient() returned nil client when no error was expected")
					return
				}

				// Verify client fields are properly initialized
				if client.factory == nil {
					t.Error("NewClient() client.factory is nil")
				}
				if client.out == nil {
					t.Error("NewClient() client.out is nil")
				}
			}
		})
	}
}
