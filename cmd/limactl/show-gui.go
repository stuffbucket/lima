// SPDX-FileCopyrightText: Copyright The Lima Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/lima-vm/lima/v2/pkg/driverutil"
	"github.com/lima-vm/lima/v2/pkg/store"
)

func newShowGUICommand() *cobra.Command {
	showGUICmd := &cobra.Command{
		Use:   "show-gui INSTANCE",
		Short: "Open the graphical display window for the instance",
		Long: `Open the graphical display window for the instance.

This command launches the native display window for instances using VZ driver
with display enabled.

IMPORTANT - VZ Framework Limitation:
The GUI window is opened during VM startup (limactl start). This command 
cannot reopen a closed window - that would require restarting the VM. 
Additionally, closing the VZ GUI window will STOP the VM (by design in 
Apple's Virtualization.framework).

If you need a detachable GUI that doesn't stop the VM when closed, use 
QEMU with SPICE display instead.

Requirements:
- Instance must be running
- Instance must use VZ driver (vmType: vz)
- Display must be enabled (video.display: "vz" or "default")`,
		Args:              WrapArgsError(cobra.ExactArgs(1)),
		RunE:              showGUIAction,
		ValidArgsFunction: showGUIBashComplete,
		SilenceErrors:     true,
		GroupID:           advancedCommand,
	}

	return showGUICmd
}

func showGUIAction(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	instName := args[0]

	// Check if instance exists
	inst, err := store.Inspect(ctx, instName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("instance %q does not exist, run `limactl create %s` to create a new instance", instName, instName)
		}
		return err
	}

	// Check if instance is running
	if inst.Status != "Running" {
		return fmt.Errorf("instance %q is not running (status: %s), run `limactl start %s` to start it", instName, inst.Status, instName)
	}

	// Check GUI info from instance
	if inst.GUI == nil || !inst.GUI.Enabled {
		displayType := "none"
		if inst.GUI != nil {
			displayType = inst.GUI.Display
		}
		return fmt.Errorf("GUI is not enabled for instance %q (display: %s)", instName, displayType)
	}

	if !inst.GUI.CanRunGUI {
		return fmt.Errorf("GUI is not supported for instance %q (driver: %s, display: %s)", instName, inst.VMType, inst.GUI.Display)
	}

	// Special handling for VZ: GUI cannot be reopened on running VM
	if inst.VMType == "vz" {
		return fmt.Errorf("VZ GUI window cannot be reopened after VM startup. "+
			"The window is created during 'limactl start' and closing it stops the VM. "+
			"To see the GUI again, stop and restart the instance: 'limactl stop %s && limactl start %s'", instName, instName)
	}

	// Get the configured driver for this instance
	configuredDriver, err := driverutil.CreateConfiguredDriver(inst, inst.SSHLocalPort)
	if err != nil {
		return fmt.Errorf("failed to create driver for instance %q: %w", instName, err)
	}

	// Launch the GUI
	logrus.Infof("Launching GUI window for instance %q...", instName)
	if err := configuredDriver.RunGUI(); err != nil {
		return fmt.Errorf("failed to launch GUI: %w", err)
	}

	return nil
}

func showGUIBashComplete(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	// Only complete running instances with GUI support
	instances, directive := bashCompleteInstanceNames(cmd)

	// Filter to only running instances with GUI support
	var guiInstances []string
	for _, instName := range instances {
		if instInfo, err := store.Inspect(cmd.Context(), instName); err == nil {
			if instInfo.Status == "Running" {
				if configuredDriver, err := driverutil.CreateConfiguredDriver(instInfo, instInfo.SSHLocalPort); err == nil {
					info := configuredDriver.Info()
					if info.Features.CanRunGUI {
						guiInstances = append(guiInstances, instName)
					}
				}
			}
		}
	}

	return guiInstances, directive
}
