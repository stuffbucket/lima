# SPICE Client Package

This package provides SPICE (Simple Protocol for Independent Computing Environments) client support for Lima VMs.

## Overview

The `spiceclient` package enables Lima to connect to SPICE-enabled QEMU VMs, providing high-performance graphical display capabilities. This is particularly useful when using the `qemu-spice` build of QEMU, which includes enhanced SPICE support.

## Features

- **Auto-detection of SPICE viewers**: Automatically finds and launches available SPICE client applications
- **Multiple viewer support**: Works with `remote-viewer`, `spicy`, and `virt-viewer`
- **Connection parsing**: Parses SPICE connection strings from QEMU display configuration
- **TCP and Unix socket support**: Supports both network and local socket connections
- **Audio streaming**: Supports SPICE audio when properly configured

## Supported SPICE Viewers

### macOS
- `remote-viewer` (from virt-viewer package, can be installed via Homebrew)
- `spicy` (from spice-gtk package)

### Linux
- `remote-viewer` (most common, available in most distributions)
- `spicy` (from spice-gtk)
- `virt-viewer`

### Windows
- `remote-viewer.exe`
- `spicy.exe`
- `virt-viewer.exe`

## Installation of SPICE Viewers

### macOS
```bash
# Install virt-viewer (includes remote-viewer)
brew install virt-viewer

# Or install spice-gtk
brew install spice-gtk
```

### Linux (Ubuntu/Debian)
```bash
sudo apt-get install virt-viewer
# or
sudo apt-get install spice-client-gtk
```

### Linux (Fedora/RHEL)
```bash
sudo dnf install virt-viewer
# or
sudo dnf install spice-gtk
```

## Usage with Lima

### Configure SPICE Display in Lima YAML

```yaml
vmType: qemu
video:
  display: "spice,port=5930,addr=127.0.0.1,disable-ticketing=on"
```

### Using QEMU with SPICE

Set the QEMU executable to use the qemu-spice build:

```bash
export QEMU_SYSTEM_X86_64="/opt/homebrew/opt/qemu-spice/bin/qemu-system-x86_64"
```

### Launch SPICE Viewer

The SPICE viewer can be launched automatically when starting the VM or manually via the Lima driver API.

## SPICE Display Options

### Basic SPICE
```yaml
video:
  display: "spice"
```

### SPICE with Custom Port
```yaml
video:
  display: "spice,port=5930"
```

### SPICE with Password
```yaml
video:
  display: "spice,port=5930,password=mysecret"
```

### SPICE with Unix Socket
```yaml
video:
  display: "spice+unix:///var/run/lima/instance/spice.sock"
```

### SPICE with GL (OpenGL acceleration)
```yaml
video:
  display: "spice-app,gl=on"
```

### SPICE with Audio
```yaml
video:
  display: "spice,port=5930"
  spice:
    audio: true
audio:
  device: "default"
```

## API Usage

### Launch a SPICE Viewer

```go
import "github.com/lima-vm/lima/v2/pkg/spiceclient"

conn := &spiceclient.Connection{
    Host: "127.0.0.1",
    Port: "5930",
    Audio: true,  // Enable audio streaming
}

err := spiceclient.LaunchViewer(ctx, conn)
```

### Parse SPICE Connection String
```go
conn, err := spiceclient.GetConnectionInfo("spice,port=5930,addr=127.0.0.1")
if err != nil {
    // handle error
}
// conn.Host == "127.0.0.1"
// conn.Port == "5930"
```

## Integration with QEMU Driver

The SPICE client is automatically integrated with Lima's QEMU driver:

- When `video.display` starts with `"spice"`, the driver will use SPICE instead of VNC
- Password changes via `ChangeDisplayPassword()` work with SPICE
- `DisplayConnection()` returns SPICE connection information
- `RunGUI()` automatically launches a SPICE viewer

## Comparison with VNC

### SPICE Advantages

- Better graphics performance
- Native resolution and multi-monitor support
- Audio streaming support
- USB redirection
- Clipboard sharing
- Better compression and lower bandwidth usage

### VNC Advantages
- Simpler protocol
- More widely supported
- Lower resource usage on the host

## Troubleshooting

### No SPICE viewer found
Install one of the supported SPICE viewers (see Installation section above).

### Connection refused
Ensure QEMU is started with SPICE display enabled and the port is not blocked by firewall.

### Password authentication failed
Check that the password is correctly set in the display configuration and matches what was set via QMP.

## References

- [SPICE Protocol](https://www.spice-space.org/)
- [QEMU SPICE Support](https://www.qemu.org/docs/master/system/invocation.html#options-display)
- [UTM CocoaSpice](https://github.com/utmapp/CocoaSpice) - Inspiration for this implementation
- [Virt-viewer](https://gitlab.com/virt-viewer/virt-viewer)
