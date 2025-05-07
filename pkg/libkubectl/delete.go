package libkubectl

import (
	"bytes"
	"context"
	"fmt"

	"k8s.io/kubectl/pkg/cmd/delete"
)

func (c *Client) Delete(ctx context.Context, manifests []string) (string, error) {
	buf := new(bytes.Buffer)

	cmd := delete.NewCmdDelete(c.factory, c.streams)
	cmd.SetArgs(manifestFilesToArgs(manifests))
	cmd.Flags().Set("ignore-not-found", "true")
	cmd.SetOut(buf)

	if err := cmd.ExecuteContext(ctx); err != nil {
		return "", fmt.Errorf("error deleting resources: %w", err)
	}

	return buf.String(), nil
}
