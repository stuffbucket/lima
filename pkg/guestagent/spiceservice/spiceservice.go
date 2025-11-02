// SPDX-FileCopyrightText: Copyright The Lima Authors
// SPDX-License-Identifier: Apache-2.0

//go:build !linux

package spiceservice

import (
	"context"
)

// SpiceStatus represents the status of SPICE-related services
type SpiceStatus struct {
	AgentInstalled bool
	AgentRunning   bool
	VPortExists    bool
	ClipboardReady bool
	ErrorMessage   string
}

// DetectSpiceStatus returns a stub status for non-Linux platforms
func DetectSpiceStatus(ctx context.Context) *SpiceStatus {
	return &SpiceStatus{
		ErrorMessage: "SPICE agent only available on Linux guests",
	}
}

// EnsureSpiceAgent is a no-op on non-Linux platforms
func EnsureSpiceAgent(ctx context.Context) error {
	return nil
}
