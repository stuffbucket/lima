// SPDX-FileCopyrightText: Copyright The Lima Authors
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"fmt"

	"github.com/lima-vm/lima/v2/pkg/driverutil"
	"github.com/lima-vm/lima/v2/pkg/limatype"
)

// populateGUIInfo populates GUI-related information in the instance
func populateGUIInfo(inst *limatype.Instance) {
	if inst.Config == nil || inst.Config.Video.Display == nil {
		return
	}

	gui := &limatype.GUIInfo{
		Display: *inst.Config.Video.Display,
		Enabled: *inst.Config.Video.Display != "none",
	}

	// Determine if GUI can be run based on driver capabilities
	if inst.Status == limatype.StatusRunning || inst.Status == limatype.StatusStopped {
		if configuredDriver, err := driverutil.CreateConfiguredDriver(inst, inst.SSHLocalPort); err == nil {
			info := configuredDriver.Info()
			gui.CanRunGUI = info.Features.CanRunGUI
		}
	}

	// Get resolution if configured
	if inst.Config.Video.VZ.Width != nil && inst.Config.Video.VZ.Height != nil {
		gui.Resolution = fmt.Sprintf("%dx%d", *inst.Config.Video.VZ.Width, *inst.Config.Video.VZ.Height)
	} else if gui.Display == "vz" || gui.Display == "default" {
		gui.Resolution = "1920x1200" // default
	}

	// Check clipboard sharing
	if inst.Config.Video.VZ.DisableClipboard != nil {
		gui.ClipboardShared = !*inst.Config.Video.VZ.DisableClipboard
	} else {
		// Default is enabled for VZ with display
		gui.ClipboardShared = gui.Enabled && (gui.Display == "vz" || gui.Display == "default")
	}

	// Check audio
	if inst.Config.Audio.Device != nil {
		gui.AudioEnabled = *inst.Config.Audio.Device != "none"
	}

	inst.GUI = gui
}
