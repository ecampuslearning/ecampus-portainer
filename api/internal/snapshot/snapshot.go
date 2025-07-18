package snapshot

import (
	"context"
	"crypto/tls"
	"errors"
	"time"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/agent"
	"github.com/portainer/portainer/api/crypto"
	"github.com/portainer/portainer/api/dataservices"
	"github.com/portainer/portainer/api/pendingactions"
	endpointsutils "github.com/portainer/portainer/pkg/endpoints"

	"github.com/rs/zerolog/log"
)

// Service represents a service to manage environment(endpoint) snapshots.
// It provides an interface to start background snapshots as well as
// specific Docker/Kubernetes environment(endpoint) snapshot methods.
type Service struct {
	dataStore                 dataservices.DataStore
	snapshotIntervalCh        chan time.Duration
	snapshotIntervalInSeconds float64
	dockerSnapshotter         portainer.DockerSnapshotter
	kubernetesSnapshotter     portainer.KubernetesSnapshotter
	shutdownCtx               context.Context
	pendingActionsService     *pendingactions.PendingActionsService
}

// NewService creates a new instance of a service
func NewService(
	snapshotIntervalFromFlag string,
	dataStore dataservices.DataStore,
	dockerSnapshotter portainer.DockerSnapshotter,
	kubernetesSnapshotter portainer.KubernetesSnapshotter,
	shutdownCtx context.Context,
	pendingActionsService *pendingactions.PendingActionsService,
) (*Service, error) {
	interval, err := parseSnapshotFrequency(snapshotIntervalFromFlag, dataStore)
	if err != nil {
		return nil, err
	}

	return &Service{
		dataStore:                 dataStore,
		snapshotIntervalCh:        make(chan time.Duration),
		snapshotIntervalInSeconds: interval,
		dockerSnapshotter:         dockerSnapshotter,
		kubernetesSnapshotter:     kubernetesSnapshotter,
		shutdownCtx:               shutdownCtx,
		pendingActionsService:     pendingActionsService,
	}, nil
}

// NewBackgroundSnapshotter queues snapshots of existing edge environments that
// do not have one already
func NewBackgroundSnapshotter(dataStore dataservices.DataStore, tunnelService portainer.ReverseTunnelService) {
	if err := dataStore.ViewTx(func(tx dataservices.DataStoreTx) error {
		endpoints, err := tx.Endpoint().Endpoints()
		if err != nil {
			return err
		}

		for _, e := range endpoints {
			if !endpointsutils.HasDirectConnectivity(&e) {
				continue
			}

			s, err := tx.Snapshot().Read(e.ID)
			if dataservices.IsErrObjectNotFound(err) ||
				(err == nil && s.Docker == nil && s.Kubernetes == nil) {
				if err := tunnelService.Open(&e); err != nil {
					log.Error().Err(err).Msg("could not open the tunnel")
				}
			}
		}

		return nil
	}); err != nil {
		log.Error().Err(err).Msg("background snapshotter failure")

		return
	}
}

func parseSnapshotFrequency(snapshotInterval string, dataStore dataservices.DataStore) (float64, error) {
	if snapshotInterval == "" {
		settings, err := dataStore.Settings().Settings()
		if err != nil {
			return 0, err
		}

		snapshotInterval = settings.SnapshotInterval
		if snapshotInterval == "" {
			snapshotInterval = portainer.DefaultSnapshotInterval
		}
	}

	snapshotFrequency, err := time.ParseDuration(snapshotInterval)
	if err != nil {
		return 0, err
	}

	return snapshotFrequency.Seconds(), nil
}

// Start will start a background routine to execute periodic snapshots of environments(endpoints)
func (service *Service) Start() {
	go service.startSnapshotLoop()
}

// SetSnapshotInterval sets the snapshot interval and resets the service
func (service *Service) SetSnapshotInterval(snapshotInterval string) error {
	interval, err := time.ParseDuration(snapshotInterval)
	if err != nil {
		return err
	}

	service.snapshotIntervalCh <- interval

	return nil
}

// SupportDirectSnapshot checks whether an environment(endpoint) can be used to trigger a direct a snapshot.
// It is mostly true for all environments(endpoints) except Edge and Azure environments(endpoints).
func SupportDirectSnapshot(endpoint *portainer.Endpoint) bool {
	switch endpoint.Type {
	case portainer.EdgeAgentOnDockerEnvironment, portainer.EdgeAgentOnKubernetesEnvironment, portainer.AzureEnvironment:
		return false
	}

	return true
}

// SnapshotEndpoint will create a snapshot of the environment(endpoint) based on the environment(endpoint) type.
// If the snapshot is a success, it will be associated to the environment(endpoint).
func (service *Service) SnapshotEndpoint(endpoint *portainer.Endpoint) error {
	if endpoint.Type == portainer.AgentOnDockerEnvironment || endpoint.Type == portainer.AgentOnKubernetesEnvironment {
		var err error
		var tlsConfig *tls.Config

		if endpoint.TLSConfig.TLS {
			tlsConfig, err = crypto.CreateTLSConfigurationFromDisk(endpoint.TLSConfig.TLSCACertPath, endpoint.TLSConfig.TLSCertPath, endpoint.TLSConfig.TLSKeyPath, endpoint.TLSConfig.TLSSkipVerify)
			if err != nil {
				return err
			}
		}

		_, version, err := agent.GetAgentVersionAndPlatform(endpoint.URL, tlsConfig)
		if err != nil {
			return err
		}

		endpoint.Agent.Version = version
	}

	switch endpoint.Type {
	case portainer.AzureEnvironment:
		return nil
	case portainer.KubernetesLocalEnvironment, portainer.AgentOnKubernetesEnvironment, portainer.EdgeAgentOnKubernetesEnvironment:
		return service.snapshotKubernetesEndpoint(endpoint)
	}

	return service.snapshotDockerEndpoint(endpoint)
}

func (service *Service) Create(snapshot portainer.Snapshot) error {
	return service.dataStore.Snapshot().Create(&snapshot)
}

func (service *Service) FillSnapshotData(endpoint *portainer.Endpoint, includeRaw bool) error {
	return FillSnapshotData(service.dataStore, endpoint, includeRaw)
}

func (service *Service) snapshotKubernetesEndpoint(endpoint *portainer.Endpoint) error {
	kubernetesSnapshot, err := service.kubernetesSnapshotter.CreateSnapshot(endpoint)
	if err != nil {
		return err
	}

	if kubernetesSnapshot != nil {
		snapshot := &portainer.Snapshot{EndpointID: endpoint.ID, Kubernetes: kubernetesSnapshot}

		return service.dataStore.Snapshot().Create(snapshot)
	}

	return nil
}

func (service *Service) snapshotDockerEndpoint(endpoint *portainer.Endpoint) error {
	dockerSnapshot, err := service.dockerSnapshotter.CreateSnapshot(endpoint)
	if err != nil {
		return err
	}

	if err := validateContainerEngineCompatibility(endpoint, dockerSnapshot); err != nil {
		return err
	}

	if dockerSnapshot != nil {
		snapshot := &portainer.Snapshot{EndpointID: endpoint.ID, Docker: dockerSnapshot}

		return service.dataStore.Snapshot().Create(snapshot)
	}

	return nil
}

func validateContainerEngineCompatibility(endpoint *portainer.Endpoint, dockerSnapshot *portainer.DockerSnapshot) error {
	if endpoint.ContainerEngine == portainer.ContainerEngineDocker && dockerSnapshot.IsPodman {
		err := errors.New("the Docker environment option doesn't support Podman environments. Please select the Podman option instead.")
		log.Error().Err(err).Str("endpoint", endpoint.Name).Msg(err.Error())
		return err
	}
	if endpoint.ContainerEngine == portainer.ContainerEnginePodman && !dockerSnapshot.IsPodman {
		err := errors.New("the Podman environment option doesn't support Docker environments. Please select the Docker option instead.")
		log.Error().Err(err).Str("endpoint", endpoint.Name).Msg(err.Error())
		return err
	}
	return nil
}

func (service *Service) startSnapshotLoop() {
	ticker := time.NewTicker(time.Duration(service.snapshotIntervalInSeconds) * time.Second)

	err := service.snapshotEndpoints()
	if err != nil {
		log.Error().Err(err).Msg("background schedule error (environment snapshot)")
	}

	for {
		select {
		case <-ticker.C:
			err := service.snapshotEndpoints()
			if err != nil {
				log.Error().Err(err).Msg("background schedule error (environment snapshot)")
			}
		case <-service.shutdownCtx.Done():
			log.Debug().Msg("shutting down snapshotting")
			ticker.Stop()

			return
		case interval := <-service.snapshotIntervalCh:
			ticker.Reset(interval)
		}
	}
}

func (service *Service) snapshotEndpoints() error {
	endpoints, err := service.dataStore.Endpoint().Endpoints()
	if err != nil {
		return err
	}

	for _, endpoint := range endpoints {
		if !SupportDirectSnapshot(&endpoint) || endpoint.URL == "" {
			continue
		}

		snapshotError := service.SnapshotEndpoint(&endpoint)

		if err := service.dataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
			updateEndpointStatus(tx, &endpoint, snapshotError, service.pendingActionsService)

			return nil
		}); err != nil {
			log.Error().
				Err(err).
				Int("endpoint_id", int(endpoint.ID)).
				Msg("unable to update environment status")
		}
	}

	return nil
}

func updateEndpointStatus(tx dataservices.DataStoreTx, endpoint *portainer.Endpoint, snapshotError error, pendingActionsService *pendingactions.PendingActionsService) {
	latestEndpointReference, err := tx.Endpoint().Endpoint(endpoint.ID)
	if latestEndpointReference == nil {
		log.Debug().
			Str("endpoint", endpoint.Name).
			Str("URL", endpoint.URL).Err(err).
			Msg("background schedule error (environment snapshot), environment not found inside the database anymore")

		return
	}

	latestEndpointReference.Status = portainer.EndpointStatusUp

	if snapshotError != nil {
		log.Debug().
			Str("endpoint", endpoint.Name).
			Str("URL", endpoint.URL).Err(err).
			Msg("background schedule error (environment snapshot), unable to create snapshot")

		latestEndpointReference.Status = portainer.EndpointStatusDown
	}

	latestEndpointReference.Agent.Version = endpoint.Agent.Version

	if err := tx.Endpoint().UpdateEndpoint(latestEndpointReference.ID, latestEndpointReference); err != nil {
		log.Debug().
			Str("endpoint", endpoint.Name).
			Str("URL", endpoint.URL).Err(err).
			Msg("background schedule error (environment snapshot), unable to update environment")
	}

	// Run the pending actions
	if latestEndpointReference.Status == portainer.EndpointStatusUp {
		pendingActionsService.Execute(endpoint.ID)
	}
}

// FetchDockerID fetches info.Swarm.Cluster.ID if environment(endpoint) is swarm and info.ID otherwise
func FetchDockerID(snapshot portainer.DockerSnapshot) (string, error) {
	info := snapshot.SnapshotRaw.Info

	if !snapshot.Swarm {
		return info.ID, nil
	}

	if info.Swarm.Cluster == nil {
		return "", errors.New("swarm environment is missing cluster info snapshot")
	}

	return info.Swarm.Cluster.ID, nil
}

func FillSnapshotData(tx dataservices.DataStoreTx, endpoint *portainer.Endpoint, includeRaw bool) error {
	var snapshot *portainer.Snapshot
	var err error

	if includeRaw {
		snapshot, err = tx.Snapshot().Read(endpoint.ID)
	} else {
		snapshot, err = tx.Snapshot().ReadWithoutSnapshotRaw(endpoint.ID)
	}

	if tx.IsErrObjectNotFound(err) {
		endpoint.Snapshots = []portainer.DockerSnapshot{}
		endpoint.Kubernetes.Snapshots = []portainer.KubernetesSnapshot{}

		return nil
	} else if err != nil {
		return err
	}

	if snapshot.Docker != nil {
		endpoint.Snapshots = []portainer.DockerSnapshot{*snapshot.Docker}
	}

	if snapshot.Kubernetes != nil {
		endpoint.Kubernetes.Snapshots = []portainer.KubernetesSnapshot{*snapshot.Kubernetes}
	}

	return nil
}
