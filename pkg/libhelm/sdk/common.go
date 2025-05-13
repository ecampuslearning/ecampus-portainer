package sdk

import (
	"os"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
)

// loadAndValidateChartWithPathOptions locates and loads the chart, and validates it.
// it also checks for chart dependencies and updates them if necessary.
// it returns the chart information.
func (hspm *HelmSDKPackageManager) loadAndValidateChartWithPathOptions(chartPathOptions *action.ChartPathOptions, chartName, version string, repoURL string, dependencyUpdate bool, operation string) (*chart.Chart, error) {
	// Locate and load the chart
	chartPathOptions.RepoURL = repoURL
	chartPathOptions.Version = version
	chartPath, err := chartPathOptions.LocateChart(chartName, hspm.settings)
	if err != nil {
		log.Error().
			Str("context", "HelmClient").
			Str("chart", chartName).
			Err(err).
			Msg("Failed to locate chart for helm " + operation)
		return nil, errors.Wrapf(err, "failed to find the helm chart at the path: %s/%s", repoURL, chartName)
	}

	chartReq, err := loader.Load(chartPath)
	if err != nil {
		log.Error().
			Str("context", "HelmClient").
			Str("chart_path", chartPath).
			Err(err).
			Msg("Failed to load chart for helm " + operation)
		return nil, errors.Wrap(err, "failed to load chart for helm "+operation)
	}

	// Check chart dependencies to make sure all are present in /charts
	if chartDependencies := chartReq.Metadata.Dependencies; chartDependencies != nil {
		if err := action.CheckDependencies(chartReq, chartDependencies); err != nil {
			err = errors.Wrap(err, "failed to check chart dependencies for helm "+operation)
			if !dependencyUpdate {
				return nil, err
			}

			log.Debug().
				Str("context", "HelmClient").
				Str("chart", chartName).
				Msg("Updating chart dependencies for helm " + operation)

			providers := getter.All(hspm.settings)
			manager := &downloader.Manager{
				Out:              os.Stdout,
				ChartPath:        chartPath,
				Keyring:          chartPathOptions.Keyring,
				SkipUpdate:       false,
				Getters:          providers,
				RepositoryConfig: hspm.settings.RepositoryConfig,
				RepositoryCache:  hspm.settings.RepositoryCache,
				Debug:            hspm.settings.Debug,
			}
			if err := manager.Update(); err != nil {
				log.Error().
					Str("context", "HelmClient").
					Str("chart", chartName).
					Err(err).
					Msg("Failed to update chart dependencies for helm " + operation)
				return nil, errors.Wrap(err, "failed to update chart dependencies for helm "+operation)
			}

			// Reload the chart with the updated Chart.lock file.
			if chartReq, err = loader.Load(chartPath); err != nil {
				log.Error().
					Str("context", "HelmClient").
					Str("chart_path", chartPath).
					Err(err).
					Msg("Failed to reload chart after dependency update for helm " + operation)
				return nil, errors.Wrap(err, "failed to reload chart after dependency update for helm "+operation)
			}
		}
	}

	return chartReq, nil
}
