// SPDX-FileCopyrightText: Copyright The Lima Authors
// SPDX-License-Identifier: Apache-2.0

package spiceservice

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// SpiceStatus represents the status of SPICE-related services
type SpiceStatus struct {
	AgentInstalled bool   // Whether spice-vdagent package is installed
	AgentRunning   bool   // Whether spice-vdagentd service is running
	VPortExists    bool   // Whether /dev/vport* exists (virtio console)
	ClipboardReady bool   // Whether clipboard sharing is functional
	ErrorMessage   string // Any error encountered
}

// DetectSpiceStatus checks the current SPICE configuration
func DetectSpiceStatus(ctx context.Context) *SpiceStatus {
	status := &SpiceStatus{}

	// Check if virtio console port exists
	status.VPortExists = checkVirtioPort()

	// Check if spice-vdagent is installed
	status.AgentInstalled = checkSpiceInstalled(ctx)

	// Check if spice-vdagentd service is running
	if status.AgentInstalled {
		status.AgentRunning = checkSpiceRunning(ctx)
	}

	// Clipboard is ready if all components are present
	status.ClipboardReady = status.VPortExists && status.AgentInstalled && status.AgentRunning

	// Generate error message if not ready
	if !status.ClipboardReady {
		status.ErrorMessage = buildErrorMessage(status)
	}

	return status
}

// EnsureSpiceAgent attempts to install and start spice-vdagent if needed
func EnsureSpiceAgent(ctx context.Context) error {
	status := DetectSpiceStatus(ctx)

	// If everything is ready, nothing to do
	if status.ClipboardReady {
		logrus.Info("SPICE agent already configured and running")
		return nil
	}

	// Can't proceed without virtio port
	if !status.VPortExists {
		logrus.Warn("SPICE virtio port not found - clipboard sharing requires VZ display configuration on host")
		return fmt.Errorf("virtio console port not available (host SPICE not configured)")
	}

	// Try to install spice-vdagent if not present
	if !status.AgentInstalled {
		logrus.Info("Installing spice-vdagent package...")
		if err := installSpiceAgent(ctx); err != nil {
			return fmt.Errorf("failed to install spice-vdagent: %w", err)
		}
		logrus.Info("spice-vdagent package installed successfully")
		status.AgentInstalled = true
	}

	// Try to start/enable the service if not running
	if !status.AgentRunning {
		logrus.Info("Starting spice-vdagentd service...")
		if err := startSpiceService(ctx); err != nil {
			return fmt.Errorf("failed to start spice-vdagentd: %w", err)
		}
		logrus.Info("spice-vdagentd service started successfully")
	}

	return nil
}

// checkVirtioPort checks if virtio console port device exists
func checkVirtioPort() bool {
	// Check for /dev/vport* devices
	matches, err := os.ReadDir("/dev")
	if err != nil {
		return false
	}

	for _, entry := range matches {
		if strings.HasPrefix(entry.Name(), "vport") {
			return true
		}
	}

	// Also check for virtio-ports directory
	if _, err := os.Stat("/sys/class/virtio-ports"); err == nil {
		entries, err := os.ReadDir("/sys/class/virtio-ports")
		if err == nil && len(entries) > 0 {
			return true
		}
	}

	return false
}

// checkSpiceInstalled checks if spice-vdagent package is installed
func checkSpiceInstalled(ctx context.Context) bool {
	// Try multiple methods to detect installation

	// Method 1: Check if binary exists
	if _, err := exec.LookPath("spice-vdagentd"); err == nil {
		return true
	}

	// Method 2: Try dpkg (Debian/Ubuntu)
	ctx2, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx2, "dpkg", "-l", "spice-vdagent")
	if err := cmd.Run(); err == nil {
		return true
	}

	// Method 3: Try rpm (RHEL/Fedora)
	ctx3, cancel3 := context.WithTimeout(ctx, 2*time.Second)
	defer cancel3()
	cmd3 := exec.CommandContext(ctx3, "rpm", "-q", "spice-vdagent")
	if err := cmd3.Run(); err == nil {
		return true
	}

	return false
}

// checkSpiceRunning checks if spice-vdagentd service is running
func checkSpiceRunning(ctx context.Context) bool {
	ctx2, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Check systemd service status
	cmd := exec.CommandContext(ctx2, "systemctl", "is-active", "spice-vdagentd")
	output, err := cmd.Output()
	if err == nil && strings.TrimSpace(string(output)) == "active" {
		return true
	}

	// Fallback: Check if process is running
	cmd2 := exec.CommandContext(ctx, "pgrep", "-x", "spice-vdagentd")
	if err := cmd2.Run(); err == nil {
		return true
	}

	return false
}

// installSpiceAgent attempts to install spice-vdagent package
func installSpiceAgent(ctx context.Context) error {
	// Try different package managers
	packageManagers := []struct {
		cmd     string
		args    []string
		pkgName string
	}{
		{"apt-get", []string{"install", "-y"}, "spice-vdagent"},
		{"dnf", []string{"install", "-y"}, "spice-vdagent"},
		{"yum", []string{"install", "-y"}, "spice-vdagent"},
		{"zypper", []string{"install", "-y"}, "spice-vdagent"},
		{"pacman", []string{"-S", "--noconfirm"}, "spice-vdagent"},
	}

	for _, pm := range packageManagers {
		if _, err := exec.LookPath(pm.cmd); err != nil {
			continue // Package manager not available
		}

		ctx2, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()

		args := append(pm.args, pm.pkgName)
		cmd := exec.CommandContext(ctx2, pm.cmd, args...)
		if output, err := cmd.CombinedOutput(); err != nil {
			logrus.Debugf("Failed to install with %s: %v (output: %s)", pm.cmd, err, string(output))
			continue
		}

		logrus.Infof("Successfully installed spice-vdagent using %s", pm.cmd)
		return nil
	}

	return fmt.Errorf("no supported package manager found or installation failed")
}

// startSpiceService attempts to start and enable spice-vdagentd service
func startSpiceService(ctx context.Context) error {
	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Enable the service to start on boot
	enableCmd := exec.CommandContext(ctx2, "systemctl", "enable", "spice-vdagentd")
	if output, err := enableCmd.CombinedOutput(); err != nil {
		logrus.Warnf("Failed to enable spice-vdagentd: %v (output: %s)", err, string(output))
	}

	// Start the service
	ctx3, cancel3 := context.WithTimeout(ctx, 10*time.Second)
	defer cancel3()
	startCmd := exec.CommandContext(ctx3, "systemctl", "start", "spice-vdagentd")
	if output, err := startCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start spice-vdagentd: %w (output: %s)", err, string(output))
	}

	// Verify it's running
	time.Sleep(500 * time.Millisecond)
	if !checkSpiceRunning(ctx) {
		return fmt.Errorf("service started but not running")
	}

	return nil
}

// buildErrorMessage creates a descriptive error message based on status
func buildErrorMessage(status *SpiceStatus) string {
	var reasons []string

	if !status.VPortExists {
		reasons = append(reasons, "virtio console port not found (host SPICE not configured)")
	}
	if !status.AgentInstalled {
		reasons = append(reasons, "spice-vdagent package not installed")
	}
	if status.AgentInstalled && !status.AgentRunning {
		reasons = append(reasons, "spice-vdagentd service not running")
	}

	if len(reasons) == 0 {
		return ""
	}

	return fmt.Sprintf("SPICE clipboard not ready: %s", strings.Join(reasons, "; "))
}
