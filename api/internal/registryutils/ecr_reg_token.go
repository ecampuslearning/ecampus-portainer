package registryutils

import (
	"time"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/aws/ecr"
	"github.com/portainer/portainer/api/dataservices"

	"github.com/rs/zerolog/log"
)

func isRegTokenValid(registry *portainer.Registry) (valid bool) {
	return registry.AccessToken != "" && registry.AccessTokenExpiry > time.Now().Unix()
}

func doGetRegToken(tx dataservices.DataStoreTx, registry *portainer.Registry) error {
	ecrClient := ecr.NewService(registry.Username, registry.Password, registry.Ecr.Region)
	accessToken, expiryAt, err := ecrClient.GetAuthorizationToken()
	if err != nil {
		return err
	}

	registry.AccessToken = *accessToken
	registry.AccessTokenExpiry = expiryAt.Unix()

	return tx.Registry().Update(registry.ID, registry)
}

func parseRegToken(registry *portainer.Registry) (username, password string, err error) {
	return ecr.NewService(registry.Username, registry.Password, registry.Ecr.Region).
		ParseAuthorizationToken(registry.AccessToken)
}

func EnsureRegTokenValid(tx dataservices.DataStoreTx, registry *portainer.Registry) error {
	if registry.Type != portainer.EcrRegistry {
		return nil
	}

	if isRegTokenValid(registry) {
		log.Debug().Msg("current ECR token is still valid")

		return nil
	}

	if err := doGetRegToken(tx, registry); err != nil {
		log.Debug().Msg("refresh ECR token")

		return err
	}

	return nil
}

func GetRegEffectiveCredential(registry *portainer.Registry) (username, password string, err error) {
	username = registry.Username
	password = registry.Password

	if registry.Type == portainer.EcrRegistry {
		username, password, err = parseRegToken(registry)
	}

	return
}

// PrepareRegistryCredentials consolidates the common pattern of ensuring valid ECR token
// and setting effective credentials on the registry when authentication is enabled.
// This function modifies the registry in-place by setting Username and Password to the effective values.
func PrepareRegistryCredentials(tx dataservices.DataStoreTx, registry *portainer.Registry) error {
	if !registry.Authentication {
		return nil
	}

	if err := EnsureRegTokenValid(tx, registry); err != nil {
		return err
	}

	username, password, err := GetRegEffectiveCredential(registry)
	if err != nil {
		return err
	}

	registry.Username = username
	registry.Password = password

	return nil
}
