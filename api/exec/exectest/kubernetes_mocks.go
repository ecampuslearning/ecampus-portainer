package exectest

import (
	portainer "github.com/portainer/portainer/api"
)

type kubernetesMockDeployer struct {
	portainer.KubernetesDeployer
}

// NewKubernetesDeployer creates a mock kubernetes deployer
func NewKubernetesDeployer() *kubernetesMockDeployer {
	return &kubernetesMockDeployer{}
}
