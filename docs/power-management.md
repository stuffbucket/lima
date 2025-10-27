# Power Management in Lima

This document describes how Lima manages VM power states across different virtualization drivers.

## Overview

Lima supports graceful shutdown for all virtualization drivers (VZ, QEMU, WSL2). Each driver implements power management using the underlying hypervisor's native capabilities.

## Driver-Specific Implementations

### VZ Driver (macOS)

The VZ driver uses Apple's Virtualization.framework APIs for power management:

- **Graceful Shutdown**: Uses `VirtualMachine.RequestStop()`, which sends an ACPI power button signal to the guest
- **Timeout**: 30 seconds (configurable)
- **Force Shutdown**: If graceful shutdown times out, automatically falls back to `VirtualMachine.Stop()` for immediate termination
- **Requirements**: Guest must support ACPI (all modern Linux distributions do)

**Implementation Flow:**
1. Check if `CanRequestStop()` returns true
2. Call `RequestStop()` to send ACPI signal
3. Poll VM state every 500ms
4. After 30 seconds, force stop if VM hasn't shut down
5. Log all state transitions for debugging

**Closing GUI Window:** On VZ, closing the VM window triggers an immediate stop (Virtualization.framework behavior).

### QEMU Driver

The QEMU driver uses QMP (QEMU Machine Protocol) for power management:

- **Graceful Shutdown**: Uses QMP `system_powerdown` command
- **Timeout**: 30 seconds
- **Force Shutdown**: If timeout occurs, kills the QEMU process
- **Additional Cleanup**: Removes VNC sockets and vhost devices after shutdown

**Implementation Flow:**
1. Connect to QMP Unix socket
2. Send `{"execute": "system_powerdown"}`
3. Wait up to 30 seconds for process to exit
4. If timeout, use `Kill()` to terminate
5. Clean up network and VNC resources

### WSL2 Driver (Windows)

The WSL2 driver uses Windows Subsystem for Linux APIs:

- Uses `wsl.exe --terminate` for shutdown
- Fast termination (WSL handles shutdown internally)

## CLI Usage

```bash
# Graceful shutdown (waits for timeout if needed)
limactl stop <instance-name>

# Stop all instances
limactl stop --all

# Check VM status
limactl list
```

## Guest Requirements

For graceful shutdown to work properly:
- Guest must support ACPI power management (all modern Linux distributions do)
- Guest should have proper init system (systemd, OpenRC, etc.)
- Guest services should handle SIGTERM signals gracefully

## Comparison with Other Tools

### vs QEMU Guest Agent (qga)

- Lima uses its own custom guest agent via vsock (port 2222)
- Lima's guest agent focuses on port forwarding, filesystem events, and system info
- Power management is handled at the hypervisor level (ACPI), not via guest agent

### vs SPICE

- SPICE protocol handles display, input, clipboard, and audio only
- SPICE does NOT include power management capabilities
- Power management is separate, using QMP (QEMU) or native APIs (VZ)

## Troubleshooting

### Graceful Shutdown Times Out

If the VM doesn't shut down within 30 seconds:
1. Check guest logs for shutdown errors
2. Verify guest has proper init system
3. Check if services are hanging during shutdown
4. Force stop will occur automatically after timeout

### Immediate Stop Needed

For development/testing, you can:
- Close the VZ GUI window (VZ only - triggers immediate stop)
- Wait for timeout (automatic force stop after 30s)
- Kill the process manually (not recommended)

## Implementation Notes

- VZ and QEMU implementations are intentionally similar for consistency
- Both use 30-second timeout with force stop fallback
- Both provide detailed logging for debugging
- VZ uses native macOS APIs, QEMU uses QMP protocol
- No guest agent involvement required for power management
