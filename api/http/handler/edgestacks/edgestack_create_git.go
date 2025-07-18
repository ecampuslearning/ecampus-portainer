package edgestacks

import (
	"fmt"
	"net/http"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
	"github.com/portainer/portainer/api/filesystem"
	gittypes "github.com/portainer/portainer/api/git/types"
	httperrors "github.com/portainer/portainer/api/http/errors"
	"github.com/portainer/portainer/pkg/edge"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/validate"

	"github.com/pkg/errors"
)

type edgeStackFromGitRepositoryPayload struct {
	// Name of the stack
	// Max length: 255
	// Name must only contains lowercase characters, numbers, hyphens, or underscores
	// Name must start with a lowercase character or number
	// Example: stack-name or stack_123 or stackName
	Name string `example:"stack-name" validate:"required"`
	// URL of a Git repository hosting the Stack file
	RepositoryURL string `example:"https://github.com/openfaas/faas" validate:"required"`
	// Reference name of a Git repository hosting the Stack file
	RepositoryReferenceName string `example:"refs/heads/master"`
	// Use basic authentication to clone the Git repository
	RepositoryAuthentication bool `example:"true"`
	// Username used in basic authentication. Required when RepositoryAuthentication is true.
	RepositoryUsername string `example:"myGitUsername"`
	// Password used in basic authentication. Required when RepositoryAuthentication is true.
	RepositoryPassword string `example:"myGitPassword"`
	// Path to the Stack file inside the Git repository
	FilePathInRepository string `example:"docker-compose.yml" default:"docker-compose.yml"`
	// List of identifiers of EdgeGroups
	EdgeGroups []portainer.EdgeGroupID `example:"1" validate:"required"`
	// Deployment type to deploy this stack
	// Valid values are: 0 - 'compose', 1 - 'kubernetes'
	// compose is enabled only for docker environments
	// kubernetes is enabled only for kubernetes environments
	DeploymentType portainer.EdgeStackDeploymentType `example:"0" enums:"0,1,2"`
	// List of Registries to use for this stack
	Registries []portainer.RegistryID
	// Uses the manifest's namespaces instead of the default one
	UseManifestNamespaces bool
	// TLSSkipVerify skips SSL verification when cloning the Git repository
	TLSSkipVerify bool `example:"false"`
}

func (payload *edgeStackFromGitRepositoryPayload) Validate(r *http.Request) error {
	if len(payload.Name) == 0 {
		return httperrors.NewInvalidPayloadError("Invalid stack name")
	}

	if !edge.IsValidEdgeStackName(payload.Name) {
		return httperrors.NewInvalidPayloadError("Invalid stack name. Stack name must only consist of lowercase alpha characters, numbers, hyphens, or underscores as well as start with a lowercase character or number")
	}

	if len(payload.RepositoryURL) == 0 || !validate.IsURL(payload.RepositoryURL) {
		return httperrors.NewInvalidPayloadError("Invalid repository URL. Must correspond to a valid URL format")
	}

	if payload.RepositoryAuthentication && len(payload.RepositoryPassword) == 0 {
		return httperrors.NewInvalidPayloadError("Invalid repository credentials. Password must be specified when authentication is enabled")
	}

	if payload.DeploymentType != portainer.EdgeStackDeploymentCompose && payload.DeploymentType != portainer.EdgeStackDeploymentKubernetes {
		return httperrors.NewInvalidPayloadError("Invalid deployment type")
	}

	if len(payload.FilePathInRepository) == 0 {
		switch payload.DeploymentType {
		case portainer.EdgeStackDeploymentCompose:
			payload.FilePathInRepository = filesystem.ComposeFileDefaultName
		case portainer.EdgeStackDeploymentKubernetes:
			payload.FilePathInRepository = filesystem.ManifestFileDefaultName
		}
	}

	if len(payload.EdgeGroups) == 0 {
		return httperrors.NewInvalidPayloadError("Invalid edge groups. At least one edge group must be specified")
	}

	return nil
}

// @id EdgeStackCreateRepository
// @summary Create an EdgeStack from a git repository
// @description **Access policy**: administrator
// @tags edge_stacks
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param body body edgeStackFromGitRepositoryPayload true "stack config"
// @param dryrun query string false "if true, will not create an edge stack, but just will check the settings and return a non-persisted edge stack object"
// @success 200 {object} portainer.EdgeStack
// @failure 400 "Bad request"
// @failure 500 "Internal server error"
// @failure 503 "Edge compute features are disabled"
// @router /edge_stacks/create/repository [post]
func (handler *Handler) createEdgeStackFromGitRepository(r *http.Request, tx dataservices.DataStoreTx, dryrun bool, userID portainer.UserID) (*portainer.EdgeStack, error) {
	var payload edgeStackFromGitRepositoryPayload
	if err := request.DecodeAndValidateJSONPayload(r, &payload); err != nil {
		return nil, err
	}

	stack, err := handler.edgeStacksService.BuildEdgeStack(tx, payload.Name, payload.DeploymentType, payload.EdgeGroups, payload.Registries, payload.UseManifestNamespaces)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create edge stack object")
	}

	if dryrun {
		return stack, nil
	}

	repoConfig := gittypes.RepoConfig{
		URL:            payload.RepositoryURL,
		ReferenceName:  payload.RepositoryReferenceName,
		ConfigFilePath: payload.FilePathInRepository,
		TLSSkipVerify:  payload.TLSSkipVerify,
	}

	if payload.RepositoryAuthentication {
		repoConfig.Authentication = &gittypes.GitAuthentication{
			Username: payload.RepositoryUsername,
			Password: payload.RepositoryPassword,
		}
	}

	return handler.edgeStacksService.PersistEdgeStack(tx, stack, func(stackFolder string, relatedEndpointIds []portainer.EndpointID) (composePath string, manifestPath string, projectPath string, err error) {
		return handler.storeManifestFromGitRepository(tx, stackFolder, relatedEndpointIds, payload.DeploymentType, userID, repoConfig)
	})
}

func (handler *Handler) storeManifestFromGitRepository(tx dataservices.DataStoreTx, stackFolder string, relatedEndpointIds []portainer.EndpointID, deploymentType portainer.EdgeStackDeploymentType, currentUserID portainer.UserID, repositoryConfig gittypes.RepoConfig) (composePath, manifestPath, projectPath string, err error) {
	if hasWrongType, err := hasWrongEnvironmentType(tx.Endpoint(), relatedEndpointIds, deploymentType); err != nil {
		return "", "", "", fmt.Errorf("unable to check for existence of non fitting environments: %w", err)
	} else if hasWrongType {
		return "", "", "", errors.New("edge stack with config do not match the environment type")
	}

	projectPath = handler.FileService.GetEdgeStackProjectPath(stackFolder)
	repositoryUsername := ""
	repositoryPassword := ""
	if repositoryConfig.Authentication != nil && repositoryConfig.Authentication.Password != "" {
		repositoryUsername = repositoryConfig.Authentication.Username
		repositoryPassword = repositoryConfig.Authentication.Password
	}

	if err := handler.GitService.CloneRepository(projectPath, repositoryConfig.URL, repositoryConfig.ReferenceName, repositoryUsername, repositoryPassword, repositoryConfig.TLSSkipVerify); err != nil {
		return "", "", "", err
	}

	if deploymentType == portainer.EdgeStackDeploymentCompose {
		return repositoryConfig.ConfigFilePath, "", projectPath, nil
	}

	if deploymentType == portainer.EdgeStackDeploymentKubernetes {
		return "", repositoryConfig.ConfigFilePath, projectPath, nil
	}

	errMessage := fmt.Sprintf("unknown deployment type: %d", deploymentType)
	return "", "", "", httperrors.NewInvalidPayloadError(errMessage)
}
