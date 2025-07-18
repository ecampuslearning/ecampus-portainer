package helm

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/datastore"
	"github.com/portainer/portainer/api/exec/exectest"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/internal/testhelpers"
	helper "github.com/portainer/portainer/api/internal/testhelpers"
	"github.com/portainer/portainer/api/jwt"
	"github.com/portainer/portainer/api/kubernetes"
	"github.com/portainer/portainer/pkg/libhelm/options"
	"github.com/portainer/portainer/pkg/libhelm/release"
	"github.com/portainer/portainer/pkg/libhelm/test"

	"github.com/segmentio/encoding/json"
	"github.com/stretchr/testify/assert"
)

func Test_helmList(t *testing.T) {
	is := assert.New(t)

	_, store := datastore.MustNewTestStore(t, true, true)

	err := store.Endpoint().Create(&portainer.Endpoint{ID: 1})
	assert.NoError(t, err, "error creating environment")

	err = store.User().Create(&portainer.User{Username: "admin", Role: portainer.AdministratorRole})
	assert.NoError(t, err, "error creating a user")

	jwtService, err := jwt.NewService("1h", store)
	is.NoError(err, "Error initialising jwt service")

	kubernetesDeployer := exectest.NewKubernetesDeployer()
	helmPackageManager := test.NewMockHelmPackageManager()
	kubeClusterAccessService := kubernetes.NewKubeClusterAccessService("", "", "")
	h := NewHandler(helper.NewTestRequestBouncer(), store, jwtService, kubernetesDeployer, helmPackageManager, kubeClusterAccessService)

	// Install a single chart.  We expect to get these values back
	options := options.InstallOptions{Name: "nginx-1", Chart: "nginx", Namespace: "default"}
	h.helmPackageManager.Upgrade(options)

	t.Run("helmList", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/1/kubernetes/helm", nil)
		ctx := security.StoreTokenData(req, &portainer.TokenData{ID: 1, Username: "admin", Role: 1})
		req = req.WithContext(ctx)
		testhelpers.AddTestSecurityCookie(req, "Bearer dummytoken")

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusOK, rr.Code, "Status should be 200 OK")

		body, err := io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")

		data := []release.ReleaseElement{}
		json.Unmarshal(body, &data)
		if is.Equal(1, len(data), "Expected one chart entry") {
			is.EqualValues(options.Name, data[0].Name, "Name doesn't match")
			is.EqualValues(options.Chart, data[0].Chart, "Chart doesn't match")
		}
	})
}
