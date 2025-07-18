package edgestacks

import (
	"fmt"
	"net/http"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
	"github.com/portainer/portainer/api/filesystem"
	httperrors "github.com/portainer/portainer/api/http/errors"
	"github.com/portainer/portainer/pkg/edge"
	"github.com/portainer/portainer/pkg/libhttp/request"

	"github.com/pkg/errors"
)

type edgeStackFromStringPayload struct {
	// Name of the stack
	// Max length: 255
	// Name must only contains lowercase characters, numbers, hyphens, or underscores
	// Name must start with a lowercase character or number
	// Example: stack-name or stack_123 or stackName
	Name string `example:"stack-name" validate:"required"`
	// Content of the Stack file
	StackFileContent string `example:"version: 3\n services:\n web:\n image:nginx" validate:"required"`
	// List of identifiers of EdgeGroups
	EdgeGroups []portainer.EdgeGroupID `example:"1"`
	// Deployment type to deploy this stack
	// Valid values are: 0 - 'compose', 1 - 'kubernetes'
	// compose is enabled only for docker environments
	// kubernetes is enabled only for kubernetes environments
	DeploymentType portainer.EdgeStackDeploymentType `example:"0" enums:"0,1,2"`
	// List of Registries to use for this stack
	Registries []portainer.RegistryID
	// Uses the manifest's namespaces instead of the default one
	UseManifestNamespaces bool
}

func (payload *edgeStackFromStringPayload) Validate(r *http.Request) error {
	if len(payload.Name) == 0 {
		return httperrors.NewInvalidPayloadError("Invalid stack name")
	}

	if !edge.IsValidEdgeStackName(payload.Name) {
		return httperrors.NewInvalidPayloadError("Invalid stack name. Stack name must only consist of lowercase alpha characters, numbers, hyphens, or underscores as well as start with a lowercase character or number")
	}

	if len(payload.StackFileContent) == 0 {
		return httperrors.NewInvalidPayloadError("Invalid stack file content")
	}

	if len(payload.EdgeGroups) == 0 {
		return httperrors.NewInvalidPayloadError("Edge Groups are mandatory for an Edge stack")
	}

	if payload.DeploymentType != portainer.EdgeStackDeploymentCompose && payload.DeploymentType != portainer.EdgeStackDeploymentKubernetes {
		return httperrors.NewInvalidPayloadError("Invalid deployment type")
	}

	return nil
}

// @id EdgeStackCreateString
// @summary Create an EdgeStack from a text
// @description **Access policy**: administrator
// @tags edge_stacks
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param body body edgeStackFromStringPayload true "stack config"
// @param dryrun query string false "if true, will not create an edge stack, but just will check the settings and return a non-persisted edge stack object"
// @success 200 {object} portainer.EdgeStack
// @failure 400 "Bad request"
// @failure 500 "Internal server error"
// @failure 503 "Edge compute features are disabled"
// @router /edge_stacks/create/string [post]
func (handler *Handler) createEdgeStackFromFileContent(r *http.Request, tx dataservices.DataStoreTx, dryrun bool) (*portainer.EdgeStack, error) {
	var payload edgeStackFromStringPayload
	if err := request.DecodeAndValidateJSONPayload(r, &payload); err != nil {
		return nil, err
	}

	stack, err := handler.edgeStacksService.BuildEdgeStack(tx, payload.Name, payload.DeploymentType, payload.EdgeGroups, payload.Registries, payload.UseManifestNamespaces)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Edge stack object")
	}

	if dryrun {
		return stack, nil
	}

	return handler.edgeStacksService.PersistEdgeStack(tx, stack, func(stackFolder string, relatedEndpointIds []portainer.EndpointID) (composePath string, manifestPath string, projectPath string, err error) {
		return handler.storeFileContent(tx, stackFolder, payload.DeploymentType, relatedEndpointIds, []byte(payload.StackFileContent))
	})
}

func (handler *Handler) storeFileContent(tx dataservices.DataStoreTx, stackFolder string, deploymentType portainer.EdgeStackDeploymentType, relatedEndpointIds []portainer.EndpointID, fileContent []byte) (composePath, manifestPath, projectPath string, err error) {
	if hasWrongType, err := hasWrongEnvironmentType(tx.Endpoint(), relatedEndpointIds, deploymentType); err != nil {
		return "", "", "", fmt.Errorf("unable to check for existence of non fitting environments: %w", err)
	} else if hasWrongType {
		return "", "", "", errors.New("edge stack with config do not match the environment type")
	}

	if deploymentType == portainer.EdgeStackDeploymentCompose {
		composePath = filesystem.ComposeFileDefaultName

		projectPath, err := handler.FileService.StoreEdgeStackFileFromBytes(stackFolder, composePath, fileContent)
		if err != nil {
			return "", "", "", err
		}

		return composePath, "", projectPath, nil
	}

	if deploymentType == portainer.EdgeStackDeploymentKubernetes {
		manifestPath = filesystem.ManifestFileDefaultName

		projectPath, err := handler.FileService.StoreEdgeStackFileFromBytes(stackFolder, manifestPath, fileContent)
		if err != nil {
			return "", "", "", err
		}

		return "", manifestPath, projectPath, nil
	}

	errMessage := fmt.Sprintf("invalid deployment type: %d", deploymentType)
	return "", "", "", httperrors.NewInvalidPayloadError(errMessage)
}
