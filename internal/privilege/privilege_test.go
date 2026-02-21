package privilege

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckPrivileges(t *testing.T) {
	err := CheckPrivileges()
	// CheckPrivileges now actually checks for required privileges
	// It should return an error when running without sufficient privileges
	if HasRootPrivileges() {
		// If running as root/admin, it should pass
		assert.NoError(t, err)
	} else {
		// If not running with privileges, it should return an error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient privileges")
	}
}

func TestHasRootPrivileges(t *testing.T) {
	isRoot := HasRootPrivileges()

	if runtime.GOOS == "windows" {
		// On Windows, check if running as admin
		// The actual result depends on how the test is run
		t.Logf("Running with admin privileges: %v", isRoot)
	} else {
		// On Unix systems, we can check the effective user ID
		expectedRoot := os.Geteuid() == 0
		assert.Equal(t, expectedRoot, isRoot)
	}
}

func TestCheckUnixPrivileges(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix privilege test on Windows")
	}

	err := checkUnixPrivileges()
	// checkUnixPrivileges now actually checks for required privileges
	if os.Geteuid() == 0 {
		// If running as root, it should pass
		assert.NoError(t, err)
	} else {
		// If not running as root, it should return an error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient privileges")
	}
}

func TestCheckWindowsPrivileges(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows privilege test on non-Windows platform")
	}

	err := checkWindowsPrivileges()
	// checkWindowsPrivileges now actually checks for admin privileges
	if hasWindowsAdminPrivileges() {
		// If running as admin, it should pass
		assert.NoError(t, err)
	} else {
		// If not running as admin, it should return an error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient privileges")
	}
}

func TestElevatePrivileges(t *testing.T) {
	// Skip if already running with privileges
	if HasRootPrivileges() {
		t.Skip("Skipping elevation test as already running with privileges")
	}

	// This is a potentially dangerous test as it might trigger UAC/sudo
	// Only run it in specific test environments
	if os.Getenv("GOTUNNEL_TEST_ELEVATION") != "1" {
		t.Skip("Skipping elevation test. Set GOTUNNEL_TEST_ELEVATION=1 to run")
	}

	err := ElevatePrivileges()
	// Just verify the function runs without panicking
	// The actual result depends on the environment
	t.Logf("Privilege elevation attempt result: %v", err)
}

func TestPrivilegeChecksIntegration(t *testing.T) {
	// Integration test combining multiple privilege checks
	isRoot := HasRootPrivileges()
	err := CheckPrivileges()

	// CheckPrivileges now actually checks for privileges
	if isRoot {
		// If running with privileges, CheckPrivileges should pass
		assert.NoError(t, err, "Privilege check should pass when running with sufficient privileges")
	} else {
		// If not running with privileges, CheckPrivileges should fail
		assert.Error(t, err, "Privilege check should fail when running without sufficient privileges")
		assert.Contains(t, err.Error(), "insufficient privileges")
	}

	t.Logf("Has root privileges: %v", isRoot)
}
