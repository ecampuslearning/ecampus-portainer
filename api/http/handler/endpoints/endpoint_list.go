package endpoints

import (
	"net/http"
	"strconv"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/internal/endpointutils"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

const (
	EdgeDeviceIntervalMultiplier = 2
	EdgeDeviceIntervalAdd        = 20
)

// @id EndpointList
// @summary List environments(endpoints)
// @description List all environments(endpoints) based on the current user authorizations. Will
// @description return all environments(endpoints) if using an administrator or team leader account otherwise it will
// @description only return authorized environments(endpoints).
// @description **Access policy**: restricted
// @tags endpoints
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param start query int false "Start searching from"
// @param limit query int false "Limit results to this value"
// @param sort query sortKey false "Sort results by this value" Enum("Name", "Group", "Status", "LastCheckIn", "EdgeID")
// @param order query int false "Order sorted results by desc/asc" Enum("asc", "desc")
// @param search query string false "Search query"
// @param groupIds query []int false "List environments(endpoints) of these groups"
// @param status query []int false "List environments(endpoints) by this status"
// @param types query []int false "List environments(endpoints) of this type"
// @param tagIds query []int false "search environments(endpoints) with these tags (depends on tagsPartialMatch)"
// @param tagsPartialMatch query bool false "If true, will return environment(endpoint) which has one of tagIds, if false (or missing) will return only environments(endpoints) that has all the tags"
// @param endpointIds query []int false "will return only these environments(endpoints)"
// @param excludeIds query []int false "will exclude these environments(endpoints)"
// @param provisioned query bool false "If true, will return environment(endpoint) that were provisioned"
// @param agentVersions query []string false "will return only environments with on of these agent versions"
// @param edgeAsync query bool false "if exists true show only edge async agents, false show only standard edge agents. if missing, will show both types (relevant only for edge agents)"
// @param edgeDeviceUntrusted query bool false "if true, show only untrusted edge agents, if false show only trusted edge agents (relevant only for edge agents)"
// @param edgeCheckInPassedSeconds query number false "if bigger then zero, show only edge agents that checked-in in the last provided seconds (relevant only for edge agents)"
// @param excludeSnapshots query bool false "if true, the snapshot data won't be retrieved"
// @param excludeSnapshotRaw query bool false "if true, the SnapshotRaw field won't be retrieved"
// @param name query string false "will return only environments(endpoints) with this name"
// @param edgeStackId query portainer.EdgeStackID false "will return the environements of the specified edge stack"
// @param edgeStackStatus query string false "only applied when edgeStackId exists. Filter the returned environments based on their deployment status in the stack (not the environment status!)" Enum("Pending", "Ok", "Error", "Acknowledged", "Remove", "RemoteUpdateSuccess", "ImagesPulled")
// @param edgeGroupIds query []int false "List environments(endpoints) of these edge groups"
// @param excludeEdgeGroupIds query []int false "Exclude environments(endpoints) of these edge groups"
// @success 200 {array} portainer.Endpoint "Endpoints"
// @failure 500 "Server error"
// @router /endpoints [get]
func (handler *Handler) endpointList(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	start, _ := request.RetrieveNumericQueryParameter(r, "start", true)
	if start != 0 {
		start--
	}

	limit, _ := request.RetrieveNumericQueryParameter(r, "limit", true)
	sortField, _ := request.RetrieveQueryParameter(r, "sort", true)
	sortOrder, _ := request.RetrieveQueryParameter(r, "order", true)
	excludeRaw, _ := request.RetrieveBooleanQueryParameter(r, "excludeSnapshotRaw", true)

	endpointGroups, err := handler.DataStore.EndpointGroup().ReadAll()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve environment groups from the database", err)
	}

	edgeGroups, err := handler.DataStore.EdgeGroup().ReadAll()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve edge groups from the database", err)
	}

	endpoints, err := handler.DataStore.Endpoint().Endpoints()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve environments from the database", err)
	}

	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve settings from the database", err)
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	query, err := parseQuery(r)
	if err != nil {
		return httperror.BadRequest("Invalid query parameters", err)
	}

	filteredEndpoints, totalAvailableEndpoints, err := handler.filterEndpointsByQuery(endpoints, query, endpointGroups, edgeGroups, settings, securityContext)
	if err != nil {
		return httperror.InternalServerError("Unable to filter endpoints", err)
	}
	filteredEndpoints = security.FilterEndpoints(filteredEndpoints, endpointGroups, securityContext)

	sortEnvironmentsByField(filteredEndpoints, endpointGroups, getSortKey(sortField), sortOrder == "desc")

	filteredEndpointCount := len(filteredEndpoints)

	paginatedEndpoints := paginateEndpoints(filteredEndpoints, start, limit)

	for idx := range paginatedEndpoints {
		hideFields(&paginatedEndpoints[idx])

		paginatedEndpoints[idx].ComposeSyntaxMaxVersion = handler.ComposeStackManager.ComposeSyntaxMaxVersion()
		if paginatedEndpoints[idx].EdgeCheckinInterval == 0 {
			paginatedEndpoints[idx].EdgeCheckinInterval = settings.EdgeAgentCheckinInterval
		}

		endpointutils.UpdateEdgeEndpointHeartbeat(&paginatedEndpoints[idx], settings)

		if !query.excludeSnapshots {
			if err := handler.SnapshotService.FillSnapshotData(&paginatedEndpoints[idx], !excludeRaw); err != nil {
				return httperror.InternalServerError("Unable to add snapshot data", err)
			}
		}
	}

	w.Header().Set("X-Total-Count", strconv.Itoa(filteredEndpointCount))
	w.Header().Set("X-Total-Available", strconv.Itoa(totalAvailableEndpoints))

	return response.JSON(w, paginatedEndpoints)
}

func paginateEndpoints(endpoints []portainer.Endpoint, start, limit int) []portainer.Endpoint {
	if limit == 0 {
		return endpoints
	}

	endpointCount := len(endpoints)

	start = min(max(start, 0), endpointCount)
	end := min(start+limit, endpointCount)

	return endpoints[start:end]
}

func getEndpointGroup(groupID portainer.EndpointGroupID, groups []portainer.EndpointGroup) portainer.EndpointGroup {
	var endpointGroup portainer.EndpointGroup
	for _, group := range groups {
		if group.ID == groupID {
			endpointGroup = group

			break
		}
	}

	return endpointGroup
}
