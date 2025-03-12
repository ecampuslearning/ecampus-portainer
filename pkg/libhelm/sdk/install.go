package sdk

import (
	"os"

	"github.com/pkg/errors"
	"github.com/portainer/portainer/pkg/libhelm/options"
	"github.com/portainer/portainer/pkg/libhelm/release"
	"github.com/rs/zerolog/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/postrender"
)

// Install implements the HelmPackageManager interface by using the Helm SDK to install a chart.
func (hspm *HelmSDKPackageManager) Install(installOpts options.InstallOptions) (*release.Release, error) {
	log.Debug().
		Str("context", "HelmClient").
		Str("chart", installOpts.Chart).
		Str("name", installOpts.Name).
		Str("namespace", installOpts.Namespace).
		Str("repo", installOpts.Repo).
		Bool("wait", installOpts.Wait).
		Msg("Installing Helm chart")

	if installOpts.Name == "" {
		log.Error().
			Str("context", "HelmClient").
			Str("chart", installOpts.Chart).
			Str("name", installOpts.Name).
			Str("namespace", installOpts.Namespace).
			Str("repo", installOpts.Repo).
			Bool("wait", installOpts.Wait).
			Msg("Name is required for helm release installation")
		return nil, errors.New("name is required for helm release installation")
	}

	// Initialize action configuration with kubernetes config
	actionConfig := new(action.Configuration)
	err := hspm.initActionConfig(actionConfig, installOpts.Namespace, installOpts.KubernetesClusterAccess)
	if err != nil {
		log.Error().
			Str("context", "HelmClient").
			Str("chart", installOpts.Chart).
			Str("namespace", installOpts.Namespace).
			Err(err).
			Msg("Failed to initialize helm configuration for helm release installation")
		return nil, errors.Wrap(err, "failed to initialize helm configuration for helm release installation")
	}

	installClient, err := initInstallClient(actionConfig, installOpts)
	if err != nil {
		log.Error().
			Str("context", "HelmClient").
			Err(err).
			Msg("Failed to initialize helm install client for helm release installation")
		return nil, errors.Wrap(err, "failed to initialize helm install client for helm release installation")
	}

	values, err := hspm.GetHelmValuesFromFile(installOpts.ValuesFile)
	if err != nil {
		log.Error().
			Str("context", "HelmClient").
			Err(err).
			Msg("Failed to get Helm values from file for helm release installation")
		return nil, errors.Wrap(err, "failed to get Helm values from file for helm release installation")
	}

	chart, err := hspm.loadAndValidateChart(installClient, installOpts)
	if err != nil {
		log.Error().
			Str("context", "HelmClient").
			Err(err).
			Msg("Failed to load and validate chart for helm release installation")
		return nil, errors.Wrap(err, "failed to load and validate chart for helm release installation")
	}

	// Run the installation
	log.Info().
		Str("context", "HelmClient").
		Str("chart", installOpts.Chart).
		Str("name", installOpts.Name).
		Str("namespace", installOpts.Namespace).
		Msg("Running chart installation for helm release")

	helmRelease, err := installClient.Run(chart, values)
	if err != nil {
		log.Error().
			Str("context", "HelmClient").
			Str("chart", installOpts.Chart).
			Str("name", installOpts.Name).
			Str("namespace", installOpts.Namespace).
			Err(err).
			Msg("Failed to install helm chart for helm release installation")
		return nil, errors.Wrap(err, "helm was not able to install the chart for helm release installation")
	}

	return &release.Release{
		Name:      helmRelease.Name,
		Namespace: helmRelease.Namespace,
		Chart: release.Chart{
			Metadata: &release.Metadata{
				Name:       helmRelease.Chart.Metadata.Name,
				Version:    helmRelease.Chart.Metadata.Version,
				AppVersion: helmRelease.Chart.Metadata.AppVersion,
			},
		},
		Labels:   helmRelease.Labels,
		Version:  helmRelease.Version,
		Manifest: helmRelease.Manifest,
	}, nil
}

// loadAndValidateChart locates and loads the chart, and validates it.
// it also checks for chart dependencies and updates them if necessary.
// it returns the chart information.
func (hspm *HelmSDKPackageManager) loadAndValidateChart(installClient *action.Install, installOpts options.InstallOptions) (*chart.Chart, error) {
	// Locate and load the chart
	chartPath, err := installClient.ChartPathOptions.LocateChart(installOpts.Chart, hspm.settings)
	if err != nil {
		log.Error().
			Str("context", "HelmClient").
			Str("chart", installOpts.Chart).
			Err(err).
			Msg("Failed to locate chart for helm release installation")
		return nil, errors.Wrapf(err, "failed to find the helm chart at the path: %s/%s", installOpts.Repo, installOpts.Chart)
	}

	chartReq, err := loader.Load(chartPath)
	if err != nil {
		log.Error().
			Str("context", "HelmClient").
			Str("chart_path", chartPath).
			Err(err).
			Msg("Failed to load chart for helm release installation")
		return nil, errors.Wrap(err, "failed to load chart for helm release installation")
	}

	// Check chart dependencies to make sure all are present in /charts
	if chartDependencies := chartReq.Metadata.Dependencies; chartDependencies != nil {
		if err := action.CheckDependencies(chartReq, chartDependencies); err != nil {
			err = errors.Wrap(err, "failed to check chart dependencies for helm release installation")
			if !installClient.DependencyUpdate {
				return nil, err
			}

			log.Debug().
				Str("context", "HelmClient").
				Str("chart", installOpts.Chart).
				Msg("Updating chart dependencies for helm release installation")

			providers := getter.All(hspm.settings)
			manager := &downloader.Manager{
				Out:              os.Stdout,
				ChartPath:        chartPath,
				Keyring:          installClient.ChartPathOptions.Keyring,
				SkipUpdate:       false,
				Getters:          providers,
				RepositoryConfig: hspm.settings.RepositoryConfig,
				RepositoryCache:  hspm.settings.RepositoryCache,
				Debug:            hspm.settings.Debug,
			}
			if err := manager.Update(); err != nil {
				log.Error().
					Str("context", "HelmClient").
					Str("chart", installOpts.Chart).
					Err(err).
					Msg("Failed to update chart dependencies for helm release installation")
				return nil, errors.Wrap(err, "failed to update chart dependencies for helm release installation")
			}

			// Reload the chart with the updated Chart.lock file.
			if chartReq, err = loader.Load(chartPath); err != nil {
				log.Error().
					Str("context", "HelmClient").
					Str("chart_path", chartPath).
					Err(err).
					Msg("Failed to reload chart after dependency update for helm release installation")
				return nil, errors.Wrap(err, "failed to reload chart after dependency update for helm release installation")
			}
		}
	}

	return chartReq, nil
}

// initInstallClient initializes the install client with the given options
// and return the install client.
func initInstallClient(actionConfig *action.Configuration, installOpts options.InstallOptions) (*action.Install, error) {
	installClient := action.NewInstall(actionConfig)
	installClient.CreateNamespace = true
	installClient.DependencyUpdate = true

	installClient.ReleaseName = installOpts.Name
	installClient.Namespace = installOpts.Namespace
	installClient.ChartPathOptions.RepoURL = installOpts.Repo
	installClient.Wait = installOpts.Wait
	if installOpts.PostRenderer != "" {
		postRenderer, err := postrender.NewExec(installOpts.PostRenderer)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create post renderer")
		}
		installClient.PostRenderer = postRenderer
	}

	return installClient, nil
}
