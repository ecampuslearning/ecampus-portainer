package sdk

import (
	"os"
	"testing"

	"github.com/portainer/portainer/pkg/libhelm/options"
	"github.com/portainer/portainer/pkg/libhelm/test"
	"github.com/stretchr/testify/assert"
)

func createValuesFile(values string) (string, error) {
	file, err := os.CreateTemp("", "helm-values")
	if err != nil {
		return "", err
	}

	_, err = file.WriteString(values)
	if err != nil {
		file.Close()
		return "", err
	}

	err = file.Close()
	if err != nil {
		return "", err
	}

	return file.Name(), nil
}

func Test_Install(t *testing.T) {
	test.EnsureIntegrationTest(t)
	is := assert.New(t)

	// Create a new SDK package manager
	hspm := NewHelmSDKPackageManager()

	t.Run("successfully installs nginx chart with name test-nginx", func(t *testing.T) {
		// SDK equivalent of: helm install test-nginx --repo https://kubernetes.github.io/ingress-nginx nginx
		installOpts := options.InstallOptions{
			Name:  "test-nginx",
			Chart: "ingress-nginx",
			Repo:  "https://kubernetes.github.io/ingress-nginx",
		}

		release, err := hspm.Install(installOpts)
		if release != nil {
			defer hspm.Uninstall(options.UninstallOptions{
				Name: "test-nginx",
			})
		}

		is.NoError(err, "should successfully install release")
		is.NotNil(release, "should return non-nil release")
		is.Equal("test-nginx", release.Name, "release name should match")
		is.Equal(1, release.Version, "release version should be 1")
		is.NotEmpty(release.Manifest, "release manifest should not be empty")
	})

	t.Run("successfully installs nginx with values", func(t *testing.T) {
		// SDK equivalent of: helm install test-nginx-2 --repo https://kubernetes.github.io/ingress-nginx nginx --values /tmp/helm-values3161785816
		values, err := createValuesFile("service:\n  port:  8081")
		is.NoError(err, "should create a values file")

		defer os.Remove(values)

		installOpts := options.InstallOptions{
			Name:       "test-nginx-2",
			Chart:      "ingress-nginx",
			Repo:       "https://kubernetes.github.io/ingress-nginx",
			ValuesFile: values,
		}
		release, err := hspm.Install(installOpts)
		if release != nil {
			defer hspm.Uninstall(options.UninstallOptions{
				Name: "test-nginx-2",
			})
		}

		is.NoError(err, "should successfully install release")
		is.NotNil(release, "should return non-nil release")
		is.Equal("test-nginx-2", release.Name, "release name should match")
		is.Equal(1, release.Version, "release version should be 1")
		is.NotEmpty(release.Manifest, "release manifest should not be empty")
	})

	t.Run("successfully installs portainer chart with name portainer-test", func(t *testing.T) {
		// SDK equivalent of: helm install portainer-test portainer --repo https://portainer.github.io/k8s/
		installOpts := options.InstallOptions{
			Name:  "portainer-test",
			Chart: "portainer",
			Repo:  "https://portainer.github.io/k8s/",
		}
		release, err := hspm.Install(installOpts)
		if release != nil {
			defer hspm.Uninstall(options.UninstallOptions{
				Name: installOpts.Name,
			})
		}

		is.NoError(err, "should successfully install release")
		is.NotNil(release, "should return non-nil release")
		is.Equal("portainer-test", release.Name, "release name should match")
		is.Equal(1, release.Version, "release version should be 1")
		is.NotEmpty(release.Manifest, "release manifest should not be empty")
	})

	t.Run("install with values as string", func(t *testing.T) {
		// First create a values file since InstallOptions doesn't support values as string directly
		values, err := createValuesFile("service:\n  port:  8082")
		is.NoError(err, "should create a values file")

		defer os.Remove(values)

		// Install with values file
		installOpts := options.InstallOptions{
			Name:       "test-nginx-3",
			Chart:      "ingress-nginx",
			Repo:       "https://kubernetes.github.io/ingress-nginx",
			ValuesFile: values,
		}
		release, err := hspm.Install(installOpts)
		if release != nil {
			defer hspm.Uninstall(options.UninstallOptions{
				Name: "test-nginx-3",
			})
		}

		is.NoError(err, "should successfully install release")
		is.NotNil(release, "should return non-nil release")
		is.Equal("test-nginx-3", release.Name, "release name should match")
	})

	t.Run("install with namespace", func(t *testing.T) {
		// Install with namespace
		installOpts := options.InstallOptions{
			Name:      "test-nginx-4",
			Chart:     "ingress-nginx",
			Repo:      "https://kubernetes.github.io/ingress-nginx",
			Namespace: "default",
		}
		release, err := hspm.Install(installOpts)
		if release != nil {
			defer hspm.Uninstall(options.UninstallOptions{
				Name:      "test-nginx-4",
				Namespace: "default",
			})
		}

		is.NoError(err, "should successfully install release")
		is.NotNil(release, "should return non-nil release")
		is.Equal("test-nginx-4", release.Name, "release name should match")
		is.Equal("default", release.Namespace, "release namespace should match")
	})

	t.Run("returns an error when name is not provided", func(t *testing.T) {
		installOpts := options.InstallOptions{
			Chart: "ingress-nginx",
			Repo:  "https://kubernetes.github.io/ingress-nginx",
		}
		_, err := hspm.Install(installOpts)

		is.Error(err, "should return an error when name is not provided")
		is.Equal(err.Error(), "name is required")
	})

	t.Run("install with invalid chart", func(t *testing.T) {
		// Install with invalid chart
		installOpts := options.InstallOptions{
			Name:  "test-invalid",
			Chart: "non-existent-chart",
			Repo:  "https://kubernetes.github.io/ingress-nginx",
		}
		_, err := hspm.Install(installOpts)
		is.Error(err, "should return error when chart doesn't exist")
		is.Equal(err.Error(), "failed to find the helm chart at the path: https://kubernetes.github.io/ingress-nginx/non-existent-chart")
	})

	t.Run("install with invalid repo", func(t *testing.T) {
		// Install with invalid repo
		installOpts := options.InstallOptions{
			Name:  "test-invalid-repo",
			Chart: "nginx",
			Repo:  "https://non-existent-repo.example.com",
		}
		_, err := hspm.Install(installOpts)
		is.Error(err, "should return error when repo doesn't exist")
	})
}
