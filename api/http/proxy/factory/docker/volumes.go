package docker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/proxy/factory/utils"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/internal/authorization"
	"github.com/portainer/portainer/api/internal/snapshot"

	"github.com/docker/docker/client"
)

const volumeObjectIdentifier = "ResourceID"

func getInheritedResourceControlFromVolumeLabels(dockerClient *client.Client, endpointID portainer.EndpointID, volumeID string, resourceControls []portainer.ResourceControl) (*portainer.ResourceControl, error) {
	volume, err := dockerClient.VolumeInspect(context.Background(), volumeID)
	if err != nil {
		return nil, err
	}

	stackResourceID := getStackResourceIDFromLabels(volume.Labels, endpointID)
	if stackResourceID != "" {
		return authorization.GetResourceControlByResourceIDAndType(stackResourceID, portainer.StackResourceControl, resourceControls), nil
	}

	return nil, nil
}

// volumeListOperation extracts the response as a JSON object, loop through the volume array
// decorate and/or filter the volumes based on resource controls before rewriting the response.
func (transport *Transport) volumeListOperation(response *http.Response, executor *operationExecutor) error {
	// VolumeList response is a JSON object
	// https://docs.docker.com/engine/api/v1.28/#operation/VolumeList
	responseObject, err := utils.GetResponseAsJSONObject(response)
	if err != nil {
		return err
	}

	// The "Volumes" field contains the list of volumes as an array of JSON objects
	if responseObject["Volumes"] == nil {
		return utils.RewriteResponse(response, responseObject, http.StatusOK)
	}

	volumeData := responseObject["Volumes"].([]any)

	for _, volumeObject := range volumeData {
		volume := volumeObject.(map[string]any)

		if err := transport.decorateVolumeResponseWithResourceID(volume); err != nil {
			return fmt.Errorf("failed decorating volume response: %w", err)
		}
	}

	resourceOperationParameters := &resourceOperationParameters{
		resourceIdentifierAttribute: volumeObjectIdentifier,
		resourceType:                portainer.VolumeResourceControl,
		labelsObjectSelector:        selectorVolumeLabels,
	}

	volumeData, err = transport.applyAccessControlOnResourceList(resourceOperationParameters, volumeData, executor)
	if err != nil {
		return err
	}

	// Overwrite the original volume list
	responseObject["Volumes"] = volumeData

	return utils.RewriteResponse(response, responseObject, http.StatusOK)
}

// volumeInspectOperation extracts the response as a JSON object, verify that the user
// has access to the volume based on any existing resource control and either rewrite an access denied response or a decorated volume.
func (transport *Transport) volumeInspectOperation(response *http.Response, executor *operationExecutor) error {
	// VolumeInspect response is a JSON object
	// https://docs.docker.com/engine/api/v1.28/#operation/VolumeInspect
	responseObject, err := utils.GetResponseAsJSONObject(response)
	if err != nil {
		return err
	}

	if err := transport.decorateVolumeResponseWithResourceID(responseObject); err != nil {
		return fmt.Errorf("failed decorating volume response: %w", err)
	}

	resourceOperationParameters := &resourceOperationParameters{
		resourceIdentifierAttribute: volumeObjectIdentifier,
		resourceType:                portainer.VolumeResourceControl,
		labelsObjectSelector:        selectorVolumeLabels,
	}

	return transport.applyAccessControlOnResource(resourceOperationParameters, responseObject, response, executor)
}

func (transport *Transport) decorateVolumeResponseWithResourceID(responseObject map[string]any) error {
	if responseObject["Name"] == nil {
		return errors.New("missing identifier in Docker resource detail response")
	}

	resourceID, err := transport.getVolumeResourceID(responseObject["Name"].(string))
	if err != nil {
		return fmt.Errorf("failed fetching resource id: %w", err)
	}

	responseObject[volumeObjectIdentifier] = resourceID

	return nil
}

// selectorVolumeLabels retrieve the labels object associated to the volume object.
// Labels are available under the "Labels" property.
// API schema references:
// https://docs.docker.com/engine/api/v1.28/#operation/VolumeInspect
// https://docs.docker.com/engine/api/v1.28/#operation/VolumeList
func selectorVolumeLabels(responseObject map[string]any) map[string]any {
	return utils.GetJSONObject(responseObject, "Labels")
}

func (transport *Transport) decorateVolumeResourceCreationOperation(request *http.Request, resourceType portainer.ResourceControlType) (*http.Response, error) {
	tokenData, err := security.RetrieveTokenData(request)
	if err != nil {
		return nil, err
	}

	volumeID := request.Header.Get("X-Portainer-VolumeName")

	if volumeID != "" {
		agentTargetHeader := request.Header.Get(portainer.PortainerAgentTargetHeader)
		cli, err := transport.dockerClientFactory.CreateClient(transport.endpoint, agentTargetHeader, nil)
		if err != nil {
			return nil, err
		}
		defer cli.Close()

		if _, err := cli.VolumeInspect(context.Background(), volumeID); err == nil {
			return &http.Response{
				StatusCode: http.StatusConflict,
			}, errors.New("a volume with the same name already exists")
		}
	}

	response, err := transport.executeDockerRequest(request)
	if err != nil {
		return response, err
	}

	if response.StatusCode == http.StatusCreated {
		err = transport.decorateVolumeCreationResponse(response, resourceType, tokenData.ID)
	}

	return response, err
}

func (transport *Transport) decorateVolumeCreationResponse(response *http.Response, resourceType portainer.ResourceControlType, userID portainer.UserID) error {
	responseObject, err := utils.GetResponseAsJSONObject(response)
	if err != nil {
		return err
	}

	if responseObject["Name"] == nil {
		return errors.New("missing identifier in Docker resource creation response")
	}

	resourceID, err := transport.getVolumeResourceID(responseObject["Name"].(string))
	if err != nil {
		return fmt.Errorf("failed fetching resource id: %w", err)
	}

	resourceControl, err := transport.createPrivateResourceControl(resourceID, resourceType, userID)
	if err != nil {
		return err
	}

	responseObject[volumeObjectIdentifier] = resourceID

	responseObject = decorateObject(responseObject, resourceControl)

	return utils.RewriteResponse(response, responseObject, http.StatusOK)
}

func (transport *Transport) restrictedVolumeOperation(requestPath string, request *http.Request) (*http.Response, error) {
	if request.Method == http.MethodGet {
		return transport.rewriteOperation(request, transport.volumeInspectOperation)
	}

	volumeName := path.Base(requestPath)

	resourceID, err := transport.getVolumeResourceID(volumeName)
	if err != nil {
		return nil, err
	}

	if request.Method == http.MethodDelete {
		return transport.executeGenericResourceDeletionOperation(request, resourceID, volumeName, portainer.VolumeResourceControl)
	}

	return transport.restrictedResourceOperation(request, resourceID, volumeName, portainer.VolumeResourceControl, false)
}

func (transport *Transport) getVolumeResourceID(volumeName string) (string, error) {
	dockerID, err := transport.getDockerID()
	if err != nil {
		return "", fmt.Errorf("failed fetching docker id: %w", err)
	}

	return fmt.Sprintf("%s_%s", volumeName, dockerID), nil
}

func (transport *Transport) getDockerID() (string, error) {
	transport.mu.Lock()
	defer transport.mu.Unlock()

	// Local cache
	if transport.dockerID != "" {
		return transport.dockerID, nil
	}

	// Snapshot cache
	if transport.snapshotService != nil {
		endpoint := portainer.Endpoint{ID: transport.endpoint.ID}

		if err := transport.snapshotService.FillSnapshotData(&endpoint, true); err == nil && len(endpoint.Snapshots) > 0 {
			if dockerID, err := snapshot.FetchDockerID(endpoint.Snapshots[0]); err == nil {
				transport.dockerID = dockerID
				return dockerID, nil
			}
		}
	}

	// Remote value
	client, err := transport.dockerClientFactory.CreateClient(transport.endpoint, "", nil)
	if err != nil {
		return "", err
	}
	defer client.Close()

	info, err := client.Info(context.Background())
	if err != nil {
		return "", err
	}

	if info.Swarm.Cluster != nil {
		transport.dockerID = info.Swarm.Cluster.ID
		return transport.dockerID, nil
	}

	transport.dockerID = info.ID

	return transport.dockerID, nil
}
