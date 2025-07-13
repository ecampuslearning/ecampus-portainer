package sdk

import (
	"testing"

	"github.com/portainer/portainer/pkg/libhelm/options"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func Test_InitActionConfig(t *testing.T) {
	is := assert.New(t)
	hspm := NewHelmSDKPackageManager()

	t.Run("with nil k8sAccess should use default kubeconfig", func(t *testing.T) {
		actionConfig := new(action.Configuration)
		err := hspm.(*HelmSDKPackageManager).initActionConfig(actionConfig, "default", nil)

		// The function should not fail by design, even when not running in a k8s environment
		is.NoError(err, "should not return error when not in k8s environment")
	})

	t.Run("with k8sAccess should create in-memory config", func(t *testing.T) {
		actionConfig := new(action.Configuration)
		k8sAccess := &options.KubernetesClusterAccess{
			ClusterServerURL: "https://kubernetes.default.svc",
			AuthToken:        "test-token",
		}

		// The function should not fail by design
		err := hspm.(*HelmSDKPackageManager).initActionConfig(actionConfig, "default", k8sAccess)
		is.NoError(err, "should not return error when using in-memory config")
	})

	t.Run("with k8sAccess and CA file should create config with CA", func(t *testing.T) {
		actionConfig := new(action.Configuration)
		k8sAccess := &options.KubernetesClusterAccess{
			ClusterServerURL:         "https://kubernetes.default.svc",
			AuthToken:                "test-token",
			CertificateAuthorityFile: "/path/to/ca.crt",
		}

		// The function should not fail by design
		err := hspm.(*HelmSDKPackageManager).initActionConfig(actionConfig, "default", k8sAccess)
		is.NoError(err, "should not return error when using in-memory config with CA")
	})
}

func Test_ClientConfigGetter(t *testing.T) {
	is := assert.New(t)

	// Create a mock client config
	configAPI := api.NewConfig()

	// Create cluster
	cluster := api.NewCluster()
	cluster.Server = "https://kubernetes.default.svc"
	cluster.InsecureSkipTLSVerify = true

	// Create auth info
	authInfo := api.NewAuthInfo()
	authInfo.Token = "test-token"

	// Create context
	context := api.NewContext()
	context.Cluster = "test-cluster"
	context.AuthInfo = "test-user"
	context.Namespace = "default"

	// Add to config
	configAPI.Clusters["test-cluster"] = cluster
	configAPI.AuthInfos["test-user"] = authInfo
	configAPI.Contexts["test-context"] = context
	configAPI.CurrentContext = "test-context"

	clientConfig := clientcmd.NewDefaultClientConfig(*configAPI, &clientcmd.ConfigOverrides{})

	// Create client config getter
	clientGetter, err := newRESTClientGetter(clientConfig, "default")
	is.NoError(err, "should not return error when creating client getter")

	// Test ToRESTConfig
	restConfig, err := clientGetter.ToRESTConfig()
	is.NoError(err, "should not return error when creating REST config")
	is.NotNil(restConfig, "should return non-nil REST config")
	is.Equal("https://kubernetes.default.svc", restConfig.Host, "host should be https://kubernetes.default.svc")
	is.Equal("test-token", restConfig.BearerToken, "bearer token should be test-token")

	// Test ToDiscoveryClient
	discoveryClient, err := clientGetter.ToDiscoveryClient()
	is.NoError(err, "should not return error when creating discovery client")
	is.NotNil(discoveryClient, "should return non-nil discovery client")

	// Test ToRESTMapper
	restMapper, err := clientGetter.ToRESTMapper()
	is.NoError(err, "should not return error when creating REST mapper")
	is.NotNil(restMapper, "should return non-nil REST mapper")

	// Test ToRawKubeConfigLoader
	config := clientGetter.ToRawKubeConfigLoader()
	is.NotNil(config, "should return non-nil config loader")
}

func Test_ParseValues(t *testing.T) {
	is := assert.New(t)
	hspm := NewHelmSDKPackageManager()

	t.Run("should parse valid YAML values", func(t *testing.T) {
		yamlData := []byte(`
service:
  type: ClusterIP
  port: 80
resources:
  limits:
    cpu: 100m
    memory: 128Mi
`)
		values, err := hspm.(*HelmSDKPackageManager).parseValues(yamlData)
		is.NoError(err, "should parse valid YAML without error")
		is.NotNil(values, "should return non-nil values")

		// Verify structure
		service, ok := values["service"].(map[string]interface{})
		is.True(ok, "service should be a map")
		is.Equal("ClusterIP", service["type"], "service type should be ClusterIP")
		is.Equal(float64(80), service["port"], "service port should be 80")

		resources, ok := values["resources"].(map[string]interface{})
		is.True(ok, "resources should be a map")
		limits, ok := resources["limits"].(map[string]interface{})
		is.True(ok, "limits should be a map")
		is.Equal("100m", limits["cpu"], "cpu limit should be 100m")
		is.Equal("128Mi", limits["memory"], "memory limit should be 128Mi")
	})

	t.Run("should handle invalid YAML", func(t *testing.T) {
		yamlData := []byte(`
service:
  type: ClusterIP
  port: 80
  invalid yaml
`)
		_, err := hspm.(*HelmSDKPackageManager).parseValues(yamlData)
		is.Error(err, "should return error for invalid YAML")
	})

	t.Run("should handle empty YAML", func(t *testing.T) {
		yamlData := []byte(``)
		values, err := hspm.(*HelmSDKPackageManager).parseValues(yamlData)
		is.NoError(err, "should not return error for empty YAML")
		is.NotNil(values, "should return non-nil values for empty YAML")
		is.Len(values, 0, "should return empty map for empty YAML")
	})
}
