package sdk

import (
	"testing"
	"time"

	"github.com/portainer/portainer/pkg/libhelm/types"
	"github.com/stretchr/testify/assert"
)

func Test_NewHelmSDKPackageManager(t *testing.T) {
	is := assert.New(t)

	// Test that NewHelmSDKPackageManager returns a non-nil HelmPackageManager
	manager := NewHelmSDKPackageManager()
	is.NotNil(manager, "should return non-nil HelmPackageManager")

	// Test that the returned manager is of the correct type
	_, ok := manager.(*HelmSDKPackageManager)
	is.True(ok, "should return a *HelmSDKPackageManager")

	// Test that the manager has the expected fields
	sdkManager := manager.(*HelmSDKPackageManager)
	is.NotNil(sdkManager.settings, "should have non-nil settings")
	is.Equal(300*time.Second, sdkManager.timeout, "should have 5 minute timeout")

	// Test that the manager implements the HelmPackageManager interface
	var _ types.HelmPackageManager = manager
}
