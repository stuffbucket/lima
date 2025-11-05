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
		Short: "Open the graphical display window for the instance.",
		Long: `Open the graphical display window for the instance.

For VZ instances:
- If the GUI window is minimized or hidden, this brings it back to the foreground
- The window is initially created during VM startup (limactl start)
- Closing the VZ GUI window will STOP the VM (by design in Apple's Virtualization.framework)
- If you closed the window, you must restart the VM to see it again

For QEMU/SPICE instances:
- Launches a new viewer window that can be closed and reopened without affecting the VM

Requirements:
- Instance must be running
- Display must be enabled (VZ: video.display="vz", QEMU: video.display with SPICE)`,
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
