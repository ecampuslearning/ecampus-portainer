package sdk

import (
	"os"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// GetHelmValuesFromFile reads the values file and parses it into a map[string]any
// and returns the map.
func (hspm *HelmSDKPackageManager) GetHelmValuesFromFile(valuesFile string) (map[string]any, error) {
	var vals map[string]any
	if valuesFile != "" {
		log.Debug().
			Str("context", "HelmClient").
			Str("values_file", valuesFile).
			Msg("Reading values file")

		valuesData, err := os.ReadFile(valuesFile)
		if err != nil {
			log.Error().
				Str("context", "HelmClient").
				Str("values_file", valuesFile).
				Err(err).
				Msg("Failed to read values file")
			return nil, errors.Wrap(err, "failed to read values file")
		}

		vals, err = hspm.parseValues(valuesData)
		if err != nil {
			log.Error().
				Str("context", "HelmClient").
				Str("values_file", valuesFile).
				Err(err).
				Msg("Failed to parse values file")
			return nil, errors.Wrap(err, "failed to parse values file")
		}
	}

	return vals, nil
}
