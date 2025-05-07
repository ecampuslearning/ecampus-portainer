package libkubectl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManifestFilesToArgsHelper(t *testing.T) {
	tests := []struct {
		name          string
		manifestFiles []string
		expectedArgs  []string
	}{
		{
			name:          "empty list",
			manifestFiles: []string{},
			expectedArgs:  []string{},
		},
		{
			name:          "single manifest",
			manifestFiles: []string{"manifest.yaml"},
			expectedArgs:  []string{"-f", "manifest.yaml"},
		},
		{
			name:          "multiple manifests",
			manifestFiles: []string{"manifest1.yaml", "manifest2.yaml"},
			expectedArgs:  []string{"-f", "manifest1.yaml", "-f", "manifest2.yaml"},
		},
		{
			name:          "manifests with whitespace",
			manifestFiles: []string{" manifest1.yaml ", "  manifest2.yaml"},
			expectedArgs:  []string{"-f", "manifest1.yaml", "-f", "manifest2.yaml"},
		},
		{
			name:          "kubernetes resource definitions",
			manifestFiles: []string{"deployment/nginx", "service/web"},
			expectedArgs:  []string{"-f", "deployment/nginx", "-f", "service/web"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := manifestFilesToArgs(tt.manifestFiles)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}
