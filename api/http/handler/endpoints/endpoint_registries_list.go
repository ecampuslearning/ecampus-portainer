package endpoints

import (
	"net/http"

	"github.com/pkg/errors"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/internal/endpointutils"
	"github.com/portainer/portainer/api/kubernetes"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id endpointRegistriesList
// @summary List Registries on environment
// @description List all registries based on the current user authorizations in current environment.
// @description **Access policy**: authenticated
// @tags endpoints
// @param namespace query string false "required if kubernetes environment, will show registries by namespace"
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "Environment(Endpoint) identifier"
// @success 200 {array} portainer.Registry "Success"
// @failure 500 "Server error"
// @router /endpoints/{id}/registries [get]
func (handler *Handler) endpointRegistriesList(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid environment identifier route variable", err)
	}

	var registries []portainer.Registry
	if err := handler.DataStore.ViewTx(func(tx dataservices.DataStoreTx) error {
		registries, err = handler.listRegistries(tx, r, portainer.EndpointID(endpointID))
		return err
	}); err != nil {
		var httpErr *httperror.HandlerError
		if errors.As(err, &httpErr) {
			return httpErr
		}

		return httperror.InternalServerError("Unexpected error", err)
	}

	return response.JSON(w, registries)
}

func (handler *Handler) listRegistries(tx dataservices.DataStoreTx, r *http.Request, endpointID portainer.EndpointID) ([]portainer.Registry, error) {
	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	user, err := tx.User().Read(securityContext.UserID)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve user from the database", err)
	}

	endpoint, err := tx.Endpoint().Endpoint(endpointID)
	if tx.IsErrObjectNotFound(err) {
		return nil, httperror.NotFound("Unable to find an environment with the specified identifier inside the database", err)
	} else if err != nil {
		return nil, httperror.InternalServerError("Unable to find an environment with the specified identifier inside the database", err)
	}

	isAdmin := securityContext.IsAdmin

	registries, err := tx.Registry().ReadAll()
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve registries from the database", err)
	}

	registries, handleError := handler.filterRegistriesByAccess(tx, r, registries, endpoint, user, securityContext.UserMemberships)
	if handleError != nil {
		return nil, handleError
	}

	for idx := range registries {
		hideRegistryFields(&registries[idx], !isAdmin)
	}

	return registries, err
}

func (handler *Handler) filterRegistriesByAccess(tx dataservices.DataStoreTx, r *http.Request, registries []portainer.Registry, endpoint *portainer.Endpoint, user *portainer.User, memberships []portainer.TeamMembership) ([]portainer.Registry, *httperror.HandlerError) {
	if !endpointutils.IsKubernetesEndpoint(endpoint) {
		return security.FilterRegistries(registries, user, memberships, endpoint.ID), nil
	}

	return handler.filterKubernetesEndpointRegistries(tx, r, registries, endpoint, user, memberships)
}

func (handler *Handler) filterKubernetesEndpointRegistries(tx dataservices.DataStoreTx, r *http.Request, registries []portainer.Registry, endpoint *portainer.Endpoint, user *portainer.User, memberships []portainer.TeamMembership) ([]portainer.Registry, *httperror.HandlerError) {
	namespaceParam, _ := request.RetrieveQueryParameter(r, "namespace", true)
	isAdmin, err := security.IsAdmin(r)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to check user role", err)
	}

	if namespaceParam != "" {
		if authorized, err := handler.isNamespaceAuthorized(endpoint, namespaceParam, user.ID, memberships, isAdmin); err != nil {
			return nil, httperror.NotFound("Unable to check for namespace authorization", err)
		} else if !authorized {
			return nil, httperror.Forbidden("User is not authorized to use namespace", errors.New("user is not authorized to use namespace"))
		}

		return filterRegistriesByNamespaces(registries, endpoint.ID, []string{namespaceParam}), nil
	}

	if isAdmin {
		return registries, nil
	}

	return handler.filterKubernetesRegistriesByUserRole(tx, r, registries, endpoint, user)
}

func (handler *Handler) isNamespaceAuthorized(endpoint *portainer.Endpoint, namespace string, userId portainer.UserID, memberships []portainer.TeamMembership, isAdmin bool) (bool, error) {
	if isAdmin || namespace == "" {
		return true, nil
	}

	if !endpoint.Kubernetes.Configuration.RestrictDefaultNamespace && namespace == kubernetes.DefaultNamespace {
		return true, nil
	}

	kcl, err := handler.K8sClientFactory.GetPrivilegedKubeClient(endpoint)
	if err != nil {
		return false, errors.Wrap(err, "unable to retrieve kubernetes client")
	}

	accessPolicies, err := kcl.GetNamespaceAccessPolicies()
	if err != nil {
		return false, errors.Wrap(err, "unable to retrieve environment's namespaces policies")
	}

	namespacePolicy, ok := accessPolicies[namespace]
	if !ok {
		return false, nil
	}

	return security.AuthorizedAccess(userId, memberships, namespacePolicy.UserAccessPolicies, namespacePolicy.TeamAccessPolicies), nil
}

func filterRegistriesByNamespaces(registries []portainer.Registry, endpointId portainer.EndpointID, namespaces []string) []portainer.Registry {
	filteredRegistries := []portainer.Registry{}

	for _, registry := range registries {
		if registryAccessPoliciesContainsNamespace(registry.RegistryAccesses[endpointId], namespaces) {
			filteredRegistries = append(filteredRegistries, registry)
		}
	}

	return filteredRegistries
}

func registryAccessPoliciesContainsNamespace(registryAccess portainer.RegistryAccessPolicies, namespaces []string) bool {
	for _, authorizedNamespace := range registryAccess.Namespaces {
		for _, namespace := range namespaces {
			if namespace == authorizedNamespace {
				return true
			}
		}
	}
	return false
}

func (handler *Handler) filterKubernetesRegistriesByUserRole(tx dataservices.DataStoreTx, r *http.Request, registries []portainer.Registry, endpoint *portainer.Endpoint, user *portainer.User) ([]portainer.Registry, *httperror.HandlerError) {
	err := handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint)
	if errors.Is(err, security.ErrAuthorizationRequired) {
		return nil, httperror.Forbidden("User is not authorized", err)
	}
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	userNamespaces, err := handler.userNamespaces(tx, endpoint, user)
	if err != nil {
		return nil, httperror.InternalServerError("unable to retrieve user namespaces", err)
	}

	return filterRegistriesByNamespaces(registries, endpoint.ID, userNamespaces), nil
}

func (handler *Handler) userNamespaces(tx dataservices.DataStoreTx, endpoint *portainer.Endpoint, user *portainer.User) ([]string, error) {
	kcl, err := handler.K8sClientFactory.GetPrivilegedKubeClient(endpoint)
	if err != nil {
		return nil, err
	}

	namespaceAuthorizations, err := kcl.GetNamespaceAccessPolicies()
	if err != nil {
		return nil, err
	}

	userMemberships, err := tx.TeamMembership().TeamMembershipsByUserID(user.ID)
	if err != nil {
		return nil, err
	}

	var userNamespaces []string
	for namespace, namespaceAuthorization := range namespaceAuthorizations {
		if _, ok := namespaceAuthorization.UserAccessPolicies[user.ID]; ok {
			userNamespaces = append(userNamespaces, namespace)
			continue
		}
		for _, userTeam := range userMemberships {
			if _, ok := namespaceAuthorization.TeamAccessPolicies[userTeam.TeamID]; ok {
				userNamespaces = append(userNamespaces, namespace)
				continue
			}
		}
	}
	return userNamespaces, nil
}

func hideRegistryFields(registry *portainer.Registry, hideAccesses bool) {
	registry.Password = ""
	registry.ManagementConfiguration = nil
	if hideAccesses {
		registry.RegistryAccesses = nil
	}
}
