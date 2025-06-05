package sdk

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/portainer/portainer/pkg/libhelm/options"
	"github.com/rs/zerolog/log"
	"helm.sh/helm/v3/pkg/action"
)

var errRequiredShowOptions = errors.New("chart, repo and output format are required")

// Show implements the HelmPackageManager interface by using the Helm SDK to show chart information.
// It supports showing chart values, readme, and chart details based on the provided ShowOptions.
func (hspm *HelmSDKPackageManager) Show(showOpts options.ShowOptions) ([]byte, error) {
	if showOpts.Chart == "" || showOpts.Repo == "" || showOpts.OutputFormat == "" {
		log.Error().
			Str("context", "HelmClient").
			Str("chart", showOpts.Chart).
			Str("repo", showOpts.Repo).
			Str("output_format", string(showOpts.OutputFormat)).
			Msg("Missing required show options")
		return nil, errRequiredShowOptions
	}

	log.Debug().
		Str("context", "HelmClient").
		Str("chart", showOpts.Chart).
		Str("repo", showOpts.Repo).
		Str("output_format", string(showOpts.OutputFormat)).
		Msg("Showing chart information")

	// Initialize action configuration (no namespace or cluster access needed)
	actionConfig := new(action.Configuration)
	err := hspm.initActionConfig(actionConfig, "", nil)
	if err != nil {
		// error is already logged in initActionConfig
		return nil, fmt.Errorf("failed to initialize helm configuration: %w", err)
	}

	// Create temporary directory for chart download
	tempDir, err := os.MkdirTemp("", "helm-show-*")
	if err != nil {
		log.Error().
			Str("context", "HelmClient").
			Err(err).
			Msg("Failed to create temp directory")
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Create showClient action
	showClient, err := initShowClient(actionConfig, showOpts)
	if err != nil {
		log.Error().
			Str("context", "HelmClient").
			Err(err).
			Msg("Failed to initialize helm show client")
		return nil, fmt.Errorf("failed to initialize helm show client: %w", err)
	}

	// Locate and load the chart
	log.Debug().
		Str("context", "HelmClient").
		Str("chart", showOpts.Chart).
		Str("repo", showOpts.Repo).
		Msg("Locating chart")

	chartPath, err := showClient.ChartPathOptions.LocateChart(showOpts.Chart, hspm.settings)
	if err != nil {
		log.Error().
			Str("context", "HelmClient").
			Str("chart", showOpts.Chart).
			Str("repo", showOpts.Repo).
			Err(err).
			Msg("Failed to locate chart")
		return nil, fmt.Errorf("failed to locate chart: %w", err)
	}

	// Get the output based on the requested format
	output, err := showClient.Run(chartPath)
	if err != nil {
		log.Error().
			Str("context", "HelmClient").
			Str("chart_path", chartPath).
			Str("output_format", string(showOpts.OutputFormat)).
			Err(err).
			Msg("Failed to show chart info")
		return nil, fmt.Errorf("failed to show chart info: %w", err)
	}

	log.Debug().
		Str("context", "HelmClient").
		Str("chart", showOpts.Chart).
		Int("output_size", len(output)).
		Msg("Successfully retrieved chart information")

	return []byte(output), nil
}

// initShowClient initializes the show client with the given options
// and return the show client.
func initShowClient(actionConfig *action.Configuration, showOpts options.ShowOptions) (*action.Show, error) {
	showClient := action.NewShowWithConfig(action.ShowAll, actionConfig)
	showClient.ChartPathOptions.RepoURL = showOpts.Repo
	showClient.ChartPathOptions.Version = showOpts.Version // If version is "", it will use the latest version

	// Set output type based on ShowOptions
	switch showOpts.OutputFormat {
	case options.ShowAll:
		showClient.OutputFormat = action.ShowAll
	case options.ShowChart:
		showClient.OutputFormat = action.ShowChart
	case options.ShowValues:
		showClient.OutputFormat = action.ShowValues
	case options.ShowReadme:
		showClient.OutputFormat = action.ShowReadme
	default:
		log.Error().
			Str("context", "HelmClient").
			Str("output_format", string(showOpts.OutputFormat)).
			Msg("Unsupported output format")
		return nil, fmt.Errorf("unsupported output format: %s", showOpts.OutputFormat)
	}

	return showClient, nil
}
