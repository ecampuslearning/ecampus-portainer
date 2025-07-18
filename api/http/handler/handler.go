package handler

import (
	"net/http"
	"strings"

	"github.com/portainer/portainer/api/http/handler/auth"
	"github.com/portainer/portainer/api/http/handler/backup"
	"github.com/portainer/portainer/api/http/handler/customtemplates"
	"github.com/portainer/portainer/api/http/handler/docker"
	"github.com/portainer/portainer/api/http/handler/edgegroups"
	"github.com/portainer/portainer/api/http/handler/edgejobs"
	"github.com/portainer/portainer/api/http/handler/edgestacks"
	"github.com/portainer/portainer/api/http/handler/endpointedge"
	"github.com/portainer/portainer/api/http/handler/endpointgroups"
	"github.com/portainer/portainer/api/http/handler/endpointproxy"
	"github.com/portainer/portainer/api/http/handler/endpoints"
	"github.com/portainer/portainer/api/http/handler/file"
	"github.com/portainer/portainer/api/http/handler/gitops"
	"github.com/portainer/portainer/api/http/handler/helm"
	"github.com/portainer/portainer/api/http/handler/hostmanagement/openamt"
	"github.com/portainer/portainer/api/http/handler/kubernetes"
	"github.com/portainer/portainer/api/http/handler/ldap"
	"github.com/portainer/portainer/api/http/handler/motd"
	"github.com/portainer/portainer/api/http/handler/registries"
	"github.com/portainer/portainer/api/http/handler/resourcecontrols"
	"github.com/portainer/portainer/api/http/handler/roles"
	"github.com/portainer/portainer/api/http/handler/settings"
	"github.com/portainer/portainer/api/http/handler/ssl"
	"github.com/portainer/portainer/api/http/handler/stacks"
	"github.com/portainer/portainer/api/http/handler/storybook"
	"github.com/portainer/portainer/api/http/handler/system"
	"github.com/portainer/portainer/api/http/handler/tags"
	"github.com/portainer/portainer/api/http/handler/teammemberships"
	"github.com/portainer/portainer/api/http/handler/teams"
	"github.com/portainer/portainer/api/http/handler/templates"
	"github.com/portainer/portainer/api/http/handler/upload"
	"github.com/portainer/portainer/api/http/handler/users"
	"github.com/portainer/portainer/api/http/handler/webhooks"
	"github.com/portainer/portainer/api/http/handler/websocket"
)

// Handler is a collection of all the service handlers.
type Handler struct {
	AuthHandler            *auth.Handler
	BackupHandler          *backup.Handler
	CustomTemplatesHandler *customtemplates.Handler
	DockerHandler          *docker.Handler
	EdgeGroupsHandler      *edgegroups.Handler
	EdgeJobsHandler        *edgejobs.Handler
	EdgeStacksHandler      *edgestacks.Handler
	EndpointEdgeHandler    *endpointedge.Handler
	EndpointGroupHandler   *endpointgroups.Handler
	EndpointHandler        *endpoints.Handler
	EndpointHelmHandler    *helm.Handler
	EndpointProxyHandler   *endpointproxy.Handler
	GitOperationHandler    *gitops.Handler
	HelmTemplatesHandler   *helm.Handler
	KubernetesHandler      *kubernetes.Handler
	FileHandler            *file.Handler
	LDAPHandler            *ldap.Handler
	MOTDHandler            *motd.Handler
	RegistryHandler        *registries.Handler
	ResourceControlHandler *resourcecontrols.Handler
	RoleHandler            *roles.Handler
	SettingsHandler        *settings.Handler
	SSLHandler             *ssl.Handler
	OpenAMTHandler         *openamt.Handler
	StackHandler           *stacks.Handler
	StorybookHandler       *storybook.Handler
	SystemHandler          *system.Handler
	TagHandler             *tags.Handler
	TeamMembershipHandler  *teammemberships.Handler
	TeamHandler            *teams.Handler
	TemplatesHandler       *templates.Handler
	UploadHandler          *upload.Handler
	UserHandler            *users.Handler
	WebSocketHandler       *websocket.Handler
	WebhookHandler         *webhooks.Handler
	UserHelmHandler        *helm.Handler
}

// @title PortainerCE API
// @version 2.31.0
// @description.markdown api-description.md
// @termsOfService

// @contact.email info@portainer.io

// @license.name zlib
// @license.url https://github.com/portainer/portainer/blob/develop/LICENSE

// @host
// @BasePath /api
// @schemes http https

// @securitydefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-KEY

// @securitydefinitions.apikey jwt
// @in header
// @name Authorization

// @tag.name auth
// @tag.description Authenticate against Portainer HTTP API
// @tag.name backup
// @tag.description Manage backups
// @tag.name custom_templates
// @tag.description Manage Custom Templates
// @tag.name docker
// @tag.description Manage Docker resources
// @tag.name edge
// @tag.description Manage Edge related environment(endpoint) settings
// @tag.name edge_groups
// @tag.description Manage Edge Groups
// @tag.name edge_jobs
// @tag.description Manage Edge Jobs
// @tag.name edge_stacks
// @tag.description Manage Edge Stacks
// @tag.name edge_templates
// @tag.description Manage Edge Templates
// @tag.name endpoint_groups
// @tag.description Manage environment(endpoint) groups
// @tag.name endpoints
// @tag.description Manage Docker environments(endpoints)
// @tag.name gitops
// @tag.description Operate git repository
// @tag.name helm
// @tag.description Manage Helm charts
// @tag.name intel
// @tag.description Manage Intel AMT settings
// @tag.name kubernetes
// @tag.description Manage Kubernetes cluster
// @tag.name ldap
// @tag.description Manage LDAP settings
// @tag.name motd
// @tag.description Fetch the message of the day
// @tag.name registries
// @tag.description Manage Docker registries
// @tag.name resource_controls
// @tag.description Manage access control on Docker resources
// @tag.name roles
// @tag.description Manage roles
// @tag.name settings
// @tag.description Manage Portainer settings
// @tag.name ssl
// @tag.description Manage ssl settings
// @tag.name stacks
// @tag.description Manage stacks
// @tag.name status
// @tag.description Information about the Portainer instance
// @tag.name system
// @tag.description Manage Portainer system
// @tag.name tags
// @tag.description Manage tags
// @tag.name team_memberships
// @tag.description Manage team memberships
// @tag.name teams
// @tag.description Manage teams
// @tag.name templates
// @tag.description Manage App Templates
// @tag.name upload
// @tag.description Upload files
// @tag.name users
// @tag.description Manage users
// @tag.name webhooks
// @tag.description Manage webhooks
// @tag.name websocket
// @tag.description Create exec sessions using websockets

// ServeHTTP delegates a request to the appropriate subhandler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasPrefix(r.URL.Path, "/api/endpoints") && strings.Contains(r.URL.Path, "/edge/"):
		h.EndpointEdgeHandler.ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/auth"):
		http.StripPrefix("/api", h.AuthHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/backup"):
		http.StripPrefix("/api", h.BackupHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/restore"):
		http.StripPrefix("/api", h.BackupHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/custom_templates"):
		http.StripPrefix("/api", h.CustomTemplatesHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/edge_stacks"):
		http.StripPrefix("/api", h.EdgeStacksHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/edge_groups"):
		http.StripPrefix("/api", h.EdgeGroupsHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/edge_jobs"):
		http.StripPrefix("/api", h.EdgeJobsHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/endpoint_groups"):
		http.StripPrefix("/api", h.EndpointGroupHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/kubernetes"):
		http.StripPrefix("/api", h.KubernetesHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/docker"):
		http.StripPrefix("/api", h.DockerHandler).ServeHTTP(w, r)

	// Helm subpath under kubernetes -> /api/endpoints/{id}/kubernetes/helm
	case strings.HasPrefix(r.URL.Path, "/api/endpoints/") && strings.Contains(r.URL.Path, "/kubernetes/helm"):
		http.StripPrefix("/api/endpoints", h.EndpointHelmHandler).ServeHTTP(w, r)

	case strings.HasPrefix(r.URL.Path, "/api/endpoints"):
		switch {
		case strings.Contains(r.URL.Path, "/docker/"):
			http.StripPrefix("/api/endpoints", h.EndpointProxyHandler).ServeHTTP(w, r)
		case strings.Contains(r.URL.Path, "/kubernetes/"):
			http.StripPrefix("/api/endpoints", h.EndpointProxyHandler).ServeHTTP(w, r)
		case strings.Contains(r.URL.Path, "/azure/"):
			http.StripPrefix("/api/endpoints", h.EndpointProxyHandler).ServeHTTP(w, r)
		case strings.Contains(r.URL.Path, "/agent/"):
			http.StripPrefix("/api/endpoints", h.EndpointProxyHandler).ServeHTTP(w, r)
		default:
			http.StripPrefix("/api", h.EndpointHandler).ServeHTTP(w, r)
		}
	case strings.HasPrefix(r.URL.Path, "/api/gitops"):
		http.StripPrefix("/api", h.GitOperationHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/ldap"):
		http.StripPrefix("/api", h.LDAPHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/motd"):
		http.StripPrefix("/api", h.MOTDHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/registries"):
		http.StripPrefix("/api", h.RegistryHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/resource_controls"):
		http.StripPrefix("/api", h.ResourceControlHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/roles"):
		http.StripPrefix("/api", h.RoleHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/settings"):
		http.StripPrefix("/api", h.SettingsHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/stacks"):
		http.StripPrefix("/api", h.StackHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/status"):
		http.StripPrefix("/api", h.SystemHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/system"):
		http.StripPrefix("/api", h.SystemHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/tags"):
		http.StripPrefix("/api", h.TagHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/templates/helm"):
		http.StripPrefix("/api", h.HelmTemplatesHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/templates"):
		http.StripPrefix("/api", h.TemplatesHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/upload"):
		http.StripPrefix("/api", h.UploadHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/users"):
		http.StripPrefix("/api", h.UserHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/ssl"):
		http.StripPrefix("/api", h.SSLHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/open_amt"):
		http.StripPrefix("/api", h.OpenAMTHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/teams"):
		http.StripPrefix("/api", h.TeamHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/team_memberships"):
		http.StripPrefix("/api", h.TeamMembershipHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/websocket"):
		http.StripPrefix("/api", h.WebSocketHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/webhooks"):
		http.StripPrefix("/api", h.WebhookHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/storybook"):
		http.StripPrefix("/storybook", h.StorybookHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/"):
		h.FileHandler.ServeHTTP(w, r)
	}
}
