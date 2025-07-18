package helm

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/portainer/portainer/api/http/middlewares"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/kubernetes"
	"github.com/portainer/portainer/api/kubernetes/validation"
	"github.com/portainer/portainer/pkg/libhelm/options"
	"github.com/portainer/portainer/pkg/libhelm/release"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type installChartPayload struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Chart     string `json:"chart"`
	Repo      string `json:"repo"`
	Values    string `json:"values"`
	Version   string `json:"version"`
	Atomic    bool   `json:"atomic"`
}

var errChartNameInvalid = errors.New("invalid chart name. " +
	"Chart name must consist of lower case alphanumeric characters, '-' or '.'," +
	" and must start and end with an alphanumeric character",
)

// @id HelmInstall
// @summary Install Helm Chart
// @description
// @description **Access policy**: authenticated
// @tags helm
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path int true "Environment(Endpoint) identifier"
// @param payload body installChartPayload true "Chart details"
// @success 201 {object} release.Release "Created"
// @failure 401 "Unauthorized"
// @failure 404 "Environment(Endpoint) or ServiceAccount not found"
// @failure 500 "Server error"
// @router /endpoints/{id}/kubernetes/helm [post]
func (handler *Handler) helmInstall(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload installChartPayload
	if err := request.DecodeAndValidateJSONPayload(r, &payload); err != nil {
		return httperror.BadRequest("Invalid Helm install payload", err)
	}

	release, err := handler.installChart(r, payload)
	if err != nil {
		return httperror.InternalServerError("Unable to install a chart", err)
	}

	return response.JSONWithStatus(w, release, http.StatusCreated)
}

func (p *installChartPayload) Validate(_ *http.Request) error {
	var required []string
	if p.Repo == "" {
		required = append(required, "repo")
	}

	if p.Name == "" {
		required = append(required, "name")
	}

	if p.Namespace == "" {
		required = append(required, "namespace")
	}

	if p.Chart == "" {
		required = append(required, "chart")
	}

	if len(required) > 0 {
		return fmt.Errorf("required field(s) missing: %s", strings.Join(required, ", "))
	}

	if errs := validation.IsDNS1123Subdomain(p.Name); len(errs) > 0 {
		return errChartNameInvalid
	}

	return nil
}

func (handler *Handler) installChart(r *http.Request, p installChartPayload) (*release.Release, error) {
	clusterAccess, httperr := handler.getHelmClusterAccess(r)
	if httperr != nil {
		return nil, httperr.Err
	}

	installOpts := options.InstallOptions{
		Name:                    p.Name,
		Chart:                   p.Chart,
		Version:                 p.Version,
		Namespace:               p.Namespace,
		Repo:                    p.Repo,
		Atomic:                  p.Atomic,
		KubernetesClusterAccess: clusterAccess,
	}

	if p.Values != "" {
		file, err := os.CreateTemp("", "helm-values")
		if err != nil {
			return nil, err
		}
		defer os.Remove(file.Name())

		if _, err := file.WriteString(p.Values); err != nil {
			file.Close()
			return nil, err
		}

		if err := file.Close(); err != nil {
			return nil, err
		}

		installOpts.ValuesFile = file.Name()
	}

	release, err := handler.helmPackageManager.Upgrade(installOpts)
	if err != nil {
		return nil, err
	}

	manifest, err := handler.applyPortainerLabelsToHelmAppManifest(r, installOpts, release.Manifest)
	if err != nil {
		return nil, err
	}

	if err := handler.updateHelmAppManifest(r, manifest, installOpts.Namespace); err != nil {
		return nil, err
	}

	return release, nil
}

// applyPortainerLabelsToHelmAppManifest will patch all the resources deployed in the helm release manifest
// with portainer specific labels. This is to mark the resources as managed by portainer - hence the helm apps
// wont appear external in the portainer UI.
func (handler *Handler) applyPortainerLabelsToHelmAppManifest(r *http.Request, installOpts options.InstallOptions, manifest string) ([]byte, error) {
	// Patch helm release by adding with portainer labels to all deployed resources
	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve user details from authentication token")
	}

	user, err := handler.dataStore.User().Read(tokenData.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load user information from the database")
	}

	appLabels := kubernetes.GetHelmAppLabels(installOpts.Name, user.Username)

	labeledManifest, err := kubernetes.AddAppLabels([]byte(manifest), appLabels)
	if err != nil {
		return nil, errors.Wrap(err, "failed to label helm release manifest")
	}

	return labeledManifest, nil
}

// updateHelmAppManifest will update the resources of helm release manifest with portainer labels using kubectl.
// The resources of the manifest will be updated in parallel and individuallly since resources of a chart
// can be deployed to different namespaces.
// NOTE: These updates will need to be re-applied when upgrading the helm release
func (handler *Handler) updateHelmAppManifest(r *http.Request, manifest []byte, namespace string) error {
	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return errors.Wrap(err, "unable to find an endpoint on request context")
	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return errors.Wrap(err, "unable to retrieve user details from authentication token")
	}

	// Extract list of YAML resources from Helm manifest
	yamlResources, err := kubernetes.ExtractDocuments(manifest, nil)
	if err != nil {
		return errors.Wrap(err, "unable to extract documents from helm release manifest")
	}

	// Deploy individual resources in parallel
	g := new(errgroup.Group)
	for _, resource := range yamlResources {
		g.Go(func() error {
			tmpfile, err := os.CreateTemp("", "helm-manifest-*.yaml")
			if err != nil {
				return errors.Wrap(err, "failed to create a tmp helm manifest file")
			}
			defer func() {
				tmpfile.Close()
				os.Remove(tmpfile.Name())
			}()

			if _, err := tmpfile.Write(resource); err != nil {
				return errors.Wrap(err, "failed to write a tmp helm manifest file")
			}

			// get resource namespace, fallback to provided namespace if not explicit on resource
			resourceNamespace, err := kubernetes.GetNamespace(resource)
			if err != nil {
				return err
			}
			if resourceNamespace == "" {
				resourceNamespace = namespace
			}

			_, err = handler.kubernetesDeployer.Deploy(tokenData.ID, endpoint, []string{tmpfile.Name()}, resourceNamespace)

			return err
		})
	}

	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "unable to patch helm release using kubectl")
	}

	return nil
}
