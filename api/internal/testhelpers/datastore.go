package testhelpers

import (
	"time"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/database"
	"github.com/portainer/portainer/api/dataservices"
	"github.com/portainer/portainer/api/dataservices/errors"
	"github.com/portainer/portainer/api/slicesx"
)

var _ dataservices.DataStore = &testDatastore{}

type testDatastore struct {
	customTemplate          dataservices.CustomTemplateService
	edgeGroup               dataservices.EdgeGroupService
	edgeJob                 dataservices.EdgeJobService
	edgeStack               dataservices.EdgeStackService
	edgeStackStatus         dataservices.EdgeStackStatusService
	endpoint                dataservices.EndpointService
	endpointGroup           dataservices.EndpointGroupService
	endpointRelation        dataservices.EndpointRelationService
	helmUserRepository      dataservices.HelmUserRepositoryService
	registry                dataservices.RegistryService
	resourceControl         dataservices.ResourceControlService
	apiKeyRepositoryService dataservices.APIKeyRepository
	role                    dataservices.RoleService
	sslSettings             dataservices.SSLSettingsService
	settings                dataservices.SettingsService
	snapshot                dataservices.SnapshotService
	stack                   dataservices.StackService
	tag                     dataservices.TagService
	teamMembership          dataservices.TeamMembershipService
	team                    dataservices.TeamService
	tunnelServer            dataservices.TunnelServerService
	user                    dataservices.UserService
	version                 dataservices.VersionService
	webhook                 dataservices.WebhookService
	pendingActionsService   dataservices.PendingActionsService
	connection              portainer.Connection
}

func (d *testDatastore) Backup(path string) (string, error)                  { return "", nil }
func (d *testDatastore) Open() (bool, error)                                 { return false, nil }
func (d *testDatastore) Init() error                                         { return nil }
func (d *testDatastore) Close() error                                        { return nil }
func (d *testDatastore) UpdateTx(func(dataservices.DataStoreTx) error) error { return nil }
func (d *testDatastore) ViewTx(func(dataservices.DataStoreTx) error) error   { return nil }

func (d *testDatastore) CheckCurrentEdition() error                         { return nil }
func (d *testDatastore) MigrateData() error                                 { return nil }
func (d *testDatastore) Rollback(force bool) error                          { return nil }
func (d *testDatastore) CustomTemplate() dataservices.CustomTemplateService { return d.customTemplate }
func (d *testDatastore) EdgeGroup() dataservices.EdgeGroupService           { return d.edgeGroup }
func (d *testDatastore) EdgeJob() dataservices.EdgeJobService               { return d.edgeJob }
func (d *testDatastore) EdgeStack() dataservices.EdgeStackService           { return d.edgeStack }
func (d *testDatastore) EdgeStackStatus() dataservices.EdgeStackStatusService {
	return d.edgeStackStatus
}
func (d *testDatastore) Endpoint() dataservices.EndpointService           { return d.endpoint }
func (d *testDatastore) EndpointGroup() dataservices.EndpointGroupService { return d.endpointGroup }

func (d *testDatastore) EndpointRelation() dataservices.EndpointRelationService {
	return d.endpointRelation
}

func (d *testDatastore) HelmUserRepository() dataservices.HelmUserRepositoryService {
	return d.helmUserRepository
}
func (d *testDatastore) Registry() dataservices.RegistryService { return d.registry }
func (d *testDatastore) ResourceControl() dataservices.ResourceControlService {
	return d.resourceControl
}
func (d *testDatastore) Role() dataservices.RoleService { return d.role }
func (d *testDatastore) APIKeyRepository() dataservices.APIKeyRepository {
	return d.apiKeyRepositoryService
}
func (d *testDatastore) Settings() dataservices.SettingsService             { return d.settings }
func (d *testDatastore) Snapshot() dataservices.SnapshotService             { return d.snapshot }
func (d *testDatastore) SSLSettings() dataservices.SSLSettingsService       { return d.sslSettings }
func (d *testDatastore) Stack() dataservices.StackService                   { return d.stack }
func (d *testDatastore) Tag() dataservices.TagService                       { return d.tag }
func (d *testDatastore) TeamMembership() dataservices.TeamMembershipService { return d.teamMembership }
func (d *testDatastore) Team() dataservices.TeamService                     { return d.team }
func (d *testDatastore) TunnelServer() dataservices.TunnelServerService     { return d.tunnelServer }
func (d *testDatastore) User() dataservices.UserService                     { return d.user }
func (d *testDatastore) Version() dataservices.VersionService               { return d.version }
func (d *testDatastore) Webhook() dataservices.WebhookService               { return d.webhook }

func (d *testDatastore) PendingActions() dataservices.PendingActionsService {
	return d.pendingActionsService
}

func (d *testDatastore) Connection() portainer.Connection {
	return d.connection
}

func (d *testDatastore) IsErrObjectNotFound(e error) bool {
	return false
}

func (d *testDatastore) Export(filename string) (err error) {
	return nil
}

func (d *testDatastore) Import(filename string) (err error) {
	return nil
}

type datastoreOption = func(d *testDatastore)

// NewDatastore creates new instance of testDatastore.
// Will apply options before returning, opts will be applied from left to right.
func NewDatastore(options ...datastoreOption) *testDatastore {
	conn, _ := database.NewDatabase("boltdb", "", nil)
	d := testDatastore{connection: conn}

	for _, o := range options {
		o(&d)
	}

	return &d
}

type stubSettingsService struct {
	settings *portainer.Settings
}

func (s *stubSettingsService) BucketName() string { return "settings" }

func (s *stubSettingsService) Settings() (*portainer.Settings, error) {
	return s.settings, nil
}

func (s *stubSettingsService) UpdateSettings(settings *portainer.Settings) error {
	s.settings = settings

	return nil
}

func WithSettingsService(settings *portainer.Settings) datastoreOption {
	return func(d *testDatastore) {
		d.settings = &stubSettingsService{
			settings: settings,
		}
	}
}

type stubUserService struct {
	dataservices.UserService

	users []portainer.User
}

func (s *stubUserService) BucketName() string { return "users" }
func (s *stubUserService) ReadAll(predicates ...func(portainer.User) bool) ([]portainer.User, error) {
	filtered := s.users

	for _, p := range predicates {
		filtered = slicesx.Filter(filtered, p)
	}

	return filtered, nil
}

func (s *stubUserService) UsersByRole(role portainer.UserRole) ([]portainer.User, error) {
	return s.users, nil
}

// WithUsers testDatastore option that will instruct testDatastore to return provided users
func WithUsers(us []portainer.User) datastoreOption {
	return func(d *testDatastore) {
		d.user = &stubUserService{users: us}
	}
}

type stubEdgeJobService struct {
	dataservices.EdgeJobService

	jobs []portainer.EdgeJob
}

func (s *stubEdgeJobService) BucketName() string { return "edgejobs" }
func (s *stubEdgeJobService) ReadAll(predicates ...func(portainer.EdgeJob) bool) ([]portainer.EdgeJob, error) {
	filtered := s.jobs

	for _, p := range predicates {
		filtered = slicesx.Filter(filtered, p)
	}

	return filtered, nil
}

// WithEdgeJobs option will instruct testDatastore to return provided jobs
func WithEdgeJobs(js []portainer.EdgeJob) datastoreOption {
	return func(d *testDatastore) {
		d.edgeJob = &stubEdgeJobService{jobs: js}
	}
}

type stubEndpointRelationService struct {
	dataservices.EndpointRelationService

	relations []portainer.EndpointRelation
}

func (s *stubEndpointRelationService) BucketName() string { return "endpoint_relation" }
func (s *stubEndpointRelationService) EndpointRelations() ([]portainer.EndpointRelation, error) {
	return s.relations, nil
}

func (s *stubEndpointRelationService) EndpointRelation(ID portainer.EndpointID) (*portainer.EndpointRelation, error) {
	for _, relation := range s.relations {
		if relation.EndpointID == ID {
			return &relation, nil
		}
	}

	return nil, errors.ErrObjectNotFound
}

func (s *stubEndpointRelationService) UpdateEndpointRelation(ID portainer.EndpointID, relation *portainer.EndpointRelation) error {
	for i, r := range s.relations {
		if r.EndpointID == ID {
			s.relations[i] = *relation
		}
	}

	return nil
}

func (s *stubEndpointRelationService) AddEndpointRelationsForEdgeStack(endpointIDs []portainer.EndpointID, edgeStackID portainer.EdgeStackID) error {
	for _, endpointID := range endpointIDs {
		for i, r := range s.relations {
			if r.EndpointID == endpointID {
				s.relations[i].EdgeStacks[edgeStackID] = true
			}
		}
	}

	return nil
}

func (s *stubEndpointRelationService) RemoveEndpointRelationsForEdgeStack(endpointIDs []portainer.EndpointID, edgeStackID portainer.EdgeStackID) error {
	for _, endpointID := range endpointIDs {
		for i, r := range s.relations {
			if r.EndpointID == endpointID {
				delete(s.relations[i].EdgeStacks, edgeStackID)
			}
		}
	}

	return nil
}

// WithEndpointRelations option will instruct testDatastore to return provided jobs
func WithEndpointRelations(relations []portainer.EndpointRelation) datastoreOption {
	return func(d *testDatastore) {
		d.endpointRelation = &stubEndpointRelationService{relations: relations}
	}
}

type stubEndpointService struct {
	endpoints []portainer.Endpoint
}

func (s *stubEndpointService) BucketName() string { return "endpoint" }
func (s *stubEndpointService) Endpoint(ID portainer.EndpointID) (*portainer.Endpoint, error) {
	for _, endpoint := range s.endpoints {
		if endpoint.ID == ID {
			return &endpoint, nil
		}
	}

	return nil, errors.ErrObjectNotFound
}

func (s *stubEndpointService) EndpointIDByEdgeID(edgeID string) (portainer.EndpointID, bool) {
	for _, endpoint := range s.endpoints {
		if endpoint.EdgeID == edgeID {
			return endpoint.ID, true
		}
	}

	return 0, false
}

func (s *stubEndpointService) Heartbeat(endpointID portainer.EndpointID) (int64, bool) {
	for i, endpoint := range s.endpoints {
		if endpoint.ID == endpointID {
			return s.endpoints[i].LastCheckInDate, true
		}
	}

	return 0, false
}

func (s *stubEndpointService) UpdateHeartbeat(endpointID portainer.EndpointID) {
	for i, endpoint := range s.endpoints {
		if endpoint.ID == endpointID {
			s.endpoints[i].LastCheckInDate = time.Now().Unix()
		}
	}
}

func (s *stubEndpointService) Endpoints() ([]portainer.Endpoint, error) {
	return s.endpoints, nil
}

func (s *stubEndpointService) Create(endpoint *portainer.Endpoint) error {
	s.endpoints = append(s.endpoints, *endpoint)

	return nil
}

func (s *stubEndpointService) UpdateEndpoint(ID portainer.EndpointID, endpoint *portainer.Endpoint) error {
	for i, e := range s.endpoints {
		if e.ID == ID {
			s.endpoints[i] = *endpoint
		}
	}

	return nil
}

func (s *stubEndpointService) DeleteEndpoint(ID portainer.EndpointID) error {
	endpoints := []portainer.Endpoint{}

	for _, endpoint := range s.endpoints {
		if endpoint.ID != ID {
			endpoints = append(endpoints, endpoint)
		}
	}

	s.endpoints = endpoints

	return nil
}

func (s *stubEndpointService) GetNextIdentifier() int {
	return len(s.endpoints)
}

func (s *stubEndpointService) EndpointsByTeamID(teamID portainer.TeamID) ([]portainer.Endpoint, error) {
	endpoints := make([]portainer.Endpoint, 0)

	for _, e := range s.endpoints {
		for t := range e.TeamAccessPolicies {
			if t == teamID {
				endpoints = append(endpoints, e)
			}
		}
	}

	return endpoints, nil
}

// WithEndpoints option will instruct testDatastore to return provided environments(endpoints)
func WithEndpoints(endpoints []portainer.Endpoint) datastoreOption {
	return func(d *testDatastore) {
		d.endpoint = &stubEndpointService{endpoints: endpoints}
	}
}

type stubStacksService struct {
	dataservices.StackService
	stacks []portainer.Stack
}

func (s *stubStacksService) BucketName() string { return "stacks" }

func (s *stubStacksService) Read(ID portainer.StackID) (*portainer.Stack, error) {
	for _, stack := range s.stacks {
		if stack.ID == ID {
			return &stack, nil
		}
	}

	return nil, errors.ErrObjectNotFound
}

func (s *stubStacksService) ReadAll(predicates ...func(portainer.Stack) bool) ([]portainer.Stack, error) {
	filtered := s.stacks

	for _, p := range predicates {
		filtered = slicesx.Filter(filtered, p)
	}

	return filtered, nil
}

func (s *stubStacksService) StacksByEndpointID(endpointID portainer.EndpointID) ([]portainer.Stack, error) {
	result := make([]portainer.Stack, 0)

	for _, stack := range s.stacks {
		if stack.EndpointID == endpointID {
			result = append(result, stack)
		}
	}

	return result, nil
}

func (s *stubStacksService) RefreshableStacks() ([]portainer.Stack, error) {
	result := make([]portainer.Stack, 0)

	for _, stack := range s.stacks {
		if stack.AutoUpdate != nil {
			result = append(result, stack)
		}
	}

	return result, nil
}

func (s *stubStacksService) StackByName(name string) (*portainer.Stack, error) {
	for _, stack := range s.stacks {
		if stack.Name == name {
			return &stack, nil
		}
	}

	return nil, errors.ErrObjectNotFound
}

func (s *stubStacksService) StacksByName(name string) ([]portainer.Stack, error) {
	result := make([]portainer.Stack, 0)

	for _, stack := range s.stacks {
		if stack.Name == name {
			result = append(result, stack)
		}
	}

	return result, nil
}

func (s *stubStacksService) StackByWebhookID(webhookID string) (*portainer.Stack, error) {
	for _, stack := range s.stacks {
		if stack.AutoUpdate != nil && stack.AutoUpdate.Webhook == webhookID {
			return &stack, nil
		}
	}

	return nil, errors.ErrObjectNotFound
}

func (s *stubStacksService) GetNextIdentifier() int {
	return len(s.stacks)
}

func (s *stubStacksService) Exists(ID portainer.StackID) (bool, error) {
	return false, nil
}

// WithStacks option will instruct testDatastore to return provided stacks
func WithStacks(stacks []portainer.Stack) datastoreOption {
	return func(d *testDatastore) {
		d.stack = &stubStacksService{stacks: stacks}
	}
}
