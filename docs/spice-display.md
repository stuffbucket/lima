---
title: SPICE Display
weight: 10
---

Lima supports SPICE (Simple Protocol for Independent Computing Environments) for high-performance graphical display of VMs.

## Overview

SPICE provides better graphics performance and features compared to VNC, including:

- Native resolution and multi-monitor support
- Audio streaming
- USB device redirection
- Clipboard sharing
- Better compression and lower bandwidth usage

## Requirements

### QEMU with SPICE Support

You need QEMU compiled with SPICE support. On macOS, you can use the `qemu-spice` build:

```bash
# Install qemu-spice from stuffbucket/qemu-spice tap
brew install stuffbucket/qemu-spice/qemu-spice

# Set Lima to use the qemu-spice binary
export QEMU_SYSTEM_X86_64="/opt/homebrew/opt/qemu-spice/bin/qemu-system-x86_64"
```

### SPICE Viewer

Install a SPICE client application:

#### macOS

```bash
# Install virt-viewer (includes remote-viewer)
brew install virt-viewer
```

#### Linux

```bash
# Ubuntu/Debian
sudo apt-get install virt-viewer

# Fedora/RHEL
sudo dnf install virt-viewer
```

## Configuration

### Basic SPICE Setup

Configure your Lima instance to use SPICE display:

```yaml
vmType: qemu
video:
  display: "spice,port=5930,addr=127.0.0.1,disable-ticketing=on"
```

### SPICE with Audio

Enable audio streaming over SPICE:

```yaml
vmType: qemu
video:
  display: "spice,port=5930,addr=127.0.0.1,disable-ticketing=on"
  spice:
    audio: true
audio:
  device: "default"
```

**Note**: Audio requires a SPICE viewer that supports audio (e.g., `remote-viewer`, `virt-viewer`).

### SPICE with Password Protection

```yaml
vmType: qemu
video:
  display: "spice,port=5930,addr=127.0.0.1"
```

Then set the password after starting the VM:

```bash
limactl shell myinstance -- set-display-password <password>
```

### SPICE with OpenGL Acceleration

For better graphics performance with 3D acceleration:

```yaml
vmType: qemu
video:
  display: "spice-app,gl=on"
  spice:
    gl: true
```

### Full-Featured SPICE Configuration

Combine display, audio, OpenGL, and other SPICE features:

```yaml
vmType: qemu
video:
  display: "spice,port=5930,addr=127.0.0.1,disable-ticketing=on,gl=on"
  spice:
    gl: true
    audio: true
    agent: true
    streamingVideo: "filter"
audio:
  device: "default"
```

## Usage

### Starting a VM with SPICE

```bash
# Create a new instance with SPICE
limactl start --name=my-spice-vm my-spice-config.yaml
```

### Connecting to SPICE Display

Lima can automatically launch a SPICE viewer:

```bash
# Launch the configured SPICE viewer
limactl show-gui my-spice-vm
```

Or connect manually using `remote-viewer`:

```bash
# Get the SPICE connection details
SPICE_PORT=$(limactl show-connection my-spice-vm)

# Connect with remote-viewer
remote-viewer spice://127.0.0.1:${SPICE_PORT}
```

## SPICE Display Options

### Common Options

- `port=<port>` - SPICE server port (default: 5900)
- `addr=<address>` - Bind address (default: 127.0.0.1)
- `disable-ticketing=on` - Disable password authentication
- `password=<password>` - Set initial password
- `gl=on` - Enable OpenGL acceleration (requires spice-app)

### Examples

#### Basic SPICE on default port

```yaml
video:
  display: "spice"
```

#### SPICE on custom port

```yaml
video:
  display: "spice,port=5931"
```

#### SPICE with TLS (secure connection)

```yaml
video:
  display: "spice,tls-port=5901,x509-dir=/path/to/certs"
```

## Comparison with VNC

| Feature | SPICE | VNC |
|---------|-------|-----|
| Graphics Performance | Excellent | Good |
| Audio Support | Yes (with config) | No |
| USB Redirection | Yes | No |
| Clipboard Sharing | Yes | Limited |
| Multi-monitor | Yes | Limited |
| Bandwidth Usage | Low (compressed) | Medium |
| Setup Complexity | Medium | Low |

## Troubleshooting

### SPICE viewer not found

**Error**: `no SPICE viewer found, install remote-viewer or spicy`

**Solution**: Install a SPICE viewer application (see Requirements section).

### Connection refused

**Error**: `failed to connect to SPICE display`

**Solution**:
- Ensure the VM is running: `limactl list`
- Check the SPICE port is not blocked by firewall
- Verify SPICE is enabled in the VM configuration

### QEMU doesn't support SPICE

**Error**: `QEMU does not support SPICE display`

**Solution**: Use a QEMU build with SPICE support (see Requirements section).

### Black screen or no display

**Solution**:
- Try adding `gl=off` to disable OpenGL acceleration
- Check that the guest has video drivers installed
- Verify the display device is configured correctly

### No audio in SPICE session

**Error**: Audio doesn't work in SPICE viewer

**Solution**:
- Enable audio in both the video and audio sections:
  ```yaml
  video:
    spice:
      audio: true
  audio:
    device: "default"
  ```
- Verify your SPICE viewer supports audio (remote-viewer does, spicy does)
- Check guest OS has audio drivers installed
- On macOS host, ensure QEMU has microphone permissions if needed

## Advanced Configuration

### Using Unix Sockets

For local connections, Unix sockets can provide better performance:

```yaml
video:
  display: "spice+unix:///tmp/lima-spice.sock"
```

### Custom SPICE Arguments

For advanced SPICE configurations, you can use QEMU_SYSTEM_* environment variables:

```bash
export QEMU_SYSTEM_X86_64="/path/to/qemu-system-x86_64 -spice <custom-options>"
```

Note: This is for debugging only and not recommended for production use.

## Related Documentation

- [Video Configuration](../video/)
- [QEMU Driver](../vmtype/qemu/)
- [SPICE Protocol Documentation](https://www.spice-space.org/)
