package websocket

import (
	"net"
	"net/http"
	"time"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/ws"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/validate"

	"github.com/gorilla/websocket"
)

// @summary Attach a websocket
// @description If the nodeName query parameter is present, the request will be proxied to the underlying agent environment(endpoint).
// @description If the nodeName query parameter is not specified, the request will be upgraded to the websocket protocol and
// @description an AttachStart operation HTTP request will be created and hijacked.
// @description **Access policy**: authenticated
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
// @failure 403
// @failure 404
// @failure 500
// @router /websocket/attach [get]
func (handler *Handler) websocketAttach(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	attachID, err := request.RetrieveQueryParameter(r, "id", false)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: id", err)
	}
	if !validate.IsHexadecimal(attachID) {
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
		ID:       attachID,
		nodeName: r.FormValue("nodeName"),
	}

	err = handler.handleAttachRequest(w, r, params)
	if err != nil {
		return httperror.InternalServerError("An error occurred during websocket attach operation", err)
	}

	return nil
}

func (handler *Handler) handleAttachRequest(w http.ResponseWriter, r *http.Request, params *webSocketRequestParams) error {
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

	return hijackAttachStartOperation(websocketConn, params.endpoint, params.ID)
}

func hijackAttachStartOperation(
	websocketConn *websocket.Conn,
	endpoint *portainer.Endpoint,
	attachID string,
) error {
	conn, err := initDial(endpoint)
	if err != nil {
		return err
	}

	// When we set up a TCP connection for hijack, there could be long periods
	// of inactivity (a long running command with no output) that in certain
	// network setups may cause ECONNTIMEOUT, leaving the client in an unknown
	// state. Setting TCP KeepAlive on the socket connection will prohibit
	// ECONNTIMEOUT unless the socket connection truly is broken
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	attachStartRequest, err := createAttachStartRequest(attachID)
	if err != nil {
		return err
	}

	return ws.HijackRequest(websocketConn, conn, attachStartRequest)
}

func createAttachStartRequest(attachID string) (*http.Request, error) {
	request, err := http.NewRequest("POST", "/containers/"+attachID+"/attach?stdin=1&stdout=1&stderr=1&stream=1", nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Connection", "Upgrade")
	request.Header.Set("Upgrade", "tcp")

	return request, nil
}
