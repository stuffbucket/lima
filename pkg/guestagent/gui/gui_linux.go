// SPDX-FileCopyrightText: Copyright The Lima Authors
// SPDX-License-Identifier: Apache-2.0

package gui

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/lima-vm/lima/v2/pkg/guestagent/api"
	"github.com/lima-vm/lima/v2/pkg/guestagent/spiceservice"
)

// DetectGUIInfo detects GUI-related information from the Linux guest
func DetectGUIInfo(ctx context.Context) *api.GUIInfo {
	info := &api.GUIInfo{
		DisplayServer: "none",
		SessionActive: false,
	}

	// Detect display server type
	if detectWayland() {
		info.DisplayServer = "Wayland"
		info.Displays = getWaylandDisplays()
	} else if detectX11() {
		info.DisplayServer = "X11"
		info.Displays = getX11Displays()
	}

	// Check if any GUI session is active
	info.SessionActive = len(info.Displays) > 0

	// Get resolution if available
	if info.SessionActive {
		info.Resolution = getResolution(info.DisplayServer)
	}

	// Get idle time
	if info.SessionActive {
		info.IdleTimeMs = getIdleTime(info.DisplayServer)
	}

	// Detect SPICE agent status for clipboard sharing
	spiceStatus := spiceservice.DetectSpiceStatus(ctx)
	info.Spice = &api.SpiceAgentInfo{
		AgentInstalled: spiceStatus.AgentInstalled,
		AgentRunning:   spiceStatus.AgentRunning,
		VportExists:    spiceStatus.VPortExists,
		ClipboardReady: spiceStatus.ClipboardReady,
		ErrorMessage:   spiceStatus.ErrorMessage,
	}

	// If SPICE port exists but agent isn't running, try to auto-enable it
	if spiceStatus.VPortExists && !spiceStatus.ClipboardReady {
		logrus.Info("SPICE virtio port detected, attempting to enable clipboard sharing...")
		if err := spiceservice.EnsureSpiceAgent(ctx); err != nil {
			logrus.Warnf("Failed to auto-enable SPICE agent: %v", err)
		} else {
			// Re-detect status after enabling
			spiceStatus = spiceservice.DetectSpiceStatus(ctx)
			info.Spice.AgentInstalled = spiceStatus.AgentInstalled
			info.Spice.AgentRunning = spiceStatus.AgentRunning
			info.Spice.ClipboardReady = spiceStatus.ClipboardReady
			info.Spice.ErrorMessage = spiceStatus.ErrorMessage
		}
	}

	return info
}

// detectX11 checks if X11 is running
func detectX11() bool {
	// Check for common X11 sockets
	if _, err := os.Stat("/tmp/.X11-unix"); err == nil {
		return true
	}
	// Check if DISPLAY is set
	if os.Getenv("DISPLAY") != "" {
		return true
	}
	return false
}

// detectWayland checks if Wayland is running
func detectWayland() bool {
	// Check for Wayland socket
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		return true
	}
	// Check XDG_SESSION_TYPE
	if os.Getenv("XDG_SESSION_TYPE") == "wayland" {
		return true
	}
	return false
}

// getX11Displays returns list of active X11 displays
func getX11Displays() []string {
	displays := []string{}

	// Check DISPLAY environment variable
	if display := os.Getenv("DISPLAY"); display != "" {
		displays = append(displays, display)
		return displays
	}

	// Check /tmp/.X11-unix for active displays
	entries, err := os.ReadDir("/tmp/.X11-unix")
	if err != nil {
		return displays
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "X") {
			displayNum := strings.TrimPrefix(entry.Name(), "X")
			displays = append(displays, ":"+displayNum)
		}
	}

	return displays
}

// getWaylandDisplays returns list of active Wayland displays
func getWaylandDisplays() []string {
	displays := []string{}

	if display := os.Getenv("WAYLAND_DISPLAY"); display != "" {
		displays = append(displays, display)
	}

	return displays
}

// getResolution attempts to get the current display resolution
func getResolution(displayServer string) string {
	switch displayServer {
	case "X11":
		return getX11Resolution()
	case "Wayland":
		return getWaylandResolution()
	}
	return ""
}

// getX11Resolution gets resolution from X11
func getX11Resolution() string {
	// Try xrandr first
	if resolution := tryXrandr(); resolution != "" {
		return resolution
	}

	// Try xdpyinfo as fallback
	if resolution := tryXdpyinfo(); resolution != "" {
		return resolution
	}

	return ""
}

// tryXrandr tries to get resolution from xrandr
func tryXrandr() string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "xrandr")
	if display := os.Getenv("DISPLAY"); display != "" {
		cmd.Env = append(os.Environ(), "DISPLAY="+display)
	}

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Parse xrandr output for current resolution
	// Look for lines like "   1920x1080     60.00*+"
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "*") {
			fields := strings.Fields(line)
			if len(fields) > 0 && strings.Contains(fields[0], "x") {
				return fields[0]
			}
		}
	}

	return ""
}

// tryXdpyinfo tries to get resolution from xdpyinfo
func tryXdpyinfo() string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "xdpyinfo")
	if display := os.Getenv("DISPLAY"); display != "" {
		cmd.Env = append(os.Environ(), "DISPLAY="+display)
	}

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Look for "dimensions:    1920x1080 pixels"
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "dimensions:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1]
			}
		}
	}

	return ""
}

// getWaylandResolution gets resolution from Wayland
func getWaylandResolution() string {
	// Try wlr-randr for wlroots-based compositors
	if resolution := tryWlrRandr(); resolution != "" {
		return resolution
	}

	// Try parsing from swaymsg for Sway
	if resolution := trySwaymsg(); resolution != "" {
		return resolution
	}

	return ""
}

// tryWlrRandr tries to get resolution from wlr-randr
func tryWlrRandr() string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "wlr-randr")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Parse for current mode
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "current") {
			fields := strings.Fields(line)
			for _, field := range fields {
				if strings.Contains(field, "x") && strings.Contains(field, "@") {
					// Format: 1920x1080@60.000000
					parts := strings.Split(field, "@")
					if len(parts) > 0 {
						return parts[0]
					}
				}
			}
		}
	}

	return ""
}

// trySwaymsg tries to get resolution from swaymsg
func trySwaymsg() string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "swaymsg", "-t", "get_outputs")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Parse JSON output (simplified - look for "current_mode")
	// This is a simple string search, not full JSON parsing
	if strings.Contains(string(output), "current_mode") {
		// Try to extract resolution pattern
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "width") || strings.Contains(line, "height") {
				// This is a simplified parser - for production use proper JSON
				logrus.Debug("Found sway output, but skipping complex JSON parsing")
				break
			}
		}
	}

	return ""
}

// getIdleTime gets the idle time in milliseconds
func getIdleTime(displayServer string) int64 {
	switch displayServer {
	case "X11":
		return getX11IdleTime()
	case "Wayland":
		// Wayland idle time detection is compositor-specific and complex
		return 0
	}
	return 0
}

// getX11IdleTime gets idle time from X11 using xprintidle or xssstate
func getX11IdleTime() int64 {
	// Try xprintidle first
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "xprintidle")
	if display := os.Getenv("DISPLAY"); display != "" {
		cmd.Env = append(os.Environ(), "DISPLAY="+display)
	}

	output, err := cmd.Output()
	if err == nil {
		idleStr := strings.TrimSpace(string(output))
		if idleMs, err := strconv.ParseInt(idleStr, 10, 64); err == nil {
			return idleMs
		}
	}

	// Try xssstate as fallback
	ctx2, cancel2 := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel2()

	cmd2 := exec.CommandContext(ctx2, "xssstate", "-i")
	if display := os.Getenv("DISPLAY"); display != "" {
		cmd2.Env = append(os.Environ(), "DISPLAY="+display)
	}

	output2, err := cmd2.Output()
	if err == nil {
		output2 = bytes.TrimSpace(output2)
		if idleMs, err := strconv.ParseInt(string(output2), 10, 64); err == nil {
			return idleMs
		}
	}

	return 0
}
