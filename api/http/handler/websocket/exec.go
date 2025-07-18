package websocket

import (
	"bytes"
	"net/http"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/ws"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/validate"

	"github.com/gorilla/websocket"
	"github.com/segmentio/encoding/json"
)

type execStartOperationPayload struct {
	Tty    bool
	Detach bool
}

// @summary Execute a websocket
// @description If the nodeName query parameter is present, the request will be proxied to the underlying agent environment(endpoint).
// @description If the nodeName query parameter is not specified, the request will be upgraded to the websocket protocol and
// @description an ExecStart operation HTTP request will be created and hijacked.
// @**Access policy**: authenticated
// @security ApiKeyAuth
// @security jwt
// @tags websocket
// @accept json
// @produce json
// @param endpointId query int true "environment(endpoint) ID of the environment(endpoint) where the resource is located"
// @param nodeName query string false "node name"
// @param token query string true "JWT token used for authentication against this environment(endpoint)"
// @success 200
// @failure 400
// @failure 409
// @failure 500
// @router /websocket/exec [get]
func (handler *Handler) websocketExec(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	execID, err := request.RetrieveQueryParameter(r, "id", false)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: id", err)
	}
	if !validate.IsHexadecimal(execID) {
		return httperror.BadRequest("Invalid query parameter: id (must be hexadecimal identifier)", err)
	}

	endpointID, err := request.RetrieveNumericQueryParameter(r, "endpointId", false)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: endpointId", err)
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portainer.EndpointID(endpointID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find the environment associated to the stack inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find the environment associated to the stack inside the database", err)
	}

	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint)
	if err != nil {
		return httperror.Forbidden("Permission denied to access environment", err)
	}

	params := &webSocketRequestParams{
		endpoint: endpoint,
		ID:       execID,
		nodeName: r.FormValue("nodeName"),
	}

	err = handler.handleExecRequest(w, r, params)
	if err != nil {
		return httperror.InternalServerError("An error occurred during websocket exec operation", err)
	}

	return nil
}

func (handler *Handler) handleExecRequest(w http.ResponseWriter, r *http.Request, params *webSocketRequestParams) error {
	r.Header.Del("Origin")

	if params.endpoint.Type == portainer.AgentOnDockerEnvironment {
		return handler.proxyAgentWebsocketRequest(w, r, params)
	} else if params.endpoint.Type == portainer.EdgeAgentOnDockerEnvironment {
		return handler.proxyEdgeAgentWebsocketRequest(w, r, params)
	}

	websocketConn, err := handler.connectionUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	defer websocketConn.Close()

	return hijackExecStartOperation(websocketConn, params.endpoint, params.ID)
}

func hijackExecStartOperation(
	websocketConn *websocket.Conn,
	endpoint *portainer.Endpoint,
	execID string,
) error {
	conn, err := initDial(endpoint)
	if err != nil {
		return err
	}

	execStartRequest, err := createExecStartRequest(execID)
	if err != nil {
		return err
	}

	return ws.HijackRequest(websocketConn, conn, execStartRequest)
}

func createExecStartRequest(execID string) (*http.Request, error) {
	execStartOperationPayload := &execStartOperationPayload{
		Tty:    true,
		Detach: false,
	}

	encodedBody := bytes.NewBuffer(nil)
	err := json.NewEncoder(encodedBody).Encode(execStartOperationPayload)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", "/exec/"+execID+"/start", encodedBody)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Connection", "Upgrade")
	request.Header.Set("Upgrade", "tcp")

	return request, nil
}
