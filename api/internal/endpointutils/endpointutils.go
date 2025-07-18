package endpointutils

import (
	"errors"
	"strings"
	"time"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
	"github.com/portainer/portainer/api/kubernetes/cli"
	log "github.com/rs/zerolog/log"
)

// TODO: this file should be migrated to package/server-ce/pkg/endpoints

// IsLocalEndpoint returns true if this is a local environment(endpoint)
func IsLocalEndpoint(endpoint *portainer.Endpoint) bool {
	return strings.HasPrefix(endpoint.URL, "unix://") ||
		strings.HasPrefix(endpoint.URL, "npipe://") ||
		endpoint.Type == portainer.KubernetesLocalEnvironment
}

// IsKubernetesEndpoint returns true if this is a kubernetes environment(endpoint)
func IsKubernetesEndpoint(endpoint *portainer.Endpoint) bool {
	return endpoint.Type == portainer.KubernetesLocalEnvironment ||
		endpoint.Type == portainer.AgentOnKubernetesEnvironment ||
		endpoint.Type == portainer.EdgeAgentOnKubernetesEnvironment
}

// IsDockerEndpoint returns true if this is a docker environment(endpoint)
func IsDockerEndpoint(endpoint *portainer.Endpoint) bool {
	return endpoint.Type == portainer.DockerEnvironment ||
		endpoint.Type == portainer.AgentOnDockerEnvironment ||
		endpoint.Type == portainer.EdgeAgentOnDockerEnvironment
}

// IsEdgeEndpoint returns true if this is an Edge endpoint
func IsEdgeEndpoint(endpoint *portainer.Endpoint) bool {
	return endpoint.Type == portainer.EdgeAgentOnDockerEnvironment || endpoint.Type == portainer.EdgeAgentOnKubernetesEnvironment
}

// IsAgentEndpoint returns true if this is an Agent endpoint
func IsAgentEndpoint(endpoint *portainer.Endpoint) bool {
	return endpoint.Type == portainer.AgentOnDockerEnvironment ||
		endpoint.Type == portainer.EdgeAgentOnDockerEnvironment ||
		endpoint.Type == portainer.AgentOnKubernetesEnvironment ||
		endpoint.Type == portainer.EdgeAgentOnKubernetesEnvironment
}

// FilterByExcludeIDs receives an environment(endpoint) array and returns a filtered array using an excludeIds param
func FilterByExcludeIDs(endpoints []portainer.Endpoint, excludeIds []portainer.EndpointID) []portainer.Endpoint {
	if len(excludeIds) == 0 {
		return endpoints
	}

	filteredEndpoints := make([]portainer.Endpoint, 0)

	idsSet := make(map[portainer.EndpointID]bool)
	for _, id := range excludeIds {
		idsSet[id] = true
	}

	for _, endpoint := range endpoints {
		if !idsSet[endpoint.ID] {
			filteredEndpoints = append(filteredEndpoints, endpoint)
		}
	}

	return filteredEndpoints
}

// EndpointSet receives an environment(endpoint) array and returns a set
func EndpointSet(endpointIDs []portainer.EndpointID) map[portainer.EndpointID]bool {
	set := map[portainer.EndpointID]bool{}

	for _, endpointID := range endpointIDs {
		set[endpointID] = true
	}

	return set
}

func InitialIngressClassDetection(endpoint *portainer.Endpoint, endpointService dataservices.EndpointService, factory *cli.ClientFactory) {
	if endpoint.Kubernetes.Flags.IsServerIngressClassDetected {
		return
	}

	defer func() {
		endpoint.Kubernetes.Flags.IsServerIngressClassDetected = true

		if err := endpointService.UpdateEndpoint(endpoint.ID, endpoint); err != nil {
			log.Debug().Err(err).Msg("unable to store found IngressClasses inside the database")
		}
	}()

	cli, err := factory.GetPrivilegedKubeClient(endpoint)
	if err != nil {
		log.Debug().Err(err).Msg("unable to create kubernetes client for ingress class detection")

		return
	}

	controllers, err := cli.GetIngressControllers()
	if err != nil {
		log.Debug().Err(err).Msg("failed to fetch ingressclasses")

		return
	}

	var updatedClasses []portainer.KubernetesIngressClassConfig
	for i := range controllers {
		var updatedClass portainer.KubernetesIngressClassConfig
		updatedClass.Name = controllers[i].ClassName
		updatedClass.Type = controllers[i].Type
		updatedClasses = append(updatedClasses, updatedClass)
	}

	endpoint.Kubernetes.Configuration.IngressClasses = updatedClasses
}

func InitialMetricsDetection(endpoint *portainer.Endpoint, endpointService dataservices.EndpointService, factory *cli.ClientFactory) {
	if endpoint.Kubernetes.Flags.IsServerMetricsDetected {
		return
	}

	defer func() {
		endpoint.Kubernetes.Flags.IsServerMetricsDetected = true
		if err := endpointService.UpdateEndpoint(endpoint.ID, endpoint); err != nil {
			log.Debug().Err(err).Msg("unable to enable UseServerMetrics inside the database")
		}
	}()

	cli, err := factory.GetPrivilegedKubeClient(endpoint)
	if err != nil {
		log.Debug().Err(err).Msg("unable to create kubernetes client for initial metrics detection")

		return
	}

	if _, err := cli.GetMetrics(); err != nil {
		log.Debug().Err(err).Msg("unable to fetch metrics: leaving metrics collection disabled.")

		return
	}

	endpoint.Kubernetes.Configuration.UseServerMetrics = true
}

func storageDetect(endpoint *portainer.Endpoint, endpointService dataservices.EndpointService, factory *cli.ClientFactory) error {
	if endpoint.Kubernetes.Flags.IsServerStorageDetected {
		return nil
	}

	defer func() {
		endpoint.Kubernetes.Flags.IsServerStorageDetected = true
		if err := endpointService.UpdateEndpoint(endpoint.ID, endpoint); err != nil {
			log.Info().Err(err).Msg("unable to enable storage class inside the database")
		}
	}()

	cli, err := factory.GetPrivilegedKubeClient(endpoint)
	if err != nil {
		log.Debug().Err(err).Msg("unable to create Kubernetes client for initial storage detection")

		return err
	}

	storage, err := cli.GetStorage()
	if err != nil {
		log.Debug().Err(err).Msg("unable to fetch storage classes: leaving storage classes disabled")

		return err
	} else if len(storage) == 0 {
		log.Info().Err(err).Msg("zero storage classes found: they may be still building, retrying in 30 seconds")

		return errors.New("zero storage classes found: they may be still building, retrying in 30 seconds")
	}

	endpoint.Kubernetes.Configuration.StorageClasses = storage

	return nil
}

func InitialStorageDetection(endpoint *portainer.Endpoint, endpointService dataservices.EndpointService, factory *cli.ClientFactory) {
	if endpoint.Kubernetes.Flags.IsServerStorageDetected {
		return
	}
	defer func() {
		endpoint.Kubernetes.Flags.IsServerStorageDetected = true
		endpointService.UpdateEndpoint(
			endpoint.ID,
			endpoint,
		)
	}()

	log.Info().Msg("attempting to detect storage classes in the cluster")

	err := storageDetect(endpoint, endpointService, factory)
	if err == nil {
		return
	}
	log.Err(err).Msg("error while detecting storage classes")

	go func() {
		// Retry after 30 seconds if the initial detection failed.
		log.Info().Msg("retrying storage detection in 30 seconds")
		time.Sleep(30 * time.Second)
		err := storageDetect(endpoint, endpointService, factory)
		log.Err(err).Msg("final error while detecting storage classes")
	}()
}

func UpdateEdgeEndpointHeartbeat(endpoint *portainer.Endpoint, settings *portainer.Settings) {
	if !IsEdgeEndpoint(endpoint) {
		return
	}

	endpoint.QueryDate = time.Now().Unix()
	checkInInterval := getEndpointCheckinInterval(endpoint, settings)
	endpoint.Heartbeat = endpoint.QueryDate-endpoint.LastCheckInDate <= int64(checkInInterval*2+20)
}

func getEndpointCheckinInterval(endpoint *portainer.Endpoint, settings *portainer.Settings) int {
	if !endpoint.Edge.AsyncMode {
		if endpoint.EdgeCheckinInterval > 0 {
			return endpoint.EdgeCheckinInterval
		}

		return settings.EdgeAgentCheckinInterval
	}

	defaultInterval := 60
	intervals := [][]int{
		{endpoint.Edge.PingInterval, settings.Edge.PingInterval},
		{endpoint.Edge.CommandInterval, settings.Edge.CommandInterval},
		{endpoint.Edge.SnapshotInterval, settings.Edge.SnapshotInterval},
	}

	for i := range intervals {
		effectiveInterval := intervals[i][0]
		if effectiveInterval <= 0 {
			effectiveInterval = intervals[i][1]
		}

		if effectiveInterval > 0 && effectiveInterval < defaultInterval {
			defaultInterval = effectiveInterval
		}
	}

	return defaultInterval
}

func InitializeEdgeEndpointRelation(endpoint *portainer.Endpoint, tx dataservices.DataStoreTx) error {
	if !IsEdgeEndpoint(endpoint) {
		return nil
	}

	relation := &portainer.EndpointRelation{
		EndpointID: endpoint.ID,
		EdgeStacks: make(map[portainer.EdgeStackID]bool),
	}

	if err := tx.EndpointRelation().Create(relation); err != nil {
		return err
	}
	return nil
}
