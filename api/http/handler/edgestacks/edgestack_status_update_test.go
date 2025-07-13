package edgestacks

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	portainer "github.com/portainer/portainer/api"

	"github.com/segmentio/encoding/json"
	"github.com/stretchr/testify/require"
)

// Update Status
func TestUpdateStatusAndInspect(t *testing.T) {
	handler, rawAPIKey := setupHandler(t)

	endpoint := createEndpoint(t, handler.DataStore)
	edgeStack := createEdgeStack(t, handler.DataStore, endpoint.ID)

	// Update edge stack status
	newStatus := portainer.EdgeStackStatusError
	payload := updateStatusPayload{
		Error:      "test-error",
		Status:     &newStatus,
		EndpointID: endpoint.ID,
	}

	jsonPayload, err := json.Marshal(payload)
	require.NoError(t, err)

	r := bytes.NewBuffer(jsonPayload)
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/edge_stacks/%d/status", edgeStack.ID), r)
	require.NoError(t, err)

	req.Header.Set(portainer.PortainerAgentEdgeIDHeader, endpoint.EdgeID)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected a %d response, found: %d", http.StatusOK, rec.Code)
	}

	// Get updated edge stack
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/edge_stacks/%d", edgeStack.ID), nil)
	require.NoError(t, err)

	req.Header.Add("x-api-key", rawAPIKey)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected a %d response, found: %d", http.StatusOK, rec.Code)
	}

	updatedStack := portainer.EdgeStack{}
	err = json.NewDecoder(rec.Body).Decode(&updatedStack)
	require.NoError(t, err)

	endpointStatus, ok := updatedStack.Status[payload.EndpointID]
	require.True(t, ok)

	lastStatus := endpointStatus.Status[len(endpointStatus.Status)-1]

	if len(endpointStatus.Status) == len(edgeStack.Status[payload.EndpointID].Status) {
		t.Fatal("expected status array to be updated")
	}

	if lastStatus.Type != *payload.Status {
		t.Fatalf("expected EdgeStackStatusType %d, found %d", *payload.Status, lastStatus.Type)
	}

	if endpointStatus.EndpointID != payload.EndpointID {
		t.Fatalf("expected EndpointID %d, found %d", payload.EndpointID, endpointStatus.EndpointID)
	}
}

func TestUpdateStatusWithInvalidPayload(t *testing.T) {
	handler, _ := setupHandler(t)

	endpoint := createEndpoint(t, handler.DataStore)
	edgeStack := createEdgeStack(t, handler.DataStore, endpoint.ID)

	// Update edge stack status
	statusError := portainer.EdgeStackStatusError
	statusOk := portainer.EdgeStackStatusDeploymentReceived
	cases := []struct {
		Name                 string
		Payload              updateStatusPayload
		ExpectedErrorMessage string
		ExpectedStatusCode   int
	}{
		{
			"Update with nil Status",
			updateStatusPayload{
				Error:      "test-error",
				Status:     nil,
				EndpointID: endpoint.ID,
			},
			"Invalid status",
			400,
		},
		{
			"Update with error status and empty error message",
			updateStatusPayload{
				Error:      "",
				Status:     &statusError,
				EndpointID: endpoint.ID,
			},
			"Error message is mandatory when status is error",
			400,
		},
		{
			"Update with missing EndpointID",
			updateStatusPayload{
				Error:      "",
				Status:     &statusOk,
				EndpointID: 0,
			},
			"Invalid EnvironmentID",
			400,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			jsonPayload, err := json.Marshal(tc.Payload)
			require.NoError(t, err)

			r := bytes.NewBuffer(jsonPayload)
			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/edge_stacks/%d/status", edgeStack.ID), r)
			require.NoError(t, err)

			req.Header.Set(portainer.PortainerAgentEdgeIDHeader, endpoint.EdgeID)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tc.ExpectedStatusCode {
				t.Fatalf("expected a %d response, found: %d", tc.ExpectedStatusCode, rec.Code)
			}
		})
	}
}
