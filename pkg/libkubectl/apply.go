package libkubectl

import (
	"bytes"
	"context"
	"fmt"

	"k8s.io/kubectl/pkg/cmd/apply"
)

func (c *Client) Apply(ctx context.Context, manifests []string) (string, error) {
	buf := new(bytes.Buffer)

	cmd := apply.NewCmdApply("kubectl", c.factory, c.streams)
	cmd.SetArgs(manifestFilesToArgs(manifests))
	cmd.SetOut(buf)

	if err := cmd.ExecuteContext(ctx); err != nil {
		return "", fmt.Errorf("error applying resources: %w", err)
	}

	return buf.String(), nil
}
