package types

import (
	"github.com/portainer/portainer/pkg/libhelm/options"
	"github.com/portainer/portainer/pkg/libhelm/release"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/repo"
)

// HelmPackageManager represents a service that interfaces with Helm
type HelmPackageManager interface {
	Show(showOpts options.ShowOptions) ([]byte, error)
	SearchRepo(searchRepoOpts options.SearchRepoOptions) ([]byte, error)
	List(listOpts options.ListOptions) ([]release.ReleaseElement, error)
	Upgrade(upgradeOpts options.InstallOptions) (*release.Release, error)
	Uninstall(uninstallOpts options.UninstallOptions) error
}

type Repository interface {
	Charts() (repo.ChartVersions, error)
}

type HelmRepo struct {
	Settings *cli.EnvSettings
	Orig     *repo.Entry
}
