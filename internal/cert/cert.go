package cert

import (
	"crypto/tls"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	gotunnelErrors "github.com/johncferguson/gotunnel/internal/errors"
)

func isMkcertInstalled() bool {
	_, err := exec.LookPath("mkcert")
	return err == nil
}

type CertManager struct {
	certsDir string
}

func New(certsDir string) *CertManager {
	return &CertManager{
		certsDir: certsDir,
	}
}

func getCurrentUser() (*user.User, error) {
	return user.Current()
}

func (m *CertManager) IsMkcertAvailable() bool {
	return isMkcertInstalled()
}

func (m *CertManager) EnsureMkcertInstalled() error {
	// Check if mkcert is already installed
	if isMkcertInstalled() {
		return nil
	}

	// Install mkcert based on the platform
	var installCmd string
	if runtime.GOOS == "windows" {
		installCmd = "wsl nix-env -iA nixpkgs.mkcert"
	} else {
		installCmd = "nix-env -iA nixpkgs.mkcert"
	}

	cmdParts := strings.Fields(installCmd)
	if err := runAsUser(cmdParts[0], cmdParts[1:]...); err != nil {
		return gotunnelErrors.CertificateError("install", "", err)
	}

	return nil
}

func (m *CertManager) EnsureCert(domain string) (*tls.Certificate, error) {
	if err := os.MkdirAll(m.certsDir, 0755); err != nil {
		return nil, gotunnelErrors.Wrap(err, gotunnelErrors.ErrCodeFilesystem, "Failed to create certs directory")
	}

	certFile := filepath.Join(m.certsDir, domain+".pem")
	keyFile := filepath.Join(m.certsDir, domain+"-key.pem")

	// Check if certificate already exists
	if _, err := os.Stat(certFile); err == nil {
		if _, err := os.Stat(keyFile); err == nil {
			// Both files exist, load and return the certificate
			cert, err := tls.LoadX509KeyPair(certFile, keyFile)
			if err != nil {
				return nil, gotunnelErrors.CertificateError("load", domain, err)
			}
			return &cert, nil
		}
	}

	// Generate new certificate
	if err := runAsUser("mkcert", "-cert-file", certFile, "-key-file", keyFile, domain); err != nil {
		return nil, gotunnelErrors.CertificateError("generate", domain, err)
	}

	// Load and return the new certificate
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, gotunnelErrors.CertificateError("load", domain, err)
	}

	return &cert, nil
}
