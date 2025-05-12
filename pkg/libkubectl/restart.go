package libkubectl

import (
	"bytes"
	"context"
	"fmt"

	"k8s.io/kubectl/pkg/cmd/rollout"
)

func (c *Client) RolloutRestart(ctx context.Context, manifests []string) (string, error) {
	buf := new(bytes.Buffer)

	cmd := rollout.NewCmdRollout(c.factory, c.streams)
	args := []string{"restart"}
	args = append(args, resourcesToArgs(manifests)...)

	cmd.SetArgs(args)
	cmd.SetOut(buf)

	if err := cmd.ExecuteContext(ctx); err != nil {
		return "", fmt.Errorf("error restarting resources: %w", err)
	}

	return buf.String(), nil
}
