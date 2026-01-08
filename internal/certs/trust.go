package certs

import (
	"fmt"
	"os/exec"
	"runtime"
)

// InstallCA installs the CA certificate into the system trust store
func InstallCA() error {
	certPath, err := GetCACertPath()
	if err != nil {
		return err
	}

	switch runtime.GOOS {
	case "darwin":
		return installCADarwin(certPath)
	case "linux":
		return installCALinux(certPath)
	case "windows":
		return installCAWindows(certPath)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// UninstallCA removes the CA certificate from the system trust store
func UninstallCA() error {
	switch runtime.GOOS {
	case "darwin":
		return uninstallCADarwin()
	case "linux":
		return uninstallCALinux()
	case "windows":
		return uninstallCAWindows()
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// installCADarwin installs the CA into the macOS login keychain
func installCADarwin(certPath string) error {
	// Add to login keychain with trust settings for SSL
	cmd := exec.Command("security", "add-trusted-cert",
		"-r", "trustRoot",
		"-k", "login.keychain",
		certPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install CA on macOS: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// uninstallCADarwin removes the CA from the macOS login keychain
func uninstallCADarwin() error {
	cmd := exec.Command("security", "delete-certificate",
		"-c", "applink Local CA",
		"login.keychain",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to uninstall CA on macOS: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// installCALinux installs the CA into the system trust store
// This works for Debian/Ubuntu-based systems. Other distros may vary.
func installCALinux(certPath string) error {
	// Try the Debian/Ubuntu approach first
	destPath := "/usr/local/share/ca-certificates/applink-ca.crt"

	// Copy certificate (requires sudo)
	cpCmd := exec.Command("sudo", "cp", certPath, destPath)
	if output, err := cpCmd.CombinedOutput(); err != nil {
		// Try the RHEL/Fedora approach
		destPath = "/etc/pki/ca-trust/source/anchors/applink-ca.crt"
		cpCmd = exec.Command("sudo", "cp", certPath, destPath)
		if output2, err := cpCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to copy CA certificate: %w\nOutput: %s", err, string(output2))
		}
		_ = output // Debian/Ubuntu attempt failed, but RHEL/Fedora worked

		// Update trust on RHEL/Fedora
		updateCmd := exec.Command("sudo", "update-ca-trust")
		if output, err := updateCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to update CA trust: %w\nOutput: %s", err, string(output))
		}
		return nil
	}

	// Update certificates on Debian/Ubuntu
	updateCmd := exec.Command("sudo", "update-ca-certificates")
	if output, err := updateCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to update CA certificates: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// uninstallCALinux removes the CA from the system trust store
func uninstallCALinux() error {
	// Try Debian/Ubuntu path first
	rmCmd := exec.Command("sudo", "rm", "-f", "/usr/local/share/ca-certificates/applink-ca.crt")
	rmCmd.Run() // Ignore errors

	// Try RHEL/Fedora path
	rmCmd = exec.Command("sudo", "rm", "-f", "/etc/pki/ca-trust/source/anchors/applink-ca.crt")
	rmCmd.Run() // Ignore errors

	// Update on both systems
	exec.Command("sudo", "update-ca-certificates").Run()
	exec.Command("sudo", "update-ca-trust").Run()

	return nil
}

// installCAWindows installs the CA into the Windows user certificate store
func installCAWindows(certPath string) error {
	cmd := exec.Command("certutil", "-addstore", "-user", "Root", certPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install CA on Windows: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// uninstallCAWindows removes the CA from the Windows user certificate store
func uninstallCAWindows() error {
	// Find and delete by common name
	cmd := exec.Command("certutil", "-delstore", "-user", "Root", "applink Local CA")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to uninstall CA on Windows: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// IsCAInstalled checks if the CA is installed in the system trust store
func IsCAInstalled() bool {
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("security", "find-certificate", "-c", "applink Local CA", "login.keychain")
		return cmd.Run() == nil
	case "linux":
		// Check common locations
		cmd := exec.Command("test", "-f", "/usr/local/share/ca-certificates/applink-ca.crt")
		if cmd.Run() == nil {
			return true
		}
		cmd = exec.Command("test", "-f", "/etc/pki/ca-trust/source/anchors/applink-ca.crt")
		return cmd.Run() == nil
	case "windows":
		cmd := exec.Command("certutil", "-verifystore", "-user", "Root", "applink Local CA")
		return cmd.Run() == nil
	default:
		return false
	}
}
