package portainer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/api/types/volume"
	gittypes "github.com/portainer/portainer/api/git/types"
	models "github.com/portainer/portainer/api/http/models/kubernetes"
	"github.com/portainer/portainer/pkg/featureflags"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/segmentio/encoding/json"

	"golang.org/x/oauth2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/version"
)

type (
	// AccessPolicy represent a policy that can be associated to a user or team
	AccessPolicy struct {
		// Role identifier. Reference the role that will be associated to this access policy
		RoleID RoleID `json:"RoleId" example:"1"`
	}

	// AgentPlatform represents a platform type for an Agent
	AgentPlatform int

	// AuthenticationMethod represents the authentication method used to authenticate a user
	AuthenticationMethod int

	// Authorization represents an authorization associated to an operation
	Authorization string

	// Authorizations represents a set of authorizations associated to a role
	Authorizations map[Authorization]bool

	// AutoUpdateSettings represents the git auto sync config for stack deployment
	AutoUpdateSettings struct {
		// Auto update interval
		Interval string `example:"1m30s"`
		// A UUID generated from client
		Webhook string `example:"05de31a2-79fa-4644-9c12-faa67e5c49f0"`
		// Autoupdate job id
		JobID string `example:"15"`
		// Force update ignores repo changes
		ForceUpdate bool `example:"false"`
		// Pull latest image
		ForcePullImage bool `example:"false"`
	}

	// AzureCredentials represents the credentials used to connect to an Azure
	// environment(endpoint).
	AzureCredentials struct {
		// Azure application ID
		ApplicationID string `json:"ApplicationID" example:"eag7cdo9-o09l-9i83-9dO9-f0b23oe78db4"`
		// Azure tenant ID
		TenantID string `json:"TenantID" example:"34ddc78d-4fel-2358-8cc1-df84c8o839f5"`
		// Azure authentication key
		AuthenticationKey string `json:"AuthenticationKey" example:"cOrXoK/1D35w8YQ8nH1/8ZGwzz45JIYD5jxHKXEQknk="`
	}

	// OpenAMTConfiguration represents the credentials and configurations used to connect to an OpenAMT MPS server
	OpenAMTConfiguration struct {
		Enabled          bool   `json:"enabled"`
		MPSServer        string `json:"mpsServer"`
		MPSUser          string `json:"mpsUser"`
		MPSPassword      string `json:"mpsPassword"`
		MPSToken         string `json:"mpsToken"` // retrieved from API
		CertFileName     string `json:"certFileName"`
		CertFileContent  string `json:"certFileContent"`
		CertFilePassword string `json:"certFilePassword"`
		DomainName       string `json:"domainName"`
	}

	// OpenAMTDeviceInformation represents an AMT managed device information
	OpenAMTDeviceInformation struct {
		GUID             string                        `json:"guid"`
		HostName         string                        `json:"hostname"`
		ConnectionStatus bool                          `json:"connectionStatus"`
		PowerState       PowerState                    `json:"powerState"`
		EnabledFeatures  *OpenAMTDeviceEnabledFeatures `json:"features"`
	}

	// OpenAMTDeviceEnabledFeatures represents an AMT managed device features information
	OpenAMTDeviceEnabledFeatures struct {
		Redirection bool   `json:"redirection"`
		KVM         bool   `json:"KVM"`
		SOL         bool   `json:"SOL"`
		IDER        bool   `json:"IDER"`
		UserConsent string `json:"userConsent"`
	}

	// PowerState represents an AMT managed device power state
	PowerState int

	// CLIFlags represents the available flags on the CLI
	CLIFlags struct {
		Addr                      *string
		AddrHTTPS                 *string
		TunnelAddr                *string
		TunnelPort                *string
		AdminPassword             *string
		AdminPasswordFile         *string
		Assets                    *string
		CSP                       *bool
		Data                      *string
		FeatureFlags              *[]string
		EnableEdgeComputeFeatures *bool
		EndpointURL               *string
		Labels                    *[]Pair
		Logo                      *string
		NoAnalytics               *bool
		Templates                 *string
		TLS                       *bool
		TLSSkipVerify             *bool
		TLSCacert                 *string
		TLSCert                   *string
		TLSKey                    *string
		HTTPDisabled              *bool
		HTTPEnabled               *bool
		SSL                       *bool
		SSLCert                   *string
		SSLKey                    *string
		Rollback                  *bool
		SnapshotInterval          *string
		BaseURL                   *string
		InitialMmapSize           *int
		MaxBatchSize              *int
		MaxBatchDelay             *time.Duration
		SecretKeyName             *string
		LogLevel                  *string
		LogMode                   *string
		KubectlShellImage         *string
		PullLimitCheckDisabled    *bool
		TrustedOrigins            *string
	}

	// CustomTemplateVariableDefinition
	CustomTemplateVariableDefinition struct {
		Name         string `json:"name" example:"MY_VAR"`
		Label        string `json:"label" example:"My Variable"`
		DefaultValue string `json:"defaultValue" example:"default value"`
		Description  string `json:"description" example:"Description"`
	}

	// CustomTemplate represents a custom template
	CustomTemplate struct {
		// CustomTemplate Identifier
		ID CustomTemplateID `json:"Id" example:"1"`
		// Title of the template
		Title string `json:"Title" example:"Nginx"`
		// Description of the template
		Description string `json:"Description" example:"High performance web server"`
		// Path on disk to the repository hosting the Stack file
		ProjectPath string `json:"ProjectPath" example:"/data/custom_template/3"`
		// Path to the Stack file
		EntryPoint string `json:"EntryPoint" example:"docker-compose.yml"`
		// User identifier who created this template
		CreatedByUserID UserID `json:"CreatedByUserId" example:"3"`
		// A note that will be displayed in the UI. Supports HTML content
		Note string `json:"Note" example:"This is my <b>custom</b> template"`
		// Platform associated to the template.
		// Valid values are: 1 - 'linux', 2 - 'windows'
		Platform CustomTemplatePlatform `json:"Platform" example:"1" enums:"1,2"`
		// URL of the template's logo
		Logo string `json:"Logo" example:"https://portainer.io/img/logo.svg"`
		// Type of created stack:
		// * 1 - swarm
		// * 2 - compose
		// * 3 - kubernetes
		Type            StackType        `json:"Type" example:"1" enums:"1,2,3"`
		ResourceControl *ResourceControl `json:"ResourceControl"`
		Variables       []CustomTemplateVariableDefinition
		GitConfig       *gittypes.RepoConfig `json:"GitConfig"`
		// IsComposeFormat indicates if the Kubernetes template is created from a Docker Compose file
		IsComposeFormat bool `example:"false"`
		// EdgeTemplate indicates if this template purpose for Edge Stack
		EdgeTemplate bool `example:"false"`
	}

	// CustomTemplateID represents a custom template identifier
	CustomTemplateID int

	// CustomTemplatePlatform represents a custom template platform
	CustomTemplatePlatform int

	// DiagnosticsData represents the diagnostics data for an environment
	// this contains the logs, telnet, traceroute, dns and proxy information
	// which will be part of the DockerSnapshot and KubernetesSnapshot structs
	DiagnosticsData struct {
		Log    string            `json:"Log,omitempty"`
		Telnet map[string]string `json:"Telnet,omitempty"`
		DNS    map[string]string `json:"DNS,omitempty"`
		Proxy  map[string]string `json:"Proxy,omitempty"`
	}

	// DockerHub represents all the required information to connect and use the
	// Docker Hub
	DockerHub struct {
		// Is authentication against DockerHub enabled
		Authentication bool `json:"Authentication" example:"true"`
		// Username used to authenticate against the DockerHub
		Username string `json:"Username" example:"user"`
		// Password used to authenticate against the DockerHub
		Password string `json:"Password,omitempty" example:"passwd"`
	}

	// DockerSnapshot represents a snapshot of a specific Docker environment(endpoint) at a specific time
	DockerSnapshot struct {
		Time                    int64               `json:"Time"`
		DockerVersion           string              `json:"DockerVersion"`
		Swarm                   bool                `json:"Swarm"`
		TotalCPU                int                 `json:"TotalCPU"`
		TotalMemory             int64               `json:"TotalMemory"`
		ContainerCount          int                 `json:"ContainerCount"`
		RunningContainerCount   int                 `json:"RunningContainerCount"`
		StoppedContainerCount   int                 `json:"StoppedContainerCount"`
		HealthyContainerCount   int                 `json:"HealthyContainerCount"`
		UnhealthyContainerCount int                 `json:"UnhealthyContainerCount"`
		VolumeCount             int                 `json:"VolumeCount"`
		ImageCount              int                 `json:"ImageCount"`
		ServiceCount            int                 `json:"ServiceCount"`
		StackCount              int                 `json:"StackCount"`
		SnapshotRaw             DockerSnapshotRaw   `json:"DockerSnapshotRaw"`
		NodeCount               int                 `json:"NodeCount"`
		GpuUseAll               bool                `json:"GpuUseAll"`
		GpuUseList              []string            `json:"GpuUseList"`
		IsPodman                bool                `json:"IsPodman"`
		DiagnosticsData         *DiagnosticsData    `json:"DiagnosticsData"`
		PerformanceMetrics      *PerformanceMetrics `json:"PerformanceMetrics"`
	}

	// PerformanceMetrics represents the performance metrics of a Docker, Swarm, Podman, and Kubernetes environments
	PerformanceMetrics struct {
		CPUUsage     float64 `json:"CPUUsage,omitempty"`
		MemoryUsage  float64 `json:"MemoryUsage,omitempty"`
		NetworkUsage float64 `json:"NetworkUsage,omitempty"`
	}

	// DockerContainerSnapshot is an extent of Docker's Container struct
	// It contains some information of Docker's ContainerJSON struct
	DockerContainerSnapshot struct {
		types.Container
		Env []string `json:"Env,omitempty"` // EE-5240
	}

	// DockerSnapshotRaw represents all the information related to a snapshot as returned by the Docker API
	DockerSnapshotRaw struct {
		Containers []DockerContainerSnapshot `json:"Containers" swaggerignore:"true"`
		Volumes    volume.ListResponse       `json:"Volumes" swaggerignore:"true"`
		Networks   []network.Summary         `json:"Networks" swaggerignore:"true"`
		Images     []image.Summary           `json:"Images" swaggerignore:"true"`
		Info       system.Info               `json:"Info" swaggerignore:"true"`
		Version    types.Version             `json:"Version" swaggerignore:"true"`
	}

	// EdgeGroup represents an Edge group
	EdgeGroup struct {
		// EdgeGroup Identifier
		ID           EdgeGroupID  `json:"Id" example:"1"`
		Name         string       `json:"Name"`
		Dynamic      bool         `json:"Dynamic"`
		TagIDs       []TagID      `json:"TagIds"`
		Endpoints    []EndpointID `json:"Endpoints"`
		PartialMatch bool         `json:"PartialMatch"`
	}

	// EdgeGroupID represents an Edge group identifier
	EdgeGroupID int

	// EdgeJob represents a job that can run on Edge environments(endpoints).
	EdgeJob struct {
		// EdgeJob Identifier
		ID             EdgeJobID                          `json:"Id" example:"1"`
		Created        int64                              `json:"Created"`
		CronExpression string                             `json:"CronExpression"`
		Endpoints      map[EndpointID]EdgeJobEndpointMeta `json:"Endpoints"`
		EdgeGroups     []EdgeGroupID                      `json:"EdgeGroups"`
		Name           string                             `json:"Name"`
		ScriptPath     string                             `json:"ScriptPath"`
		Recurring      bool                               `json:"Recurring"`
		Version        int                                `json:"Version"`

		// Field used for log collection of Endpoints belonging to EdgeGroups
		GroupLogsCollection map[EndpointID]EdgeJobEndpointMeta
	}

	// EdgeJobEndpointMeta represents a meta data object for an Edge job and Environment(Endpoint) relation
	EdgeJobEndpointMeta struct {
		LogsStatus  EdgeJobLogsStatus
		CollectLogs bool
	}

	// EdgeJobID represents an Edge job identifier
	EdgeJobID int

	// EdgeJobLogsStatus represent status of logs collection job
	EdgeJobLogsStatus int

	// EdgeSchedule represents a scheduled job that can run on Edge environments(endpoints).
	//
	// Deprecated: in favor of EdgeJob
	EdgeSchedule struct {
		// EdgeSchedule Identifier
		ID             ScheduleID   `json:"Id" example:"1"`
		CronExpression string       `json:"CronExpression"`
		Script         string       `json:"Script"`
		Version        int          `json:"Version"`
		Endpoints      []EndpointID `json:"Endpoints"`
	}

	// StackDeploymentInfo records the information of a deployed stack
	StackDeploymentInfo struct {
		// Version is the version of the stack and also is the deployed version in edge agent
		Version int `json:"Version"`
		// FileVersion is the version of the stack file, used to detect changes
		FileVersion int `json:"FileVersion"`
		// ConfigHash is the commit hash of the git repository used for deploying the stack
		ConfigHash string `json:"ConfigHash,omitempty"`
	}

	// EdgeStack represents an edge stack
	EdgeStack struct {
		// EdgeStack Identifier
		ID     EdgeStackID                    `json:"Id" example:"1"`
		Name   string                         `json:"Name"`
		Status map[EndpointID]EdgeStackStatus `json:"Status"`
		// StatusArray    map[EndpointID][]EdgeStackStatus `json:"StatusArray"`
		CreationDate   int64                   `json:"CreationDate"`
		EdgeGroups     []EdgeGroupID           `json:"EdgeGroups"`
		ProjectPath    string                  `json:"ProjectPath"`
		EntryPoint     string                  `json:"EntryPoint"`
		Version        int                     `json:"Version"`
		NumDeployments int                     `json:"NumDeployments"`
		ManifestPath   string                  `json:"ManifestPath"`
		DeploymentType EdgeStackDeploymentType `json:"DeploymentType"`
		// Uses the manifest's namespaces instead of the default one
		UseManifestNamespaces bool
	}

	EdgeStackStatusForEnv struct {
		EndpointID EndpointID
		Status     []EdgeStackDeploymentStatus
		// EE only feature
		DeploymentInfo StackDeploymentInfo
		// ReadyRePullImage is a flag to indicate whether the auto update is trigger to re-pull image
		ReadyRePullImage bool `json:"ReadyRePullImage,omitempty"`
	}

	EdgeStackDeploymentType int

	// EdgeStackID represents an edge stack id
	EdgeStackID int

	EdgeStackStatusDetails struct {
		Pending             bool
		Ok                  bool
		Error               bool
		Acknowledged        bool
		Remove              bool
		RemoteUpdateSuccess bool
		ImagesPulled        bool
	}

	// EdgeStackStatus represents an edge stack status
	EdgeStackStatus struct {
		Status     []EdgeStackDeploymentStatus
		EndpointID EndpointID
		// EE only feature
		DeploymentInfo StackDeploymentInfo
		// ReadyRePullImage is a flag to indicate whether the auto update is trigger to re-pull image
		ReadyRePullImage bool `json:"ReadyRePullImage,omitempty"`

		// Deprecated
		Details *EdgeStackStatusDetails `json:"Details,omitempty"`
		// Deprecated
		Error string `json:"Error,omitempty"`
		// Deprecated
		Type EdgeStackStatusType `json:"Type,omitempty"`
	}

	// EdgeStackDeploymentStatus represents an edge stack deployment status
	EdgeStackDeploymentStatus struct {
		Time  int64
		Type  EdgeStackStatusType
		Error string `json:"Error,omitempty"`
		// EE only feature
		RollbackTo *int `json:"RollbackTo,omitempty"`
		Version    int  `json:"Version,omitempty"`
	}

	// EdgeStackStatusType represents an edge stack status type
	EdgeStackStatusType int

	// Environment(Endpoint) represents a Docker environment(endpoint) with all the info required
	// to connect to it
	Endpoint struct {
		// Environment(Endpoint) Identifier
		ID EndpointID `json:"Id" example:"1"`
		// Environment(Endpoint) name
		Name string `json:"Name" example:"my-environment"`
		// Environment(Endpoint) environment(endpoint) type. 1 for a Docker environment(endpoint), 2 for an agent on Docker environment(endpoint) or 3 for an Azure environment(endpoint).
		Type EndpointType `json:"Type" example:"1"`
		// ContainerEngine represents the container engine type. This can be 'docker' or 'podman' when interacting directly with these environmentes, otherwise '' for kubernetes environments.
		ContainerEngine string `json:"ContainerEngine" example:"docker"`
		// URL or IP address of the Docker host associated to this environment(endpoint)
		URL string `json:"URL" example:"docker.mydomain.tld:2375"`
		// Environment(Endpoint) group identifier
		GroupID EndpointGroupID `json:"GroupId" example:"1"`
		// URL or IP address where exposed containers will be reachable
		PublicURL        string           `json:"PublicURL" example:"docker.mydomain.tld:2375"`
		Gpus             []Pair           `json:"Gpus"`
		TLSConfig        TLSConfiguration `json:"TLSConfig"`
		AzureCredentials AzureCredentials `json:"AzureCredentials,omitempty"`
		// List of tag identifiers to which this environment(endpoint) is associated
		TagIDs []TagID `json:"TagIds"`
		// The status of the environment(endpoint) (1 - up, 2 - down)
		Status EndpointStatus `json:"Status" example:"1"`
		// List of snapshots
		Snapshots []DockerSnapshot `json:"Snapshots"`
		// List of user identifiers authorized to connect to this environment(endpoint)
		UserAccessPolicies UserAccessPolicies `json:"UserAccessPolicies"`
		// List of team identifiers authorized to connect to this environment(endpoint)
		TeamAccessPolicies TeamAccessPolicies `json:"TeamAccessPolicies"`
		// The identifier of the edge agent associated with this environment(endpoint)
		EdgeID string `json:"EdgeID,omitempty"`
		// The key which is used to map the agent to Portainer
		EdgeKey string `json:"EdgeKey"`
		// The check in interval for edge agent (in seconds)
		EdgeCheckinInterval int `json:"EdgeCheckinInterval" example:"5"`
		// Associated Kubernetes data
		Kubernetes KubernetesData `json:"Kubernetes"`
		// Maximum version of docker-compose
		ComposeSyntaxMaxVersion string `json:"ComposeSyntaxMaxVersion" example:"3.8"`
		// Environment(Endpoint) specific security settings
		SecuritySettings EndpointSecuritySettings
		// The identifier of the AMT Device associated with this environment(endpoint)
		AMTDeviceGUID string `json:"AMTDeviceGUID,omitempty" example:"4c4c4544-004b-3910-8037-b6c04f504633"`
		// LastCheckInDate mark last check-in date on checkin
		LastCheckInDate int64
		// QueryDate of each query with the endpoints list
		QueryDate int64
		// Heartbeat indicates the heartbeat status of an edge environment
		Heartbeat bool `json:"Heartbeat" example:"true"`

		// Whether the device has been trusted or not by the user
		UserTrusted bool `json:"UserTrusted,omitempty"`

		// Whether we need to run any "post init migrations".
		PostInitMigrations EndpointPostInitMigrations `json:"PostInitMigrations"`

		Edge EnvironmentEdgeSettings

		Agent struct {
			Version string `example:"1.0.0"`
		}

		EnableGPUManagement bool `json:"EnableGPUManagement,omitempty"`

		// Deprecated fields
		// Deprecated in DBVersion == 4
		TLS           bool   `json:"TLS,omitempty"`
		TLSCACertPath string `json:"TLSCACert,omitempty"`
		TLSCertPath   string `json:"TLSCert,omitempty"`
		TLSKeyPath    string `json:"TLSKey,omitempty"`

		// Deprecated in DBVersion == 18
		AuthorizedUsers []UserID `json:"AuthorizedUsers"`
		AuthorizedTeams []TeamID `json:"AuthorizedTeams"`

		// Deprecated in DBVersion == 22
		Tags []string `json:"Tags"`

		// Deprecated v2.18
		IsEdgeDevice bool `json:"IsEdgeDevice,omitempty"`
	}

	EnvironmentEdgeSettings struct {
		// Whether the device has been started in edge async mode
		AsyncMode bool
		// The ping interval for edge agent - used in edge async mode [seconds]
		PingInterval int `json:"PingInterval" example:"60"`
		// The snapshot interval for edge agent - used in edge async mode [seconds]
		SnapshotInterval int `json:"SnapshotInterval" example:"60"`
		// The command list interval for edge agent - used in edge async mode [seconds]
		CommandInterval int `json:"CommandInterval" example:"60"`
	}

	// EndpointAuthorizations represents the authorizations associated to a set of environments(endpoints)
	EndpointAuthorizations map[EndpointID]Authorizations

	// EndpointGroup represents a group of environments(endpoints).
	//
	// An environment(endpoint) may belong to only 1 environment(endpoint) group.
	EndpointGroup struct {
		// Environment(Endpoint) group Identifier
		ID EndpointGroupID `json:"Id" example:"1"`
		// Environment(Endpoint) group name
		Name string `json:"Name" example:"my-environment-group"`
		// Description associated to the environment(endpoint) group
		Description        string             `json:"Description" example:"Environment(Endpoint) group description"`
		UserAccessPolicies UserAccessPolicies `json:"UserAccessPolicies"`
		TeamAccessPolicies TeamAccessPolicies `json:"TeamAccessPolicies"`
		// List of tags associated to this environment(endpoint) group
		TagIDs []TagID `json:"TagIds"`

		// Deprecated fields
		Labels []Pair `json:"Labels"`

		// Deprecated in DBVersion == 18
		AuthorizedUsers []UserID `json:"AuthorizedUsers"`
		AuthorizedTeams []TeamID `json:"AuthorizedTeams"`

		// Deprecated in DBVersion == 22
		Tags []string `json:"Tags,omitempty"`
	}

	// EndpointGroupID represents an environment(endpoint) group identifier
	EndpointGroupID int

	// EndpointID represents an environment(endpoint) identifier
	EndpointID int

	// EndpointStatus represents the status of an environment(endpoint)
	EndpointStatus int

	// EndpointSyncJob represents a scheduled job that synchronize environments(endpoints) based on an external file
	// Deprecated
	EndpointSyncJob struct{}

	// EndpointSecuritySettings represents settings for an environment(endpoint)
	EndpointSecuritySettings struct {
		// Whether non-administrator should be able to use bind mounts when creating containers
		AllowBindMountsForRegularUsers bool `json:"allowBindMountsForRegularUsers" example:"false"`
		// Whether non-administrator should be able to use privileged mode when creating containers
		AllowPrivilegedModeForRegularUsers bool `json:"allowPrivilegedModeForRegularUsers" example:"false"`
		// Whether non-administrator should be able to browse volumes
		AllowVolumeBrowserForRegularUsers bool `json:"allowVolumeBrowserForRegularUsers" example:"true"`
		// Whether non-administrator should be able to use the host pid
		AllowHostNamespaceForRegularUsers bool `json:"allowHostNamespaceForRegularUsers" example:"true"`
		// Whether non-administrator should be able to use device mapping
		AllowDeviceMappingForRegularUsers bool `json:"allowDeviceMappingForRegularUsers" example:"true"`
		// Whether non-administrator should be able to manage stacks
		AllowStackManagementForRegularUsers bool `json:"allowStackManagementForRegularUsers" example:"true"`
		// Whether non-administrator should be able to use container capabilities
		AllowContainerCapabilitiesForRegularUsers bool `json:"allowContainerCapabilitiesForRegularUsers" example:"true"`
		// Whether non-administrator should be able to use sysctl settings
		AllowSysctlSettingForRegularUsers bool `json:"allowSysctlSettingForRegularUsers" example:"true"`
		// Whether host management features are enabled
		EnableHostManagementFeatures bool `json:"enableHostManagementFeatures" example:"true"`
	}

	// EndpointType represents the type of an environment(endpoint)
	EndpointType int

	// EndpointRelation represents a environment(endpoint) relation object
	EndpointRelation struct {
		EndpointID EndpointID
		EdgeStacks map[EdgeStackID]bool
	}

	// EndpointPostInitMigrations
	EndpointPostInitMigrations struct {
		MigrateIngresses bool `json:"MigrateIngresses"`
		MigrateGPUs      bool `json:"MigrateGPUs"`
	}

	// Extension represents a deprecated Portainer extension
	Extension struct {
		ID               ExtensionID                 `json:"Id" example:"1"`
		Enabled          bool                        `json:"Enabled"`
		Name             string                      `json:"Name,omitempty"`
		ShortDescription string                      `json:"ShortDescription,omitempty"`
		Description      string                      `json:"Description,omitempty"`
		DescriptionURL   string                      `json:"DescriptionURL,omitempty"`
		Price            string                      `json:"Price,omitempty"`
		PriceDescription string                      `json:"PriceDescription,omitempty"`
		Deal             bool                        `json:"Deal,omitempty"`
		Available        bool                        `json:"Available,omitempty"`
		License          ExtensionLicenseInformation `json:"License,omitempty"`
		Version          string                      `json:"Version"`
		UpdateAvailable  bool                        `json:"UpdateAvailable"`
		ShopURL          string                      `json:"ShopURL,omitempty"`
		Images           []string                    `json:"Images,omitempty"`
		Logo             string                      `json:"Logo,omitempty"`
	}

	// ExtensionID represents a extension identifier
	ExtensionID int

	// GitlabRegistryData represents data required for gitlab registry to work
	GitlabRegistryData struct {
		ProjectID   int    `json:"ProjectId"`
		InstanceURL string `json:"InstanceURL"`
		ProjectPath string `json:"ProjectPath"`
	}

	// GithubRegistryData represents data required for Github registry to work
	GithubRegistryData struct {
		UseOrganisation  bool   `json:"UseOrganisation"`
		OrganisationName string `json:"OrganisationName"`
	}

	HelmUserRepositoryID int

	// HelmUserRepositories stores a Helm repository URL for the given user
	HelmUserRepository struct {
		// Membership Identifier
		ID HelmUserRepositoryID `json:"Id" example:"1"`
		// User identifier
		UserID UserID `json:"UserId" example:"1"`
		// Helm repository URL
		URL string `json:"URL" example:"https://charts.bitnami.com/bitnami"`
	}

	// QuayRegistryData represents data required for Quay registry to work
	QuayRegistryData struct {
		UseOrganisation  bool   `json:"UseOrganisation,omitempty"`
		OrganisationName string `json:"OrganisationName"`
	}

	// EcrData represents data required for ECR registry
	EcrData struct {
		Region string `json:"Region" example:"ap-southeast-2"`
	}

	// JobType represents a job type
	JobType int

	K8sNamespaceInfo struct {
		Id                  string                 `json:"Id"`
		Name                string                 `json:"Name"`
		Status              corev1.NamespaceStatus `json:"Status"`
		Annotations         map[string]string      `json:"Annotations"`
		CreationDate        string                 `json:"CreationDate"`
		UnhealthyEventCount int                    `json:"UnhealthyEventCount"`
		NamespaceOwner      string                 `json:"NamespaceOwner"`
		IsSystem            bool                   `json:"IsSystem"`
		IsDefault           bool                   `json:"IsDefault"`
		ResourceQuota       *corev1.ResourceQuota  `json:"ResourceQuota"`
	}

	K8sNodeLimits struct {
		CPU    int64 `json:"CPU"`
		Memory int64 `json:"Memory"`
	}

	K8sNodesLimits map[string]*K8sNodeLimits

	K8sNamespaceAccessPolicy struct {
		UserAccessPolicies UserAccessPolicies `json:"UserAccessPolicies"`
		TeamAccessPolicies TeamAccessPolicies `json:"TeamAccessPolicies"`
	}

	// KubernetesData contains all the Kubernetes related environment(endpoint) information
	KubernetesData struct {
		Snapshots     []KubernetesSnapshot    `json:"Snapshots"`
		Configuration KubernetesConfiguration `json:"Configuration"`
		Flags         KubernetesFlags         `json:"Flags"`
	}

	// KubernetesFlags are used to detect if we need to run initial cluster
	// detection again.
	KubernetesFlags struct {
		IsServerMetricsDetected      bool `json:"IsServerMetricsDetected"`
		IsServerIngressClassDetected bool `json:"IsServerIngressClassDetected"`
		IsServerStorageDetected      bool `json:"IsServerStorageDetected"`
	}

	// KubernetesSnapshot represents a snapshot of a specific Kubernetes environment(endpoint) at a specific time
	KubernetesSnapshot struct {
		Time               int64               `json:"Time"`
		KubernetesVersion  string              `json:"KubernetesVersion"`
		NodeCount          int                 `json:"NodeCount"`
		TotalCPU           int64               `json:"TotalCPU"`
		TotalMemory        int64               `json:"TotalMemory"`
		DiagnosticsData    *DiagnosticsData    `json:"DiagnosticsData"`
		PerformanceMetrics *PerformanceMetrics `json:"PerformanceMetrics"`
	}

	// KubernetesConfiguration represents the configuration of a Kubernetes environment(endpoint)
	KubernetesConfiguration struct {
		UseLoadBalancer                 bool                           `json:"UseLoadBalancer"`
		UseServerMetrics                bool                           `json:"UseServerMetrics"`
		EnableResourceOverCommit        bool                           `json:"EnableResourceOverCommit"`
		ResourceOverCommitPercentage    int                            `json:"ResourceOverCommitPercentage"`
		StorageClasses                  []KubernetesStorageClassConfig `json:"StorageClasses"`
		IngressClasses                  []KubernetesIngressClassConfig `json:"IngressClasses"`
		RestrictDefaultNamespace        bool                           `json:"RestrictDefaultNamespace"`
		IngressAvailabilityPerNamespace bool                           `json:"IngressAvailabilityPerNamespace"`
		AllowNoneIngressClass           bool                           `json:"AllowNoneIngressClass"`
	}

	// KubernetesStorageClassConfig represents a Kubernetes Storage Class configuration
	KubernetesStorageClassConfig struct {
		Name                 string   `json:"Name"`
		AccessModes          []string `json:"AccessModes"`
		Provisioner          string   `json:"Provisioner"`
		AllowVolumeExpansion bool     `json:"AllowVolumeExpansion"`
	}

	// KubernetesIngressClassConfig represents a Kubernetes Ingress Class configuration
	KubernetesIngressClassConfig struct {
		Name              string   `json:"Name"`
		Type              string   `json:"Type"`
		GloballyBlocked   bool     `json:"Blocked"`
		BlockedNamespaces []string `json:"BlockedNamespaces"`
	}

	// KubernetesShellPod represents a Kubectl Shell details to facilitate pod exec functionality
	KubernetesShellPod struct {
		Namespace        string
		PodName          string
		ContainerName    string
		ShellExecCommand string
	}

	// InternalAuthSettings represents settings used for the default 'internal' authentication
	InternalAuthSettings struct {
		RequiredPasswordLength int
	}

	// LDAPGroupSearchSettings represents settings used to search for groups in a LDAP server
	LDAPGroupSearchSettings struct {
		// The distinguished name of the element from which the LDAP server will search for groups
		GroupBaseDN string `json:"GroupBaseDN" example:"dc=ldap,dc=domain,dc=tld"`
		// The LDAP search filter used to select group elements, optional
		GroupFilter string `json:"GroupFilter" example:"(objectClass=account"`
		// LDAP attribute which denotes the group membership
		GroupAttribute string `json:"GroupAttribute" example:"member"`
	}

	// LDAPSearchSettings represents settings used to search for users in a LDAP server
	LDAPSearchSettings struct {
		// The distinguished name of the element from which the LDAP server will search for users
		BaseDN string `json:"BaseDN" example:"dc=ldap,dc=domain,dc=tld"`
		// Optional LDAP search filter used to select user elements
		Filter string `json:"Filter" example:"(objectClass=account)"`
		// LDAP attribute which denotes the username
		UserNameAttribute string `json:"UserNameAttribute" example:"uid"`
	}

	// LDAPSettings represents the settings used to connect to a LDAP server
	LDAPSettings struct {
		// Enable this option if the server is configured for Anonymous access. When enabled, ReaderDN and Password will not be used
		AnonymousMode bool `json:"AnonymousMode" example:"true" validate:"validate_bool"`
		// Account that will be used to search for users
		ReaderDN string `json:"ReaderDN" example:"cn=readonly-account,dc=ldap,dc=domain,dc=tld" validate:"required_if=AnonymousMode false"`
		// Password of the account that will be used to search users
		Password string `json:"Password,omitempty" example:"readonly-password" validate:"required_if=AnonymousMode false"`
		// URL or IP address of the LDAP server
		URL       string           `json:"URL" example:"myldap.domain.tld:389" validate:"hostname_port"`
		TLSConfig TLSConfiguration `json:"TLSConfig"`
		// Whether LDAP connection should use StartTLS
		StartTLS            bool                      `json:"StartTLS" example:"true"`
		SearchSettings      []LDAPSearchSettings      `json:"SearchSettings"`
		GroupSearchSettings []LDAPGroupSearchSettings `json:"GroupSearchSettings"`
		// Automatically provision users and assign them to matching LDAP group names
		AutoCreateUsers bool `json:"AutoCreateUsers" example:"true"`
	}

	// LDAPUser represents a LDAP user
	LDAPUser struct {
		Name   string
		Groups []string
	}

	// ExtensionLicenseInformation represents information about an extension license
	ExtensionLicenseInformation struct {
		LicenseKey string `json:"LicenseKey,omitempty"`
		Company    string `json:"Company,omitempty"`
		Expiration string `json:"Expiration,omitempty"`
		Valid      bool   `json:"Valid,omitempty"`
	}

	// MembershipRole represents the role of a user within a team
	MembershipRole int

	// OAuthSettings represents the settings used to authorize with an authorization server
	OAuthSettings struct {
		ClientID             string           `json:"ClientID"`
		ClientSecret         string           `json:"ClientSecret,omitempty"`
		AccessTokenURI       string           `json:"AccessTokenURI"`
		AuthorizationURI     string           `json:"AuthorizationURI"`
		ResourceURI          string           `json:"ResourceURI"`
		RedirectURI          string           `json:"RedirectURI"`
		UserIdentifier       string           `json:"UserIdentifier"`
		Scopes               string           `json:"Scopes"`
		OAuthAutoCreateUsers bool             `json:"OAuthAutoCreateUsers"`
		DefaultTeamID        TeamID           `json:"DefaultTeamID"`
		SSO                  bool             `json:"SSO"`
		LogoutURI            string           `json:"LogoutURI"`
		KubeSecretKey        []byte           `json:"KubeSecretKey"`
		AuthStyle            oauth2.AuthStyle `json:"AuthStyle"`
	}

	// Pair defines a key/value string pair
	Pair struct {
		Name  string `json:"name" example:"name"`
		Value string `json:"value" example:"value"`
	}

	// Registry represents a Docker registry with all the info required
	// to connect to it
	Registry struct {
		// Registry Identifier
		ID RegistryID `json:"Id" example:"1"`
		// Registry Type (1 - Quay, 2 - Azure, 3 - Custom, 4 - Gitlab, 5 - ProGet, 6 - DockerHub, 7 - ECR)
		Type RegistryType `json:"Type" enums:"1,2,3,4,5,6,7"`
		// Registry Name
		Name string `json:"Name" example:"my-registry"`
		// URL or IP address of the Docker registry
		URL string `json:"URL" example:"registry.mydomain.tld:2375"`
		// Base URL, introduced for ProGet registry
		BaseURL string `json:"BaseURL" example:"registry.mydomain.tld:2375"`
		// Is authentication against this registry enabled
		Authentication bool `json:"Authentication" example:"true"`
		// Username or AccessKeyID used to authenticate against this registry
		Username string `json:"Username" example:"registry user"`
		// Password or SecretAccessKey used to authenticate against this registry
		Password                string                           `json:"Password,omitempty" example:"registry_password"`
		ManagementConfiguration *RegistryManagementConfiguration `json:"ManagementConfiguration"`
		Gitlab                  GitlabRegistryData               `json:"Gitlab"`
		Github                  GithubRegistryData               `json:"Github"`
		Quay                    QuayRegistryData                 `json:"Quay"`
		Ecr                     EcrData                          `json:"Ecr"`
		RegistryAccesses        RegistryAccesses                 `json:"RegistryAccesses"`

		// Deprecated fields
		// Deprecated in DBVersion == 31
		UserAccessPolicies UserAccessPolicies `json:"UserAccessPolicies"`
		// Deprecated in DBVersion == 31
		TeamAccessPolicies TeamAccessPolicies `json:"TeamAccessPolicies"`

		// Deprecated in DBVersion == 18
		AuthorizedUsers []UserID `json:"AuthorizedUsers"`
		// Deprecated in DBVersion == 18
		AuthorizedTeams []TeamID `json:"AuthorizedTeams"`

		// Stores temporary access token
		AccessToken       string `json:"AccessToken,omitempty"`
		AccessTokenExpiry int64  `json:"AccessTokenExpiry,omitempty"`
	}

	RegistryAccesses map[EndpointID]RegistryAccessPolicies

	RegistryAccessPolicies struct {
		UserAccessPolicies UserAccessPolicies `json:"UserAccessPolicies"`
		TeamAccessPolicies TeamAccessPolicies `json:"TeamAccessPolicies"`
		Namespaces         []string           `json:"Namespaces"`
	}

	// RegistryID represents a registry identifier
	RegistryID int

	// RegistryManagementConfiguration represents a configuration that can be used to query
	// the registry API via the registry management extension.
	RegistryManagementConfiguration struct {
		Type              RegistryType     `json:"Type"`
		Authentication    bool             `json:"Authentication"`
		Username          string           `json:"Username"`
		Password          string           `json:"Password"`
		TLSConfig         TLSConfiguration `json:"TLSConfig"`
		Ecr               EcrData          `json:"Ecr"`
		AccessToken       string           `json:"AccessToken,omitempty"`
		AccessTokenExpiry int64            `json:"AccessTokenExpiry,omitempty"`
	}

	// RegistryType represents a type of registry
	RegistryType int

	// ResourceAccessLevel represents the level of control associated to a resource
	ResourceAccessLevel int

	// ResourceControl represent a reference to a Docker resource with specific access controls
	ResourceControl struct {
		// ResourceControl Identifier
		ID ResourceControlID `json:"Id" example:"1"`
		// Docker resource identifier on which access control will be applied.\
		// In the case of a resource control applied to a stack, use the stack name as identifier
		ResourceID string `json:"ResourceId" example:"617c5f22bb9b023d6daab7cba43a57576f83492867bc767d1c59416b065e5f08"`
		// List of Docker resources that will inherit this access control
		SubResourceIDs []string `json:"SubResourceIds" example:"617c5f22bb9b023d6daab7cba43a57576f83492867bc767d1c59416b065e5f08"`
		// Type of Docker resource. Valid values are: 1- container, 2 -service
		// 3 - volume, 4 - secret, 5 - stack, 6 - config or 7 - custom template
		Type         ResourceControlType  `json:"Type" example:"1"`
		UserAccesses []UserResourceAccess `json:"UserAccesses"`
		TeamAccesses []TeamResourceAccess `json:"TeamAccesses"`
		// Permit access to the associated resource to any user
		Public bool `json:"Public" example:"true"`
		// Permit access to resource only to admins
		AdministratorsOnly bool `json:"AdministratorsOnly" example:"true"`
		System             bool `json:"System"`

		// Deprecated fields
		// Deprecated in DBVersion == 2
		OwnerID     UserID              `json:"OwnerId,omitempty"`
		AccessLevel ResourceAccessLevel `json:"AccessLevel,omitempty"`
	}

	// ResourceControlID represents a resource control identifier
	ResourceControlID int

	// ResourceControlType represents the type of resource associated to the resource control (volume, container, service...)
	ResourceControlType int

	// Role represents a set of authorizations that can be associated to a user or
	// to a team.
	Role struct {
		// Role Identifier
		ID RoleID `json:"Id" example:"1"`
		// Role name
		Name string `json:"Name" example:"HelpDesk"`
		// Role description
		Description string `json:"Description" example:"Read-only access of all resources in an environment(endpoint)"`
		// Authorizations associated to a role
		Authorizations Authorizations `json:"Authorizations"`
		Priority       int            `json:"Priority"`
	}

	// RoleID represents a role identifier
	RoleID int

	// APIKeyID represents an API key identifier
	APIKeyID int

	// APIKey represents an API key
	APIKey struct {
		ID          APIKeyID `json:"id" example:"1"`
		UserID      UserID   `json:"userId" example:"1"`
		Description string   `json:"description" example:"portainer-api-key"`
		Prefix      string   `json:"prefix"`           // API key identifier (7 char prefix)
		DateCreated int64    `json:"dateCreated"`      // Unix timestamp (UTC) when the API key was created
		LastUsed    int64    `json:"lastUsed"`         // Unix timestamp (UTC) when the API key was last used
		Digest      string   `json:"digest,omitempty"` // Digest represents SHA256 hash of the raw API key
	}

	// Schedule represents a scheduled job.
	// It only contains a pointer to one of the JobRunner implementations
	// based on the JobType.
	// NOTE: The Recurring option is only used by ScriptExecutionJob at the moment
	// Deprecated in favor of EdgeJob
	Schedule struct {
		// Schedule Identifier
		ID             ScheduleID `json:"Id" example:"1"`
		Name           string
		CronExpression string
		Recurring      bool
		Created        int64
		JobType        JobType
		EdgeSchedule   *EdgeSchedule
	}

	// ScheduleID represents a schedule identifier.
	// Deprecated in favor of EdgeJob
	ScheduleID int

	// ScriptExecutionJob represents a scheduled job that can execute a script via a privileged container
	ScriptExecutionJob struct {
		Endpoints     []EndpointID
		Image         string
		ScriptPath    string
		RetryCount    int
		RetryInterval int
	}

	GlobalDeploymentOptions struct {
		HideStacksFunctionality bool `json:"hideStacksFunctionality" example:"false"`
	}

	Edge struct {
		// The command list interval for edge agent - used in edge async mode (in seconds)
		CommandInterval int `json:"CommandInterval" example:"5"`
		// The ping interval for edge agent - used in edge async mode (in seconds)
		PingInterval int `json:"PingInterval" example:"5"`
		// The snapshot interval for edge agent - used in edge async mode (in seconds)
		SnapshotInterval int `json:"SnapshotInterval" example:"5"`

		// Deprecated 2.18
		AsyncMode bool `json:"AsyncMode,omitempty" example:"false"`
	}

	// Settings represents the application settings
	Settings struct {
		// URL to a logo that will be displayed on the login page as well as on top of the sidebar. Will use default Portainer logo when value is empty string
		LogoURL string `json:"LogoURL" example:"https://mycompany.mydomain.tld/logo.png"`
		// A list of label name & value that will be used to hide containers when querying containers
		BlackListedLabels []Pair `json:"BlackListedLabels"`
		// Active authentication method for the Portainer instance. Valid values are: 1 for internal, 2 for LDAP, or 3 for oauth
		AuthenticationMethod AuthenticationMethod          `json:"AuthenticationMethod" example:"1"`
		InternalAuthSettings InternalAuthSettings          `json:"InternalAuthSettings"`
		LDAPSettings         LDAPSettings                  `json:"LDAPSettings"`
		OAuthSettings        OAuthSettings                 `json:"OAuthSettings"`
		OpenAMTConfiguration OpenAMTConfiguration          `json:"openAMTConfiguration"`
		FeatureFlagSettings  map[featureflags.Feature]bool `json:"FeatureFlagSettings"`
		// The interval in which environment(endpoint) snapshots are created
		SnapshotInterval string `json:"SnapshotInterval" example:"5m"`
		// URL to the templates that will be displayed in the UI when navigating to App Templates
		TemplatesURL string `json:"TemplatesURL" example:"https://raw.githubusercontent.com/portainer/templates/master/templates.json"`
		// Deployment options for encouraging git ops workflows
		GlobalDeploymentOptions GlobalDeploymentOptions `json:"GlobalDeploymentOptions"`
		// The default check in interval for edge agent (in seconds)
		EdgeAgentCheckinInterval int `json:"EdgeAgentCheckinInterval" example:"5"`
		// Whether edge compute features are enabled
		EnableEdgeComputeFeatures bool `json:"EnableEdgeComputeFeatures"`
		// The duration of a user session
		UserSessionTimeout string `json:"UserSessionTimeout" example:"5m"`
		// The expiry of a Kubeconfig
		KubeconfigExpiry string `json:"KubeconfigExpiry" example:"24h"`
		// Whether telemetry is enabled
		EnableTelemetry bool `json:"EnableTelemetry" example:"false"`
		// Helm repository URL, defaults to "https://charts.bitnami.com/bitnami"
		HelmRepositoryURL string `json:"HelmRepositoryURL" example:"https://charts.bitnami.com/bitnami"`
		// KubectlImage, defaults to portainer/kubectl-shell
		KubectlShellImage string `json:"KubectlShellImage" example:"portainer/kubectl-shell"`
		// TrustOnFirstConnect makes Portainer accepting edge agent connection by default
		TrustOnFirstConnect bool `json:"TrustOnFirstConnect" example:"false"`
		// EnforceEdgeID makes Portainer store the Edge ID instead of accepting anyone
		EnforceEdgeID bool `json:"EnforceEdgeID" example:"false"`
		// Container environment parameter AGENT_SECRET
		AgentSecret string `json:"AgentSecret"`
		// EdgePortainerURL is the URL that is exposed to edge agents
		EdgePortainerURL string `json:"EdgePortainerUrl"`

		Edge Edge `json:"Edge"`

		// Deprecated fields
		DisplayDonationHeader       bool `json:"DisplayDonationHeader,omitempty"`
		DisplayExternalContributors bool `json:"DisplayExternalContributors,omitempty"`

		// Deprecated fields v26
		EnableHostManagementFeatures              bool `json:"EnableHostManagementFeatures,omitempty"`
		AllowVolumeBrowserForRegularUsers         bool `json:"AllowVolumeBrowserForRegularUsers,omitempty"`
		AllowBindMountsForRegularUsers            bool `json:"AllowBindMountsForRegularUsers,omitempty"`
		AllowPrivilegedModeForRegularUsers        bool `json:"AllowPrivilegedModeForRegularUsers,omitempty"`
		AllowHostNamespaceForRegularUsers         bool `json:"AllowHostNamespaceForRegularUsers,omitempty"`
		AllowStackManagementForRegularUsers       bool `json:"AllowStackManagementForRegularUsers,omitempty"`
		AllowDeviceMappingForRegularUsers         bool `json:"AllowDeviceMappingForRegularUsers,omitempty"`
		AllowContainerCapabilitiesForRegularUsers bool `json:"AllowContainerCapabilitiesForRegularUsers,omitempty"`

		IsDockerDesktopExtension bool `json:"IsDockerDesktopExtension,omitempty"`
	}

	// SnapshotJob represents a scheduled job that can create environment(endpoint) snapshots
	SnapshotJob struct{}

	// SoftwareEdition represents an edition of Portainer
	SoftwareEdition int

	// SSLSettings represents a pair of SSL certificate and key
	SSLSettings struct {
		CertPath    string `json:"certPath"`
		KeyPath     string `json:"keyPath"`
		SelfSigned  bool   `json:"selfSigned"`
		HTTPEnabled bool   `json:"httpEnabled"`
	}

	// Stack represents a Docker stack created via docker stack deploy
	Stack struct {
		// Stack Identifier
		ID StackID `json:"Id" example:"1"`
		// Stack name
		Name string `json:"Name" example:"myStack"`
		// Stack type. 1 for a Swarm stack, 2 for a Compose stack
		Type StackType `json:"Type" example:"2"`
		// Environment(Endpoint) identifier. Reference the environment(endpoint) that will be used for deployment
		EndpointID EndpointID `json:"EndpointId" example:"1"`
		// Cluster identifier of the Swarm cluster where the stack is deployed
		SwarmID string `json:"SwarmId" example:"jpofkc0i9uo9wtx1zesuk649w"`
		// Path to the Stack file
		EntryPoint string `json:"EntryPoint" example:"docker-compose.yml"`
		// A list of environment(endpoint) variables used during stack deployment
		Env []Pair `json:"Env"`
		//
		ResourceControl *ResourceControl `json:"ResourceControl"`
		// Stack status (1 - active, 2 - inactive)
		Status StackStatus `json:"Status" example:"1"`
		// Path on disk to the repository hosting the Stack file
		ProjectPath string `example:"/data/compose/myStack_jpofkc0i9uo9wtx1zesuk649w"`
		// The date in unix time when stack was created
		CreationDate int64 `example:"1587399600"`
		// The username which created this stack
		CreatedBy string `example:"admin"`
		// The date in unix time when stack was last updated
		UpdateDate int64 `example:"1587399600"`
		// The username which last updated this stack
		UpdatedBy string `example:"bob"`
		// Only applies when deploying stack with multiple files
		AdditionalFiles []string `json:"AdditionalFiles"`
		// The GitOps update settings of a git stack
		AutoUpdate *AutoUpdateSettings `json:"AutoUpdate"`
		// The stack deployment option
		Option *StackOption `json:"Option"`
		// The git config of this stack
		GitConfig *gittypes.RepoConfig
		// Whether the stack is from a app template
		FromAppTemplate bool `example:"false"`
		// Kubernetes namespace if stack is a kube application
		Namespace string `example:"default"`
	}

	// StackOption represents the options for stack deployment
	StackOption struct {
		// Prune services that are no longer referenced
		Prune bool `example:"false"`
	}

	// StackID represents a stack identifier (it must be composed of Name + "_" + SwarmID to create a unique identifier)
	StackID int

	// StackStatus represent a status for a stack
	StackStatus int

	// StackType represents the type of the stack (compose v2, stack deploy v3)
	StackType int

	// Status represents the application status
	Status struct {
		// Portainer API version
		Version string `json:"Version" example:"2.0.0"`
		// Server Instance ID
		InstanceID string `example:"299ab403-70a8-4c05-92f7-bf7a994d50df"`
	}

	// Tag represents a tag that can be associated to a resource
	Tag struct {
		// Tag identifier
		ID TagID `example:"1"`
		// Tag name
		Name string `json:"Name" example:"org/acme"`
		// A set of environment(endpoint) ids that have this tag
		Endpoints map[EndpointID]bool `json:"Endpoints"`
		// A set of environment(endpoint) group ids that have this tag
		EndpointGroups map[EndpointGroupID]bool `json:"EndpointGroups"`
	}

	// TagID represents a tag identifier
	TagID int

	// Team represents a list of user accounts
	Team struct {
		// Team Identifier
		ID TeamID `json:"Id" example:"1"`
		// Team name
		Name string `json:"Name" example:"developers"`
	}

	// TeamAccessPolicies represent the association of an access policy and a team
	TeamAccessPolicies map[TeamID]AccessPolicy

	// TeamID represents a team identifier
	TeamID int

	// TeamMembership represents a membership association between a user and a team.
	//
	// A user may belong to multiple teams.
	TeamMembership struct {
		// Membership Identifier
		ID TeamMembershipID `json:"Id" example:"1"`
		// User identifier
		UserID UserID `json:"UserID" example:"1"`
		// Team identifier
		TeamID TeamID `json:"TeamID" example:"1"`
		// Team role (1 for team leader and 2 for team member)
		Role MembershipRole `json:"Role" example:"1"`
	}

	// TeamMembershipID represents a team membership identifier
	TeamMembershipID int

	// TeamResourceAccess represents the level of control on a resource for a specific team
	TeamResourceAccess struct {
		TeamID      TeamID              `json:"TeamId"`
		AccessLevel ResourceAccessLevel `json:"AccessLevel"`
	}

	// Template represents an application template that can be used as an App Template
	// or an Edge template
	Template struct {
		// Mandatory container/stack fields
		// Template Identifier
		ID TemplateID `json:"id" example:"1"`
		// Template type. Valid values are: 1 (container), 2 (Swarm stack), 3 (Compose stack), 4 (Compose edge stack)
		Type TemplateType `json:"type" example:"1"`
		// Title of the template
		Title string `json:"title" example:"Nginx"`
		// Description of the template
		Description string `json:"description" example:"High performance web server"`
		// Whether the template should be available to administrators only
		AdministratorOnly bool `json:"administrator_only" example:"true"`

		// Mandatory container fields
		// Image associated to a container template. Mandatory for a container template
		Image string `json:"image" example:"nginx:latest"`

		// Mandatory stack fields
		Repository TemplateRepository `json:"repository"`

		// Mandatory Edge stack fields
		// Stack file used for this template
		StackFile string `json:"stackFile"`

		// Optional stack/container fields
		// Default name for the stack/container to be used on deployment
		Name string `json:"name,omitempty" example:"mystackname"`
		// URL of the template's logo
		Logo string `json:"logo,omitempty" example:"https://portainer.io/img/logo.svg"`
		// A list of environment(endpoint) variables used during the template deployment
		Env []TemplateEnv `json:"env,omitempty"`
		// A note that will be displayed in the UI. Supports HTML content
		Note string `json:"note,omitempty" example:"This is my <b>custom</b> template"`
		// Platform associated to the template.
		// Valid values are: 'linux', 'windows' or leave empty for multi-platform
		Platform string `json:"platform,omitempty" example:"linux"`
		// A list of categories associated to the template
		Categories []string `json:"categories,omitempty" example:"database"`

		// Optional container fields
		// The URL of a registry associated to the image for a container template
		Registry string `json:"registry,omitempty" example:"quay.io"`
		// The command that will be executed in a container template
		Command string `json:"command,omitempty" example:"ls -lah"`
		// Name of a network that will be used on container deployment if it exists inside the environment(endpoint)
		Network string `json:"network,omitempty" example:"mynet"`
		// A list of volumes used during the container template deployment
		Volumes []TemplateVolume `json:"volumes,omitempty"`
		// A list of ports exposed by the container
		Ports []string `json:"ports,omitempty" example:"8080:80/tcp"`
		// Container labels
		Labels []Pair `json:"labels,omitempty"`
		// Whether the container should be started in privileged mode
		Privileged bool `json:"privileged,omitempty" example:"true"`
		// Whether the container should be started in
		// interactive mode (-i -t equivalent on the CLI)
		Interactive bool `json:"interactive,omitempty" example:"true"`
		// Container restart policy
		RestartPolicy string `json:"restart_policy,omitempty" example:"on-failure"`
		// Container hostname
		Hostname string `json:"hostname,omitempty" example:"mycontainer"`
	}

	// TemplateEnv represents a template environment(endpoint) variable configuration
	TemplateEnv struct {
		// name of the environment(endpoint) variable
		Name string `json:"name" example:"MYSQL_ROOT_PASSWORD"`
		// Text for the label that will be generated in the UI
		Label string `json:"label,omitempty" example:"Root password"`
		// Content of the tooltip that will be generated in the UI
		Description string `json:"description,omitempty" example:"MySQL root account password"`
		// Default value that will be set for the variable
		Default string `json:"default,omitempty" example:"default_value"`
		// If set to true, will not generate any input for this variable in the UI
		Preset bool `json:"preset,omitempty" example:"false"`
		// A list of name/value that will be used to generate a dropdown in the UI
		Select []TemplateEnvSelect `json:"select,omitempty"`
	}

	// TemplateEnvSelect represents text/value pair that will be displayed as a choice for the
	// template user
	TemplateEnvSelect struct {
		// Some text that will displayed as a choice
		Text string `json:"text" example:"text value"`
		// A value that will be associated to the choice
		Value string `json:"value" example:"value"`
		// Will set this choice as the default choice
		Default bool `json:"default" example:"false"`
	}

	// TemplateID represents a template identifier
	TemplateID int

	// TemplateRepository represents the git repository configuration for a template
	TemplateRepository struct {
		// URL of a git repository used to deploy a stack template. Mandatory for a Swarm/Compose stack template
		URL string `json:"url" example:"https://github.com/portainer/portainer-compose"`
		// Path to the stack file inside the git repository
		StackFile string `json:"stackfile" example:"./subfolder/docker-compose.yml"`
	}

	// TemplateType represents the type of a template
	TemplateType int

	// TemplateVolume represents a template volume configuration
	TemplateVolume struct {
		// Path inside the container
		Container string `json:"container" example:"/data"`
		// Path on the host
		Bind string `json:"bind,omitempty" example:"/tmp"`
		// Whether the volume used should be readonly
		ReadOnly bool `json:"readonly,omitempty" example:"true"`
	}

	// TLSConfiguration represents a TLS configuration
	TLSConfiguration struct {
		// Use TLS
		TLS bool `json:"TLS" example:"true"`
		// Skip the verification of the server TLS certificate
		TLSSkipVerify bool `json:"TLSSkipVerify" example:"false"`
		// Path to the TLS CA certificate file
		TLSCACertPath string `json:"TLSCACert,omitempty" example:"/data/tls/ca.pem"`
		// Path to the TLS client certificate file
		TLSCertPath string `json:"TLSCert,omitempty" example:"/data/tls/cert.pem"`
		// Path to the TLS client key file
		TLSKeyPath string `json:"TLSKey,omitempty" example:"/data/tls/key.pem"`
	}

	// TLSFileType represents a type of TLS file required to connect to a Docker environment(endpoint).
	// It can be either a TLS CA file, a TLS certificate file or a TLS key file
	TLSFileType int

	// TokenData represents the data embedded in a JWT token
	TokenData struct {
		ID                  UserID
		Username            string
		Role                UserRole
		ForceChangePassword bool
		Token               string
	}

	// TunnelDetails represents information associated to a tunnel
	TunnelDetails struct {
		Status       string
		LastActivity time.Time
		Port         int
		Credentials  string
	}

	// TunnelServerInfo represents information associated to the tunnel server
	TunnelServerInfo struct {
		PrivateKeySeed string `json:"PrivateKeySeed"`
	}

	// User represents a user account
	User struct {
		// User Identifier
		ID       UserID `json:"Id" example:"1"`
		Username string `json:"Username" example:"bob"`
		Password string `json:"Password,omitempty" swaggerignore:"true"`
		// User role (1 for administrator account and 2 for regular account)
		Role          UserRole          `json:"Role" example:"1"`
		TokenIssueAt  int64             `json:"TokenIssueAt" example:"1"`
		ThemeSettings UserThemeSettings `json:"ThemeSettings"`
		UseCache      bool              `json:"UseCache" example:"true"`

		// Deprecated fields

		// Deprecated
		UserTheme string `json:"UserTheme,omitempty" example:"dark"`
		// Deprecated in DBVersion == 25
		PortainerAuthorizations Authorizations
		// Deprecated in DBVersion == 25
		EndpointAuthorizations EndpointAuthorizations
	}

	// UserAccessPolicies represent the association of an access policy and a user
	UserAccessPolicies map[UserID]AccessPolicy

	// UserID represents a user identifier
	UserID int

	// UserResourceAccess represents the level of control on a resource for a specific user
	UserResourceAccess struct {
		UserID      UserID              `json:"UserId"`
		AccessLevel ResourceAccessLevel `json:"AccessLevel"`
	}

	// UserRole represents the role of a user. It can be either an administrator
	// or a regular user
	UserRole int

	// UserThemeSettings represents the theme settings for a user
	UserThemeSettings struct {
		// Color represents the color theme of the UI
		Color string `json:"color" example:"dark" enums:"dark,light,highcontrast,auto"`
	}

	// Webhook represents a url webhook that can be used to update a service
	Webhook struct {
		// Webhook Identifier
		ID         WebhookID  `json:"Id" example:"1"`
		Token      string     `json:"Token"`
		ResourceID string     `json:"ResourceId"`
		EndpointID EndpointID `json:"EndpointId"`
		RegistryID RegistryID `json:"RegistryId"`
		// Type of webhook (1 - service)
		WebhookType WebhookType `json:"Type"`
	}

	// WebhookID represents a webhook identifier.
	WebhookID int

	// WebhookType represents the type of resource a webhook is related to
	WebhookType int

	Snapshot struct {
		EndpointID EndpointID          `json:"EndpointId"`
		Docker     *DockerSnapshot     `json:"Docker"`
		Kubernetes *KubernetesSnapshot `json:"Kubernetes"`
	}

	SnapshotRawMessage struct {
		EndpointID EndpointID      `json:"EndpointId"`
		Docker     json.RawMessage `json:"Docker"`
		Kubernetes json.RawMessage `json:"Kubernetes"`
	}

	// CLIService represents a service for managing CLI
	CLIService interface {
		ParseFlags(version string) (*CLIFlags, error)
		ValidateFlags(flags *CLIFlags) error
	}

	ComposeOptions struct {
		Registries []Registry
	}

	ComposeUpOptions struct {
		ComposeOptions

		// ForceRecreate forces to recreate containers
		ForceRecreate bool
		// AbortOnContainerExit will stop the deployment if a container exits.
		// This is useful when running a onetime task.
		//
		// When this is set, docker compose will output its logs to stdout
		AbortOnContainerExit bool
		Prune                bool
	}

	ComposeDownOptions struct {
		// RemoveVolumes will remove the named volumes declared in the compose file
		// and anonymous volumes attached to the stack's containers
		// Drives `docker compose down --volumes`
		RemoveVolumes bool
	}

	ComposeRunOptions struct {
		ComposeOptions

		// Remove will remove the container after it has stopped
		Remove bool
		// Args are the arguments to pass to the container
		Args []string
		// Detached will run the container in the background
		Detached bool
	}

	// ComposeStackManager represents a service to manage Compose stacks
	ComposeStackManager interface {
		ComposeSyntaxMaxVersion() string
		NormalizeStackName(name string) string
		Run(ctx context.Context, stack *Stack, endpoint *Endpoint, serviceName string, options ComposeRunOptions) error
		Up(ctx context.Context, stack *Stack, endpoint *Endpoint, options ComposeUpOptions) error
		Down(ctx context.Context, stack *Stack, endpoint *Endpoint) error
		Pull(ctx context.Context, stack *Stack, endpoint *Endpoint, options ComposeOptions) error
	}

	// CryptoService represents a service for encrypting/hashing data
	CryptoService interface {
		Hash(data string) (string, error)
		CompareHashAndData(hash string, data string) error
	}

	// DigitalSignatureService represents a service to manage digital signatures
	DigitalSignatureService interface {
		ParseKeyPair(private, public []byte) error
		GenerateKeyPair() ([]byte, []byte, error)
		EncodedPublicKey() string
		PEMHeaders() (string, string)
		CreateSignature(message string) (string, error)
	}

	// DockerSnapshotter represents a service used to create Docker environment(endpoint) snapshots
	DockerSnapshotter interface {
		CreateSnapshot(endpoint *Endpoint) (*DockerSnapshot, error)
	}

	// FileService represents a service for managing files
	FileService interface {
		GetDockerConfigPath() string
		GetFileContent(trustedRootPath, filePath string) ([]byte, error)
		Copy(fromFilePath string, toFilePath string, deleteIfExists bool) error
		Rename(oldPath, newPath string) error
		RemoveDirectory(directoryPath string) error
		StoreTLSFileFromBytes(folder string, fileType TLSFileType, data []byte) (string, error)
		GetPathForTLSFile(folder string, fileType TLSFileType) (string, error)
		DeleteTLSFile(folder string, fileType TLSFileType) error
		DeleteTLSFiles(folder string) error
		GetStackProjectPath(stackIdentifier string) string
		GetStackProjectPathByVersion(stackIdentifier string, version int, commitHash string) string
		StoreStackFileFromBytes(stackIdentifier, fileName string, data []byte) (string, error)
		StoreStackFileFromBytesByVersion(stackIdentifier, fileName string, version int, data []byte) (string, error)
		UpdateStoreStackFileFromBytes(stackIdentifier, fileName string, data []byte) (string, error)
		UpdateStoreStackFileFromBytesByVersion(stackIdentifier, fileName string, version int, commitHash string, data []byte) (string, error)
		RemoveStackFileBackup(stackIdentifier, fileName string) error
		RemoveStackFileBackupByVersion(stackIdentifier string, version int, fileName string) error
		RollbackStackFile(stackIdentifier, fileName string) error
		RollbackStackFileByVersion(stackIdentifier string, version int, fileName string) error
		GetEdgeStackProjectPath(edgeStackIdentifier string) string
		StoreEdgeStackFileFromBytes(edgeStackIdentifier, fileName string, data []byte) (string, error)
		GetEdgeStackProjectPathByVersion(edgeStackIdentifier string, version int, commitHash string) string
		StoreEdgeStackFileFromBytesByVersion(edgeStackIdentifier, fileName string, version int, data []byte) (string, error)
		FormProjectPathByVersion(projectPath string, version int, commitHash string) string
		SafeMoveDirectory(src, dst string) error
		StoreRegistryManagementFileFromBytes(folder, fileName string, data []byte) (string, error)
		KeyPairFilesExist() (bool, error)
		StoreKeyPair(private, public []byte, privatePEMHeader, publicPEMHeader string) error
		LoadKeyPair() ([]byte, []byte, error)
		WriteJSONToFile(path string, content any) error
		FileExists(path string) (bool, error)
		StoreEdgeJobFileFromBytes(identifier string, data []byte) (string, error)
		GetEdgeJobFolder(identifier string) string
		ClearEdgeJobTaskLogs(edgeJobID, taskID string) error
		GetEdgeJobTaskLogFileContent(edgeJobID, taskID string) (string, error)
		StoreEdgeJobTaskLogFileFromBytes(edgeJobID, taskID string, data []byte) error
		GetBinaryFolder() string
		StoreCustomTemplateFileFromBytes(identifier, fileName string, data []byte) (string, error)
		GetCustomTemplateProjectPath(identifier string) string
		GetTemporaryPath() (string, error)
		GetDatastorePath() string
		GetDefaultSSLCertsPath() (string, string)
		StoreSSLCertPair(cert, key []byte) (string, string, error)
		CopySSLCertPair(certPath, keyPath string) (string, string, error)
		CopySSLCACert(caCertPath string) (string, error)
		StoreMTLSCertificates(caCert, cert, key []byte) (string, string, string, error)
		GetMTLSCertificates() (string, string, string, error)
		GetDefaultChiselPrivateKeyPath() string
		StoreChiselPrivateKey(privateKey []byte) error
	}

	// GitService represents a service for managing Git
	GitService interface {
		CloneRepository(destination string, repositoryURL, referenceName, username, password string, tlsSkipVerify bool) error
		LatestCommitID(repositoryURL, referenceName, username, password string, tlsSkipVerify bool) (string, error)
		ListRefs(repositoryURL, username, password string, hardRefresh bool, tlsSkipVerify bool) ([]string, error)
		ListFiles(repositoryURL, referenceName, username, password string, dirOnly, hardRefresh bool, includeExts []string, tlsSkipVerify bool) ([]string, error)
	}

	// OpenAMTService represents a service for managing OpenAMT
	OpenAMTService interface {
		Configure(configuration OpenAMTConfiguration) error
		DeviceInformation(configuration OpenAMTConfiguration, deviceGUID string) (*OpenAMTDeviceInformation, error)
		EnableDeviceFeatures(configuration OpenAMTConfiguration, deviceGUID string, features OpenAMTDeviceEnabledFeatures) (string, error)
		ExecuteDeviceAction(configuration OpenAMTConfiguration, deviceGUID string, action string) error
	}

	// JWTService represents a service for managing JWT tokens
	JWTService interface {
		GenerateToken(data *TokenData) (string, time.Time, error)
		GenerateTokenForKubeconfig(data *TokenData) (string, error)
		ParseAndVerifyToken(token string) (*TokenData, string, time.Time, error)
		SetUserSessionDuration(userSessionDuration time.Duration)
	}

	// KubeClient represents a service used to query a Kubernetes environment(endpoint)
	KubeClient interface {
		// Access
		GetIsKubeAdmin() bool
		SetIsKubeAdmin(isKubeAdmin bool)
		GetClientNonAdminNamespaces() []string
		SetClientNonAdminNamespaces([]string)
		NamespaceAccessPoliciesDeleteNamespace(ns string) error
		UpdateNamespaceAccessPolicies(accessPolicies map[string]K8sNamespaceAccessPolicy) error
		GetNamespaceAccessPolicies() (map[string]K8sNamespaceAccessPolicy, error)
		GetNonAdminNamespaces(userID int, teamIDs []int, isRestrictDefaultNamespace bool) ([]string, error)

		// Applications
		GetApplications(namespace, nodeName string) ([]models.K8sApplication, error)
		GetApplicationsResource(namespace, node string) (models.K8sApplicationResource, error)

		// ClusterRole
		GetClusterRoles() ([]models.K8sClusterRole, error)
		DeleteClusterRoles(req models.K8sClusterRoleDeleteRequests) error

		// ConfigMap
		GetConfigMap(namespace, configMapName string) (models.K8sConfigMap, error)
		CombineConfigMapWithApplications(configMap models.K8sConfigMap) (models.K8sConfigMap, error)

		// CronJob
		GetCronJobs(namespace string) ([]models.K8sCronJob, error)
		DeleteCronJobs(payload models.K8sCronJobDeleteRequests) error

		// Event
		GetEvents(namespace string, resourceId string) ([]models.K8sEvent, error)

		// Exec
		StartExecProcess(token string, useAdminToken bool, namespace, podName, containerName string, command []string, stdin io.Reader, stdout io.Writer, errChan chan error)

		// ClusterRoleBinding
		GetClusterRoleBindings() ([]models.K8sClusterRoleBinding, error)
		DeleteClusterRoleBindings(reqs models.K8sClusterRoleBindingDeleteRequests) error

		// Dashboard
		GetDashboard() (models.K8sDashboard, error)

		// Deployment
		HasStackName(namespace string, stackName string) (bool, error)

		// Ingress
		GetIngressControllers() (models.K8sIngressControllers, error)
		GetIngress(namespace, ingressName string) (models.K8sIngressInfo, error)
		GetIngresses(namespace string) ([]models.K8sIngressInfo, error)
		CreateIngress(namespace string, info models.K8sIngressInfo, owner string) error
		DeleteIngresses(reqs models.K8sIngressDeleteRequests) error
		UpdateIngress(namespace string, info models.K8sIngressInfo) error
		CombineIngressWithService(ingress models.K8sIngressInfo) (models.K8sIngressInfo, error)
		CombineIngressesWithServices(ingresses []models.K8sIngressInfo) ([]models.K8sIngressInfo, error)

		// Job
		GetJobs(namespace string, includeCronJobChildren bool) ([]models.K8sJob, error)
		DeleteJobs(payload models.K8sJobDeleteRequests) error

		// Metrics
		GetMetrics() (models.K8sMetrics, error)

		// Namespace
		ToggleSystemState(namespaceName string, isSystem bool) error
		UpdateNamespace(info models.K8sNamespaceDetails) (*corev1.Namespace, error)
		GetNamespace(name string) (K8sNamespaceInfo, error)
		CreateNamespace(info models.K8sNamespaceDetails) (*corev1.Namespace, error)
		GetNamespaces() (map[string]K8sNamespaceInfo, error)
		CombineNamespaceWithResourceQuota(namespace K8sNamespaceInfo, w http.ResponseWriter) *httperror.HandlerError
		DeleteNamespace(namespaceName string) (*corev1.Namespace, error)
		CombineNamespacesWithResourceQuotas(namespaces map[string]K8sNamespaceInfo, w http.ResponseWriter) *httperror.HandlerError
		ConvertNamespaceMapToSlice(namespaces map[string]K8sNamespaceInfo) []K8sNamespaceInfo

		// NodeLimits
		GetNodesLimits() (K8sNodesLimits, error)
		GetMaxResourceLimits(skipNamespace string, overCommitEnabled bool, resourceOverCommitPercent int) (K8sNodeLimits, error)

		// Pod
		CreateUserShellPod(ctx context.Context, serviceAccountName, shellPodImage string) (*KubernetesShellPod, error)

		// RBAC
		IsRBACEnabled() (bool, error)

		// Registries
		DeleteRegistrySecret(registry RegistryID, namespace string) error
		CreateRegistrySecret(registry *Registry, namespace string) error
		IsRegistrySecret(namespace, secretName string) (bool, error)

		// RoleBinding
		GetRoleBindings(namespace string) ([]models.K8sRoleBinding, error)
		DeleteRoleBindings(reqs models.K8sRoleBindingDeleteRequests) error

		// Role
		DeleteRoles(reqs models.K8sRoleDeleteRequests) error

		// Secret
		GetSecrets(namespace string) ([]models.K8sSecret, error)
		GetSecret(namespace string, secretName string) (models.K8sSecret, error)
		CombineSecretWithApplications(secret models.K8sSecret) (models.K8sSecret, error)

		// ServiceAccount
		GetServiceAccounts(namespace string) ([]models.K8sServiceAccount, error)
		DeleteServiceAccounts(reqs models.K8sServiceAccountDeleteRequests) error
		SetupUserServiceAccount(int, []int, bool) error
		GetPortainerUserServiceAccount(tokendata *TokenData) (*corev1.ServiceAccount, error)
		GetServiceAccountBearerToken(userID int) (string, error)

		// Service
		GetServices(namespace string) ([]models.K8sServiceInfo, error)
		CombineServicesWithApplications(services []models.K8sServiceInfo) ([]models.K8sServiceInfo, error)
		CreateService(namespace string, info models.K8sServiceInfo) error
		DeleteServices(reqs models.K8sServiceDeleteRequests) error
		UpdateService(namespace string, info models.K8sServiceInfo) error

		// ServerVersion
		ServerVersion() (*version.Info, error)

		// Storage
		GetStorage() ([]KubernetesStorageClassConfig, error)

		// Volumes
		GetVolumes(namespace string) ([]models.K8sVolumeInfo, error)
		GetVolume(namespace, volumeName string) (*models.K8sVolumeInfo, error)
		CombineVolumesWithApplications(volumes *[]models.K8sVolumeInfo) (*[]models.K8sVolumeInfo, error)
	}

	// KubernetesDeployer represents a service to deploy a manifest inside a Kubernetes environment(endpoint)
	KubernetesDeployer interface {
		Deploy(userID UserID, endpoint *Endpoint, manifestFiles []string, namespace string) (string, error)
		Remove(userID UserID, endpoint *Endpoint, manifestFiles []string, namespace string) (string, error)
	}

	// KubernetesSnapshotter represents a service used to create Kubernetes environment(endpoint) snapshots
	KubernetesSnapshotter interface {
		CreateSnapshot(endpoint *Endpoint) (*KubernetesSnapshot, error)
	}

	// LDAPService represents a service used to authenticate users against a LDAP/AD
	LDAPService interface {
		AuthenticateUser(username, password string, settings *LDAPSettings) error
		TestConnectivity(settings *LDAPSettings) error
		GetUserGroups(username string, settings *LDAPSettings) ([]string, error)
		SearchGroups(settings *LDAPSettings) ([]LDAPUser, error)
		SearchUsers(settings *LDAPSettings) ([]string, error)
	}

	// OAuthService represents a service used to authenticate users using OAuth
	OAuthService interface {
		Authenticate(code string, configuration *OAuthSettings) (string, error)
	}

	// ReverseTunnelService represents a service used to manage reverse tunnel connections.
	ReverseTunnelService interface {
		StartTunnelServer(addr, port string, snapshotService SnapshotService) error
		StopTunnelServer() error
		GenerateEdgeKey(apiURL, tunnelAddr string, endpointIdentifier int) string
		Open(endpoint *Endpoint) error
		Config(endpointID EndpointID) TunnelDetails
		TunnelAddr(endpoint *Endpoint) (string, error)
		UpdateLastActivity(endpointID EndpointID)
		KeepTunnelAlive(endpointID EndpointID, ctx context.Context, maxKeepAlive time.Duration)
	}

	// Server defines the interface to serve the API
	Server interface {
		Start() error
	}

	// SnapshotService represents a service for managing environment(endpoint) snapshots
	SnapshotService interface {
		Start()
		SetSnapshotInterval(snapshotInterval string) error
		SnapshotEndpoint(endpoint *Endpoint) error
		FillSnapshotData(endpoint *Endpoint, includeRaw bool) error
	}

	// SwarmStackManager represents a service to manage Swarm stacks
	SwarmStackManager interface {
		Login(registries []Registry, endpoint *Endpoint) error
		Logout(endpoint *Endpoint) error
		Deploy(stack *Stack, prune bool, pullImage bool, endpoint *Endpoint) error
		Remove(stack *Stack, endpoint *Endpoint) error
		NormalizeStackName(name string) string
	}
)

const (
	// APIVersion is the version number of the Portainer API
	APIVersion = "2.31.0"
	// Support annotation for the API version ("STS" for Short-Term Support or "LTS" for Long-Term Support)
	APIVersionSupport = "STS"
	// Edition is what this edition of Portainer is called
	Edition = PortainerCE
	// ComposeSyntaxMaxVersion is a maximum supported version of the docker compose syntax
	ComposeSyntaxMaxVersion = "3.9"
	// AssetsServerURL represents the URL of the Portainer asset server
	AssetsServerURL = "https://portainer-io-assets.sfo2.digitaloceanspaces.com"
	// MessageOfTheDayURL represents the URL where Portainer MOTD message can be retrieved
	MessageOfTheDayURL = AssetsServerURL + "/motd.json"
	// VersionCheckURL represents the URL used to retrieve the latest version of Portainer
	VersionCheckURL = "https://api.github.com/repos/portainer/portainer/releases/latest"
	// PortainerAgentHeader represents the name of the header available in any agent response
	PortainerAgentHeader = "Portainer-Agent"
	// PortainerAgentEdgeIDHeader represent the name of the header containing the Edge ID associated to an agent/agent cluster
	PortainerAgentEdgeIDHeader = "X-PortainerAgent-EdgeID"
	// HTTPResponseAgentPlatform represents the name of the header containing the Agent platform
	HTTPResponseAgentPlatform = "Portainer-Agent-Platform"
	// PortainerAgentTargetHeader represent the name of the header containing the target node name
	PortainerAgentTargetHeader = "X-PortainerAgent-Target"
	// PortainerAgentSignatureHeader represent the name of the header containing the digital signature
	PortainerAgentSignatureHeader = "X-PortainerAgent-Signature"
	// PortainerAgentPublicKeyHeader represent the name of the header containing the public key
	PortainerAgentPublicKeyHeader = "X-PortainerAgent-PublicKey"
	// PortainerAgentKubernetesSATokenHeader represent the name of the header containing a Kubernetes SA token
	PortainerAgentKubernetesSATokenHeader = "X-PortainerAgent-SA-Token"
	// PortainerAgentSignatureMessage represents the message used to create a digital signature
	// to be used when communicating with an agent
	PortainerAgentSignatureMessage = "Portainer-App"
	// DefaultSnapshotInterval represents the default interval between each environment snapshot job
	DefaultSnapshotInterval = "5m"
	// DefaultEdgeAgentCheckinIntervalInSeconds represents the default interval (in seconds) used by Edge agents to checkin with the Portainer instance
	DefaultEdgeAgentCheckinIntervalInSeconds = 5
	// DefaultTemplatesURL represents the URL to the official templates supported by Portainer
	DefaultTemplatesURL = "https://raw.githubusercontent.com/portainer/templates/v3/templates.json"
	// DefaultHelmrepositoryURL represents the URL to the official templates supported by Bitnami
	DefaultHelmRepositoryURL = "https://charts.bitnami.com/bitnami"
	// DefaultUserSessionTimeout represents the default timeout after which the user session is cleared
	DefaultUserSessionTimeout = "8h"
	// DefaultUserSessionTimeout represents the default timeout after which the user session is cleared
	DefaultKubeconfigExpiry = "0"
	// DefaultKubectlShellImage represents the default image and tag for the kubectl shell
	DefaultKubectlShellImage = "portainer/kubectl-shell:" + APIVersion
	// WebSocketKeepAlive web socket keep alive for edge environments
	WebSocketKeepAlive = 1 * time.Hour
	// AuthCookieName is the name of the cookie used to store the JWT token
	AuthCookieKey = "portainer_api_key"
	// PortainerCacheHeader is used to enabled FE caching for Kubernetes resources
	PortainerCacheHeader = "X-Portainer-Cache"
	// KubectlShellImageEnvVar is the environment variable used to override the default kubectl shell image
	KubectlShellImageEnvVar = "KUBECTL_SHELL_IMAGE"
	// PullLimitCheckDisabledEnvVar is the environment variable used to disable the pull limit check
	PullLimitCheckDisabledEnvVar = "PULL_LIMIT_CHECK_DISABLED"
	// LicenseServerBaseURL represents the base URL of the API used to validate
	// an extension license.
	LicenseServerBaseURL = "https://api.portainer.io"
	// URL to validate licenses along with system metadata.
	LicenseCheckInURL = LicenseServerBaseURL + "/licenses/checkin"
	// TrustedOriginsEnvVar is the environment variable used to set the trusted origins for CSRF protection
	TrustedOriginsEnvVar = "TRUSTED_ORIGINS"
	// CSPEnvVar is the environment variable used to enable/disable the Content Security Policy
	CSPEnvVar = "CSP"
)

// List of supported features
var SupportedFeatureFlags = []featureflags.Feature{"hsts", "csp"}

const (
	_ AuthenticationMethod = iota
	// AuthenticationInternal represents the internal authentication method (authentication against Portainer API)
	AuthenticationInternal
	// AuthenticationLDAP represents the LDAP authentication method (authentication against a LDAP server)
	AuthenticationLDAP
	// AuthenticationOAuth represents the OAuth authentication method (authentication against a authorization server)
	AuthenticationOAuth
)

const (
	_ AgentPlatform = iota
	// AgentPlatformDocker represent the Docker platform (Standalone/Swarm)
	AgentPlatformDocker
	// AgentPlatformKubernetes represent the Kubernetes platform
	AgentPlatformKubernetes
)

const (
	_ EdgeJobLogsStatus = iota
	// EdgeJobLogsStatusIdle represents an idle log collection job
	EdgeJobLogsStatusIdle
	// EdgeJobLogsStatusPending represents a pending log collection job
	EdgeJobLogsStatusPending
	// EdgeJobLogsStatusCollected represents a completed log collection job
	EdgeJobLogsStatusCollected
)

const (
	_ CustomTemplatePlatform = iota
	// CustomTemplatePlatformLinux represents a custom template for linux
	CustomTemplatePlatformLinux
	// CustomTemplatePlatformWindows represents a custom template for windows
	CustomTemplatePlatformWindows
)

const (
	// EdgeStackDeploymentCompose represent an edge stack deployed using a compose file
	EdgeStackDeploymentCompose EdgeStackDeploymentType = iota
	// EdgeStackDeploymentKubernetes represent an edge stack deployed using a kubernetes manifest file
	EdgeStackDeploymentKubernetes
)

const (
	// EdgeStackStatusPending represents a pending edge stack
	EdgeStackStatusPending EdgeStackStatusType = iota
	// EdgeStackStatusDeploymentReceived represents an edge environment which received the edge stack deployment
	EdgeStackStatusDeploymentReceived
	// EdgeStackStatusError represents an edge environment which failed to deploy its edge stack
	EdgeStackStatusError
	// EdgeStackStatusAcknowledged represents an acknowledged edge stack
	EdgeStackStatusAcknowledged
	// EdgeStackStatusRemoved represents a removed edge stack
	EdgeStackStatusRemoved
	// StatusRemoteUpdateSuccess represents a successfully updated edge stack
	EdgeStackStatusRemoteUpdateSuccess
	// EdgeStackStatusImagesPulled represents a successfully images-pulling
	EdgeStackStatusImagesPulled
	// EdgeStackStatusRunning represents a running Edge stack
	EdgeStackStatusRunning
	// EdgeStackStatusDeploying represents an Edge stack which is being deployed
	EdgeStackStatusDeploying
	// EdgeStackStatusRemoving represents an Edge stack which is being removed
	EdgeStackStatusRemoving
	// EdgeStackStatusPausedDeploying represents a paused Edge stack
	EdgeStackStatusPausedDeploying
	// EdgeStackStatusRollingBack represents an Edge stack which is being rolled back
	EdgeStackStatusRollingBack
	// EdgeStackStatusRolledBack represents an Edge stack which has rolled back
	EdgeStackStatusRolledBack
	// EdgeStackStatusCompleted represents a completed Edge stack
	EdgeStackStatusCompleted
)

var edgeStackStatusTypeStr = map[EdgeStackStatusType]string{
	EdgeStackStatusPending:             "Pending",
	EdgeStackStatusDeploymentReceived:  "DeploymentReceived",
	EdgeStackStatusError:               "Error",
	EdgeStackStatusAcknowledged:        "Acknowledged",
	EdgeStackStatusRemoved:             "Removed",
	EdgeStackStatusRemoteUpdateSuccess: "RemoteUpdateSuccess",
	EdgeStackStatusImagesPulled:        "ImagesPulled",
	EdgeStackStatusRunning:             "Running",
	EdgeStackStatusDeploying:           "Deploying",
	EdgeStackStatusRemoving:            "Removing",
	EdgeStackStatusPausedDeploying:     "PausedDeploying",
	EdgeStackStatusRollingBack:         "RollingBack",
	EdgeStackStatusRolledBack:          "RolledBack",
	EdgeStackStatusCompleted:           "Completed",
}

func (s EdgeStackStatusType) String() string {
	if str, ok := edgeStackStatusTypeStr[s]; ok {
		return fmt.Sprintf("%d (%s)", s, str)
	}
	return fmt.Sprintf("%d (UNKNOWN)", s)
}

const (
	_ EndpointStatus = iota
	// EndpointStatusUp is used to represent an available environment(endpoint)
	EndpointStatusUp
	// EndpointStatusDown is used to represent an unavailable environment(endpoint)
	EndpointStatusDown
)

const (
	_ EndpointType = iota
	// DockerEnvironment represents an environment(endpoint) connected to a Docker environment(endpoint) via the Docker API or Socket
	DockerEnvironment
	// AgentOnDockerEnvironment represents an environment(endpoint) connected to a Portainer agent deployed on a Docker environment(endpoint)
	AgentOnDockerEnvironment
	// AzureEnvironment represents an environment(endpoint) connected to an Azure environment(endpoint)
	AzureEnvironment
	// EdgeAgentOnDockerEnvironment represents an environment(endpoint) connected to an Edge agent deployed on a Docker environment(endpoint)
	EdgeAgentOnDockerEnvironment
	// KubernetesLocalEnvironment represents an environment(endpoint) connected to a local Kubernetes environment(endpoint)
	KubernetesLocalEnvironment
	// AgentOnKubernetesEnvironment represents an environment(endpoint) connected to a Portainer agent deployed on a Kubernetes environment(endpoint)
	AgentOnKubernetesEnvironment
	// EdgeAgentOnKubernetesEnvironment represents an environment(endpoint) connected to an Edge agent deployed on a Kubernetes environment(endpoint)
	EdgeAgentOnKubernetesEnvironment
)

const (
	_ JobType = iota
	// SnapshotJobType is a system job used to create environment(endpoint) snapshots
	SnapshotJobType = 2
)

const (
	_ MembershipRole = iota
	// TeamLeader represents a leader role inside a team
	TeamLeader
	// TeamMember represents a member role inside a team
	TeamMember
)

const (
	_ SoftwareEdition = iota
	// PortainerCE represents the community edition of Portainer
	PortainerCE
	// PortainerBE represents the business edition of Portainer
	PortainerBE
	// PortainerEE represents the business edition of Portainer
	PortainerEE
)

const (
	_ RegistryType = iota
	// QuayRegistry represents a Quay.io registry
	QuayRegistry
	// AzureRegistry represents an ACR registry
	AzureRegistry
	// CustomRegistry represents a custom registry
	CustomRegistry
	// GitlabRegistry represents a gitlab registry
	GitlabRegistry
	// ProGetRegistry represents a proget registry
	ProGetRegistry
	// DockerHubRegistry represents a dockerhub registry
	DockerHubRegistry
	// EcrRegistry represents an ECR registry
	EcrRegistry
	// Github container registry
	GithubRegistry
)

const (
	_ ResourceAccessLevel = iota
	// ReadWriteAccessLevel represents an access level with read-write permissions on a resource
	ReadWriteAccessLevel
)

const (
	_ ResourceControlType = iota
	// ContainerResourceControl represents a resource control associated to a Docker container
	ContainerResourceControl
	// ServiceResourceControl represents a resource control associated to a Docker service
	ServiceResourceControl
	// VolumeResourceControl represents a resource control associated to a Docker volume
	VolumeResourceControl
	// NetworkResourceControl represents a resource control associated to a Docker network
	NetworkResourceControl
	// SecretResourceControl represents a resource control associated to a Docker secret
	SecretResourceControl
	// StackResourceControl represents a resource control associated to a stack composed of Docker services
	StackResourceControl
	// ConfigResourceControl represents a resource control associated to a Docker config
	ConfigResourceControl
	// CustomTemplateResourceControl represents a resource control associated to a custom template
	CustomTemplateResourceControl
	// ContainerGroupResourceControl represents a resource control associated to an Azure container group
	ContainerGroupResourceControl
)

const (
	_ StackType = iota
	// DockerSwarmStack represents a stack managed via docker stack
	DockerSwarmStack
	// DockerComposeStack represents a stack managed via docker-compose
	DockerComposeStack
	// KubernetesStack represents a stack managed via kubectl
	KubernetesStack
)

// StackStatus represents a status for a stack
const (
	_ StackStatus = iota
	StackStatusActive
	StackStatusInactive
)

const (
	_ TemplateType = iota
	// ContainerTemplate represents a container template
	ContainerTemplate
	// SwarmStackTemplate represents a template used to deploy a Swarm stack
	SwarmStackTemplate
	// ComposeStackTemplate represents a template used to deploy a Compose stack
	ComposeStackTemplate
)

const (
	// TLSFileCA represents a TLS CA certificate file
	TLSFileCA TLSFileType = iota
	// TLSFileCert represents a TLS certificate file
	TLSFileCert
	// TLSFileKey represents a TLS key file
	TLSFileKey
)

const (
	_ UserRole = iota
	// AdministratorRole represents an administrator user role
	AdministratorRole
	// StandardUserRole represents a regular user role
	StandardUserRole
)

const (
	_ WebhookType = iota
	// ServiceWebhook is a webhook for restarting a docker service
	ServiceWebhook
)

const (
	// EdgeAgentIdle represents an idle state for a tunnel connected to an Edge environment(endpoint).
	EdgeAgentIdle string = "IDLE"
	// EdgeAgentManagementRequired represents a required state for a tunnel connected to an Edge environment(endpoint)
	EdgeAgentManagementRequired string = "REQUIRED"
)

// represents an authorization type
const (
	OperationDockerContainerArchiveInfo         Authorization = "DockerContainerArchiveInfo"
	OperationDockerContainerList                Authorization = "DockerContainerList"
	OperationDockerContainerExport              Authorization = "DockerContainerExport"
	OperationDockerContainerChanges             Authorization = "DockerContainerChanges"
	OperationDockerContainerInspect             Authorization = "DockerContainerInspect"
	OperationDockerContainerTop                 Authorization = "DockerContainerTop"
	OperationDockerContainerLogs                Authorization = "DockerContainerLogs"
	OperationDockerContainerStats               Authorization = "DockerContainerStats"
	OperationDockerContainerAttachWebsocket     Authorization = "DockerContainerAttachWebsocket"
	OperationDockerContainerArchive             Authorization = "DockerContainerArchive"
	OperationDockerContainerCreate              Authorization = "DockerContainerCreate"
	OperationDockerContainerPrune               Authorization = "DockerContainerPrune"
	OperationDockerContainerKill                Authorization = "DockerContainerKill"
	OperationDockerContainerPause               Authorization = "DockerContainerPause"
	OperationDockerContainerUnpause             Authorization = "DockerContainerUnpause"
	OperationDockerContainerRestart             Authorization = "DockerContainerRestart"
	OperationDockerContainerStart               Authorization = "DockerContainerStart"
	OperationDockerContainerStop                Authorization = "DockerContainerStop"
	OperationDockerContainerWait                Authorization = "DockerContainerWait"
	OperationDockerContainerResize              Authorization = "DockerContainerResize"
	OperationDockerContainerAttach              Authorization = "DockerContainerAttach"
	OperationDockerContainerExec                Authorization = "DockerContainerExec"
	OperationDockerContainerRename              Authorization = "DockerContainerRename"
	OperationDockerContainerUpdate              Authorization = "DockerContainerUpdate"
	OperationDockerContainerPutContainerArchive Authorization = "DockerContainerPutContainerArchive"
	OperationDockerContainerDelete              Authorization = "DockerContainerDelete"
	OperationDockerImageList                    Authorization = "DockerImageList"
	OperationDockerImageSearch                  Authorization = "DockerImageSearch"
	OperationDockerImageGetAll                  Authorization = "DockerImageGetAll"
	OperationDockerImageGet                     Authorization = "DockerImageGet"
	OperationDockerImageHistory                 Authorization = "DockerImageHistory"
	OperationDockerImageInspect                 Authorization = "DockerImageInspect"
	OperationDockerImageLoad                    Authorization = "DockerImageLoad"
	OperationDockerImageCreate                  Authorization = "DockerImageCreate"
	OperationDockerImagePrune                   Authorization = "DockerImagePrune"
	OperationDockerImagePush                    Authorization = "DockerImagePush"
	OperationDockerImageTag                     Authorization = "DockerImageTag"
	OperationDockerImageDelete                  Authorization = "DockerImageDelete"
	OperationDockerImageCommit                  Authorization = "DockerImageCommit"
	OperationDockerImageBuild                   Authorization = "DockerImageBuild"
	OperationDockerNetworkList                  Authorization = "DockerNetworkList"
	OperationDockerNetworkInspect               Authorization = "DockerNetworkInspect"
	OperationDockerNetworkCreate                Authorization = "DockerNetworkCreate"
	OperationDockerNetworkConnect               Authorization = "DockerNetworkConnect"
	OperationDockerNetworkDisconnect            Authorization = "DockerNetworkDisconnect"
	OperationDockerNetworkPrune                 Authorization = "DockerNetworkPrune"
	OperationDockerNetworkDelete                Authorization = "DockerNetworkDelete"
	OperationDockerVolumeList                   Authorization = "DockerVolumeList"
	OperationDockerVolumeInspect                Authorization = "DockerVolumeInspect"
	OperationDockerVolumeCreate                 Authorization = "DockerVolumeCreate"
	OperationDockerVolumePrune                  Authorization = "DockerVolumePrune"
	OperationDockerVolumeDelete                 Authorization = "DockerVolumeDelete"
	OperationDockerExecInspect                  Authorization = "DockerExecInspect"
	OperationDockerExecStart                    Authorization = "DockerExecStart"
	OperationDockerExecResize                   Authorization = "DockerExecResize"
	OperationDockerSwarmInspect                 Authorization = "DockerSwarmInspect"
	OperationDockerSwarmUnlockKey               Authorization = "DockerSwarmUnlockKey"
	OperationDockerSwarmInit                    Authorization = "DockerSwarmInit"
	OperationDockerSwarmJoin                    Authorization = "DockerSwarmJoin"
	OperationDockerSwarmLeave                   Authorization = "DockerSwarmLeave"
	OperationDockerSwarmUpdate                  Authorization = "DockerSwarmUpdate"
	OperationDockerSwarmUnlock                  Authorization = "DockerSwarmUnlock"
	OperationDockerNodeList                     Authorization = "DockerNodeList"
	OperationDockerNodeInspect                  Authorization = "DockerNodeInspect"
	OperationDockerNodeUpdate                   Authorization = "DockerNodeUpdate"
	OperationDockerNodeDelete                   Authorization = "DockerNodeDelete"
	OperationDockerServiceList                  Authorization = "DockerServiceList"
	OperationDockerServiceInspect               Authorization = "DockerServiceInspect"
	OperationDockerServiceLogs                  Authorization = "DockerServiceLogs"
	OperationDockerServiceCreate                Authorization = "DockerServiceCreate"
	OperationDockerServiceUpdate                Authorization = "DockerServiceUpdate"
	OperationDockerServiceDelete                Authorization = "DockerServiceDelete"
	OperationDockerSecretList                   Authorization = "DockerSecretList"
	OperationDockerSecretInspect                Authorization = "DockerSecretInspect"
	OperationDockerSecretCreate                 Authorization = "DockerSecretCreate"
	OperationDockerSecretUpdate                 Authorization = "DockerSecretUpdate"
	OperationDockerSecretDelete                 Authorization = "DockerSecretDelete"
	OperationDockerConfigList                   Authorization = "DockerConfigList"
	OperationDockerConfigInspect                Authorization = "DockerConfigInspect"
	OperationDockerConfigCreate                 Authorization = "DockerConfigCreate"
	OperationDockerConfigUpdate                 Authorization = "DockerConfigUpdate"
	OperationDockerConfigDelete                 Authorization = "DockerConfigDelete"
	OperationDockerTaskList                     Authorization = "DockerTaskList"
	OperationDockerTaskInspect                  Authorization = "DockerTaskInspect"
	OperationDockerTaskLogs                     Authorization = "DockerTaskLogs"
	OperationDockerPluginList                   Authorization = "DockerPluginList"
	OperationDockerPluginPrivileges             Authorization = "DockerPluginPrivileges"
	OperationDockerPluginInspect                Authorization = "DockerPluginInspect"
	OperationDockerPluginPull                   Authorization = "DockerPluginPull"
	OperationDockerPluginCreate                 Authorization = "DockerPluginCreate"
	OperationDockerPluginEnable                 Authorization = "DockerPluginEnable"
	OperationDockerPluginDisable                Authorization = "DockerPluginDisable"
	OperationDockerPluginPush                   Authorization = "DockerPluginPush"
	OperationDockerPluginUpgrade                Authorization = "DockerPluginUpgrade"
	OperationDockerPluginSet                    Authorization = "DockerPluginSet"
	OperationDockerPluginDelete                 Authorization = "DockerPluginDelete"
	OperationDockerSessionStart                 Authorization = "DockerSessionStart"
	OperationDockerDistributionInspect          Authorization = "DockerDistributionInspect"
	OperationDockerBuildPrune                   Authorization = "DockerBuildPrune"
	OperationDockerBuildCancel                  Authorization = "DockerBuildCancel"
	OperationDockerPing                         Authorization = "DockerPing"
	OperationDockerInfo                         Authorization = "DockerInfo"
	OperationDockerEvents                       Authorization = "DockerEvents"
	OperationDockerSystem                       Authorization = "DockerSystem"
	OperationDockerVersion                      Authorization = "DockerVersion"

	OperationDockerAgentPing         Authorization = "DockerAgentPing"
	OperationDockerAgentList         Authorization = "DockerAgentList"
	OperationDockerAgentHostInfo     Authorization = "DockerAgentHostInfo"
	OperationDockerAgentBrowseDelete Authorization = "DockerAgentBrowseDelete"
	OperationDockerAgentBrowseGet    Authorization = "DockerAgentBrowseGet"
	OperationDockerAgentBrowseList   Authorization = "DockerAgentBrowseList"
	OperationDockerAgentBrowsePut    Authorization = "DockerAgentBrowsePut"
	OperationDockerAgentBrowseRename Authorization = "DockerAgentBrowseRename"

	OperationPortainerDockerHubInspect      Authorization = "PortainerDockerHubInspect"
	OperationPortainerDockerHubUpdate       Authorization = "PortainerDockerHubUpdate"
	OperationPortainerEndpointGroupCreate   Authorization = "PortainerEndpointGroupCreate"
	OperationPortainerEndpointGroupList     Authorization = "PortainerEndpointGroupList"
	OperationPortainerEndpointGroupDelete   Authorization = "PortainerEndpointGroupDelete"
	OperationPortainerEndpointGroupInspect  Authorization = "PortainerEndpointGroupInspect"
	OperationPortainerEndpointGroupUpdate   Authorization = "PortainerEndpointGroupEdit"
	OperationPortainerEndpointGroupAccess   Authorization = "PortainerEndpointGroupAccess "
	OperationPortainerEndpointList          Authorization = "PortainerEndpointList"
	OperationPortainerEndpointInspect       Authorization = "PortainerEndpointInspect"
	OperationPortainerEndpointCreate        Authorization = "PortainerEndpointCreate"
	OperationPortainerEndpointJob           Authorization = "PortainerEndpointJob"
	OperationPortainerEndpointSnapshots     Authorization = "PortainerEndpointSnapshots"
	OperationPortainerEndpointSnapshot      Authorization = "PortainerEndpointSnapshot"
	OperationPortainerEndpointUpdate        Authorization = "PortainerEndpointUpdate"
	OperationPortainerEndpointUpdateAccess  Authorization = "PortainerEndpointUpdateAccess"
	OperationPortainerEndpointDelete        Authorization = "PortainerEndpointDelete"
	OperationPortainerExtensionList         Authorization = "PortainerExtensionList"
	OperationPortainerExtensionInspect      Authorization = "PortainerExtensionInspect"
	OperationPortainerExtensionCreate       Authorization = "PortainerExtensionCreate"
	OperationPortainerExtensionUpdate       Authorization = "PortainerExtensionUpdate"
	OperationPortainerExtensionDelete       Authorization = "PortainerExtensionDelete"
	OperationPortainerMOTD                  Authorization = "PortainerMOTD"
	OperationPortainerRegistryList          Authorization = "PortainerRegistryList"
	OperationPortainerRegistryInspect       Authorization = "PortainerRegistryInspect"
	OperationPortainerRegistryCreate        Authorization = "PortainerRegistryCreate"
	OperationPortainerRegistryConfigure     Authorization = "PortainerRegistryConfigure"
	OperationPortainerRegistryUpdate        Authorization = "PortainerRegistryUpdate"
	OperationPortainerRegistryUpdateAccess  Authorization = "PortainerRegistryUpdateAccess"
	OperationPortainerRegistryDelete        Authorization = "PortainerRegistryDelete"
	OperationPortainerResourceControlCreate Authorization = "PortainerResourceControlCreate"
	OperationPortainerResourceControlUpdate Authorization = "PortainerResourceControlUpdate"
	OperationPortainerResourceControlDelete Authorization = "PortainerResourceControlDelete"
	OperationPortainerRoleList              Authorization = "PortainerRoleList"
	OperationPortainerRoleInspect           Authorization = "PortainerRoleInspect"
	OperationPortainerRoleCreate            Authorization = "PortainerRoleCreate"
	OperationPortainerRoleUpdate            Authorization = "PortainerRoleUpdate"
	OperationPortainerRoleDelete            Authorization = "PortainerRoleDelete"
	OperationPortainerScheduleList          Authorization = "PortainerScheduleList"
	OperationPortainerScheduleInspect       Authorization = "PortainerScheduleInspect"
	OperationPortainerScheduleFile          Authorization = "PortainerScheduleFile"
	OperationPortainerScheduleTasks         Authorization = "PortainerScheduleTasks"
	OperationPortainerScheduleCreate        Authorization = "PortainerScheduleCreate"
	OperationPortainerScheduleUpdate        Authorization = "PortainerScheduleUpdate"
	OperationPortainerScheduleDelete        Authorization = "PortainerScheduleDelete"
	OperationPortainerSettingsInspect       Authorization = "PortainerSettingsInspect"
	OperationPortainerSettingsUpdate        Authorization = "PortainerSettingsUpdate"
	OperationPortainerSettingsLDAPCheck     Authorization = "PortainerSettingsLDAPCheck"
	OperationPortainerStackList             Authorization = "PortainerStackList"
	OperationPortainerStackInspect          Authorization = "PortainerStackInspect"
	OperationPortainerStackFile             Authorization = "PortainerStackFile"
	OperationPortainerStackCreate           Authorization = "PortainerStackCreate"
	OperationPortainerStackMigrate          Authorization = "PortainerStackMigrate"
	OperationPortainerStackUpdate           Authorization = "PortainerStackUpdate"
	OperationPortainerStackDelete           Authorization = "PortainerStackDelete"
	OperationPortainerTagList               Authorization = "PortainerTagList"
	OperationPortainerTagCreate             Authorization = "PortainerTagCreate"
	OperationPortainerTagDelete             Authorization = "PortainerTagDelete"
	OperationPortainerTeamMembershipList    Authorization = "PortainerTeamMembershipList"
	OperationPortainerTeamMembershipCreate  Authorization = "PortainerTeamMembershipCreate"
	OperationPortainerTeamMembershipUpdate  Authorization = "PortainerTeamMembershipUpdate"
	OperationPortainerTeamMembershipDelete  Authorization = "PortainerTeamMembershipDelete"
	OperationPortainerTeamList              Authorization = "PortainerTeamList"
	OperationPortainerTeamInspect           Authorization = "PortainerTeamInspect"
	OperationPortainerTeamMemberships       Authorization = "PortainerTeamMemberships"
	OperationPortainerTeamCreate            Authorization = "PortainerTeamCreate"
	OperationPortainerTeamUpdate            Authorization = "PortainerTeamUpdate"
	OperationPortainerTeamDelete            Authorization = "PortainerTeamDelete"
	OperationPortainerTemplateList          Authorization = "PortainerTemplateList"
	OperationPortainerTemplateInspect       Authorization = "PortainerTemplateInspect"
	OperationPortainerTemplateCreate        Authorization = "PortainerTemplateCreate"
	OperationPortainerTemplateUpdate        Authorization = "PortainerTemplateUpdate"
	OperationPortainerTemplateDelete        Authorization = "PortainerTemplateDelete"
	OperationPortainerUploadTLS             Authorization = "PortainerUploadTLS"
	OperationPortainerUserList              Authorization = "PortainerUserList"
	OperationPortainerUserInspect           Authorization = "PortainerUserInspect"
	OperationPortainerUserMemberships       Authorization = "PortainerUserMemberships"
	OperationPortainerUserCreate            Authorization = "PortainerUserCreate"
	OperationPortainerUserListToken         Authorization = "PortainerUserListToken"
	OperationPortainerUserCreateToken       Authorization = "PortainerUserCreateToken"
	OperationPortainerUserRevokeToken       Authorization = "PortainerUserRevokeToken"
	OperationPortainerUserUpdate            Authorization = "PortainerUserUpdate"
	OperationPortainerUserUpdatePassword    Authorization = "PortainerUserUpdatePassword"
	OperationPortainerUserDelete            Authorization = "PortainerUserDelete"
	OperationPortainerWebsocketExec         Authorization = "PortainerWebsocketExec"
	OperationPortainerWebhookList           Authorization = "PortainerWebhookList"
	OperationPortainerWebhookCreate         Authorization = "PortainerWebhookCreate"
	OperationPortainerWebhookDelete         Authorization = "PortainerWebhookDelete"

	OperationDockerUndefined      Authorization = "DockerUndefined"
	OperationDockerAgentUndefined Authorization = "DockerAgentUndefined"
	OperationPortainerUndefined   Authorization = "PortainerUndefined"

	EndpointResourcesAccess Authorization = "EndpointResourcesAccess"

	// Deprecated operations
	OperationPortainerEndpointExtensionAdd    Authorization = "PortainerEndpointExtensionAdd"
	OperationPortainerEndpointExtensionRemove Authorization = "PortainerEndpointExtensionRemove"
	OperationIntegrationStoridgeAdmin         Authorization = "IntegrationStoridgeAdmin"
)

// GetEditionLabel returns the portainer edition label
func (e SoftwareEdition) GetEditionLabel() string {
	switch e {
	case PortainerCE:
		return "CE"
	case PortainerBE:
		return "BE"
	case PortainerEE:
		return "EE"
	}

	return "CE"
}

const (
	AzurePathContainerGroups = "/subscriptions/*/providers/Microsoft.ContainerInstance/containerGroups"
	AzurePathContainerGroup  = "/subscriptions/*/resourceGroups/*/providers/Microsoft.ContainerInstance/containerGroups/*"
)

type PerDevConfigsFilterType string

const (
	PerDevConfigsTypeFile PerDevConfigsFilterType = "file"
	PerDevConfigsTypeDir  PerDevConfigsFilterType = "dir"
)

const (
	ContainerEngineDocker = "docker"
	ContainerEnginePodman = "podman"
)
