package sdk

import (
	"time"

	"github.com/portainer/portainer/pkg/libhelm/types"
	"helm.sh/helm/v3/pkg/cli"
)

// HelmSDKPackageManager is a wrapper for the helm SDK which implements HelmPackageManager
type HelmSDKPackageManager struct {
	settings *cli.EnvSettings
	timeout  time.Duration
}

// NewHelmSDKPackageManager initializes a new HelmPackageManager service using the Helm SDK
func NewHelmSDKPackageManager() types.HelmPackageManager {
	settings := cli.New()
	return &HelmSDKPackageManager{
		settings: settings,
		timeout:  300 * time.Second, // 5 minutes default timeout
	}
}
