package compose

import (
	"github.com/docker/cli/cli/command"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
)

type ComposeDeployer struct {
	createComposeServiceFn func(command.Cli) api.Service
}

// NewComposeDeployer creates a new compose deployer
func NewComposeDeployer() *ComposeDeployer {
	return &ComposeDeployer{
		createComposeServiceFn: compose.NewComposeService,
	}
}
