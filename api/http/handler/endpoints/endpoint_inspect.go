package endpoints

import (
	"net/http"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/internal/endpointutils"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id EndpointInspect
// @summary Inspect an environment(endpoint)
// @description Retrieve details about an environment(endpoint).
// @description **Access policy**: restricted
// @tags endpoints
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "Environment(Endpoint) identifier"
// @param excludeSnapshot query bool false "if true, the snapshot data won't be retrieved"
// @param excludeSnapshotRaw query bool false "if true, the SnapshotRaw field won't be retrieved"
// @success 200 {object} portainer.Endpoint "Success"
// @failure 400 "Invalid request"
// @failure 404 "Environment(Endpoint) not found"
// @failure 500 "Server error"
// @router /endpoints/{id} [get]
func (handler *Handler) endpointInspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid environment identifier route variable", err)
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portainer.EndpointID(endpointID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an environment with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an environment with the specified identifier inside the database", err)
	}

	if err := handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint); err != nil {
		return httperror.Forbidden("Permission denied to access environment", err)
	}

	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve settings from the database", err)
	}

	hideFields(endpoint)
	endpointutils.UpdateEdgeEndpointHeartbeat(endpoint, settings)
	endpoint.ComposeSyntaxMaxVersion = handler.ComposeStackManager.ComposeSyntaxMaxVersion()

	excludeSnapshot, _ := request.RetrieveBooleanQueryParameter(r, "excludeSnapshot", true)
	excludeRaw, _ := request.RetrieveBooleanQueryParameter(r, "excludeSnapshotRaw", true)

	if !excludeSnapshot {
		if err := handler.SnapshotService.FillSnapshotData(endpoint, !excludeRaw); err != nil {
			return httperror.InternalServerError("Unable to add snapshot data", err)
		}
	}

	if endpointutils.IsKubernetesEndpoint(endpoint) {
		isServerMetricsDetected := endpoint.Kubernetes.Flags.IsServerMetricsDetected
		if !isServerMetricsDetected && handler.K8sClientFactory != nil {
			endpointutils.InitialMetricsDetection(
				endpoint,
				handler.DataStore.Endpoint(),
				handler.K8sClientFactory,
			)
		}

		isServerStorageDetected := endpoint.Kubernetes.Flags.IsServerStorageDetected
		if !isServerStorageDetected && handler.K8sClientFactory != nil {
			endpointutils.InitialStorageDetection(
				endpoint,
				handler.DataStore.Endpoint(),
				handler.K8sClientFactory,
			)
		}
	}

	// Execute endpoint pending actions
	handler.PendingActionsService.Execute(endpoint.ID)

	return response.JSON(w, endpoint)
}
