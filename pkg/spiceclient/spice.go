package spiceclient

// SPDX-FileCopyrightText: Copyright The Lima Authors
// SPDX-License-Identifier: Apache-2.0

// Package spiceclient provides SPICE client connection functionality for Lima VMs.
// This package integrates with SPICE-enabled QEMU instances to provide graphical
// display capabilities using the SPICE protocol.

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

// Connection represents a SPICE connection configuration
type Connection struct {
	Host     string
	Port     string
	Password string
	UnixPath string // For Unix socket connections
	Audio    bool   // Enable audio streaming
}

// LaunchViewer launches an external SPICE viewer application with the given connection details.
// It attempts to find and use available SPICE client applications on the system.
func LaunchViewer(ctx context.Context, conn *Connection) error {
	viewer, err := FindViewer()
	if err != nil {
		return fmt.Errorf("failed to find SPICE viewer: %w", err)
	}

	args, err := buildViewerArgs(viewer, conn)
	if err != nil {
		return fmt.Errorf("failed to build viewer arguments: %w", err)
	}

	cmd := exec.CommandContext(ctx, viewer, args...)

	logrus.Debugf("Launching SPICE viewer: %s %v", viewer, args)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start SPICE viewer: %w", err)
	}

	// Don't wait for the viewer to exit, let it run independently
	go func() {
		if err := cmd.Wait(); err != nil {
			logrus.Debugf("SPICE viewer exited with error: %v", err)
		}
	}()

	return nil
}

// FindViewer attempts to locate an available SPICE viewer on the system.
// It searches for common SPICE client applications in order of preference.
func FindViewer() (string, error) {
	var candidates []string

	switch runtime.GOOS {
	case "darwin":
		// macOS: Check for various SPICE clients
		candidates = []string{
			"remote-viewer", // virt-viewer package
			"spicy",         // spice-gtk
		}
	case "linux":
		candidates = []string{
			"remote-viewer", // Most common on Linux
			"spicy",         // spice-gtk
			"virt-viewer",
		}
	case "windows":
		candidates = []string{
			"remote-viewer.exe",
			"spicy.exe",
			"virt-viewer.exe",
		}
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	for _, viewer := range candidates {
		path, err := exec.LookPath(viewer)
		if err == nil {
			logrus.Debugf("Found SPICE viewer: %s", path)
			return path, nil
		}
	}

	return "", fmt.Errorf("no SPICE viewer found, install remote-viewer or spicy")
}

// buildViewerArgs constructs command-line arguments for the SPICE viewer based on the connection details.
func buildViewerArgs(viewer string, conn *Connection) ([]string, error) {
	var args []string

	// Determine viewer type from the executable name
	viewerName := strings.ToLower(viewer)

	if strings.Contains(viewerName, "remote-viewer") || strings.Contains(viewerName, "virt-viewer") {
		// remote-viewer and virt-viewer use SPICE URI format
		uri, err := buildSpiceURI(conn)
		if err != nil {
			return nil, err
		}
		args = []string{uri}

		// Add fullscreen option
		args = append(args, "--full-screen")

		// Disable audio if not enabled
		if !conn.Audio {
			args = append(args, "--spice-disable-audio")
		}

	} else if strings.Contains(viewerName, "spicy") {
		// spicy uses separate host/port arguments
		if conn.UnixPath != "" {
			return nil, fmt.Errorf("spicy does not support Unix socket connections")
		}

		args = []string{
			"-h", conn.Host,
			"-p", conn.Port,
		}

		if conn.Password != "" {
			args = append(args, "-w", conn.Password)
		}
	} else {
		return nil, fmt.Errorf("unknown SPICE viewer type: %s", viewer)
	}

	return args, nil
}

// buildSpiceURI constructs a SPICE connection URI from the connection details.
// Supports both TCP and Unix socket connections.
func buildSpiceURI(conn *Connection) (string, error) {
	if conn.UnixPath != "" {
		return fmt.Sprintf("spice+unix://%s", conn.UnixPath), nil
	}

	if conn.Host == "" || conn.Port == "" {
		return "", fmt.Errorf("host and port required for TCP connection")
	}

	uri := fmt.Sprintf("spice://%s:%s", conn.Host, conn.Port)

	if conn.Password != "" {
		uri += fmt.Sprintf("?password=%s", conn.Password)
	}

	return uri, nil
}

// GetConnectionInfo extracts SPICE connection information from a QEMU SPICE display string.
// Example inputs: "spice,port=5900,disable-ticketing=on" or "spice+unix:///path/to/socket"
func GetConnectionInfo(displayString string) (*Connection, error) {
	conn := &Connection{}

	// Check for Unix socket format
	if strings.HasPrefix(displayString, "spice+unix://") {
		conn.UnixPath = strings.TrimPrefix(displayString, "spice+unix://")
		return conn, nil
	}

	// Parse TCP format: "spice,port=5900,addr=127.0.0.1,disable-ticketing=on"
	if !strings.HasPrefix(displayString, "spice") {
		return nil, fmt.Errorf("invalid SPICE display string: %s", displayString)
	}

	// Set defaults
	conn.Host = "127.0.0.1"
	conn.Port = "5900"

	// Parse comma-separated options
	parts := strings.Split(displayString, ",")
	for _, part := range parts {
		if part == "spice" {
			continue
		}

		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "port":
			conn.Port = value
		case "addr":
			conn.Host = value
		case "password":
			conn.Password = value
		}
	}

	return conn, nil
}

// QuerySPICEPort queries QEMU via QMP to get the SPICE port information.
// Returns the SPICE service string (e.g., "127.0.0.1:5900").
func QuerySPICEPort(qmpSocketPath string) (string, error) {
	// Connect to QMP socket
	conn, err := net.Dial("unix", qmpSocketPath)
	if err != nil {
		return "", fmt.Errorf("failed to connect to QMP socket: %w", err)
	}
	defer conn.Close()

	// This is a simplified implementation
	// In a full implementation, you would:
	// 1. Perform QMP handshake
	// 2. Send query-spice command
	// 3. Parse the JSON response
	// For now, return an error indicating this needs QMP integration

	return "", fmt.Errorf("QMP SPICE query not yet implemented, use display configuration")
}
