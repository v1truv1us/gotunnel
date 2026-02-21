package privilege

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"

	"github.com/johncferguson/gotunnel/internal/logging"
)

func CheckPrivileges() error {
	logger, _ := logging.New(logging.DefaultConfig())
	logger.Debug("Checking privileges...")
	switch runtime.GOOS {
	case "windows":
		return checkWindowsPrivileges()
	default: // Linux, macOS, BSD, etc.
		return checkUnixPrivileges()
	}
}

func checkUnixPrivileges() error {
	// Check if we have the privileges needed for tunneling
	// We need privileges to bind to ports < 1024 and modify /etc/hosts
	if os.Geteuid() == 0 {
		return nil // Already running as root
	}

	// Check if we can bind to port 80 (requires root on most systems)
	conn, err := net.Listen("tcp", "127.0.0.1:80")
	if err != nil {
		return fmt.Errorf("insufficient privileges: cannot bind to port 80. Run with sudo or use --no-privilege-check to skip this check")
	}
	conn.Close()

	// Check if we can modify /etc/hosts
	testFile := "/etc/hosts.gotunnel_test"
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("insufficient privileges: cannot modify /etc/hosts. Run with sudo or use --no-privilege-check to skip this check")
	}
	os.Remove(testFile)

	return nil
}

func checkWindowsPrivileges() error {
	// Check if we have administrator privileges
	if hasWindowsAdminPrivileges() {
		return nil // Already running as admin
	}

	// Try to bind to port 80 (requires admin on Windows)
	conn, err := net.Listen("tcp", "127.0.0.1:80")
	if err != nil {
		return fmt.Errorf("insufficient privileges: cannot bind to port 80. Run as administrator or use --no-privilege-check to skip this check")
	}
	conn.Close()

	// Try to modify hosts file
	testFile := "C:\\Windows\\System32\\drivers\\etc\\hosts.gotunnel_test"
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("insufficient privileges: cannot modify hosts file. Run as administrator or use --no-privilege-check to skip this check")
	}
	os.Remove(testFile)

	return nil
}

func ElevatePrivileges() error {
		return ElevatePrivilegesWithLogger(nil)
}

func ElevatePrivilegesWithLogger(logger *logging.Logger) error {
		if logger == nil {
			logger, _ = logging.New(logging.DefaultConfig())
		}
		
		logger.Info("Attempting to elevate privileges...")
		if runtime.GOOS == "windows" {
			return elevateWindows()
		}
		return elevateSudo()
	}

func elevateWindows() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	cmd := exec.Command("powershell.exe", "Start-Process", exe, "-Verb", "RunAs")
	return cmd.Run()
}

func elevateSudo() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	cmd := exec.Command("sudo", exe)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// HasRootPrivileges checks if the current process has root privileges
func HasRootPrivileges() bool {
	if runtime.GOOS == "windows" {
		return hasWindowsAdminPrivileges()
	}
	return os.Geteuid() == 0
}

// hasWindowsAdminPrivileges checks if running as admin on Windows
func hasWindowsAdminPrivileges() bool {
	// On Windows, we'll check if we can write to a system directory
	// This is a simple heuristic, not perfect but works for most cases
	_, err := os.Stat("C:\\Windows\\System32")
	if err != nil {
		return false
	}
	
	// Try to create a temp file in system32 (will fail if not admin)
	testFile := "C:\\Windows\\System32\\gotunnel_admin_test.tmp"
	file, err := os.Create(testFile)
	if err != nil {
		return false
	}
	file.Close()
	os.Remove(testFile) // Clean up
	return true
}
