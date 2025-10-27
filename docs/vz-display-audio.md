---
title: VZ Display and Audio
weight: 11
---

Lima's VZ driver provides native macOS display and audio support using Apple's Virtualization.framework.

## Overview

The VZ driver offers superior display and audio integration compared to QEMU, with:

- **Native macOS windows** with full Metal acceleration
- **Built-in audio support** with both input and output
- **No external viewers needed** - everything is native
- **Excellent performance** leveraging Apple Silicon

## Requirements

- macOS 13.0 (Ventura) or later
- Apple Silicon (M1/M2/M3) or Intel Mac (with limitations on older Intel Macs)

## Display Configuration

### Basic Display Setup

Enable the native GUI window:

```yaml
vmType: vz
video:
  display: "vz"
```

Or use the default (same as "vz"):

```yaml
vmType: vz
video:
  display: "default"
```

### Custom Resolution

Configure a specific display resolution:

```yaml
vmType: vz
video:
  display: "vz"
  vz:
    width: 1920   # Window width in pixels
    height: 1200  # Window height in pixels
```

The default resolution is 1920x1200 if not specified.

**Auto-Resize (macOS 14+):**
On macOS 14 (Sonoma) and later, the display automatically reconfigures when you resize the window. The guest OS resolution will update to match the window size if it supports dynamic resolution changes (most modern Linux distributions with proper graphics drivers do).

**For older macOS versions (13.x):**
The window size is fixed at startup. To change it, update the configuration and restart the VM.

### Clipboard Sharing

Clipboard sharing between host and guest is enabled by default (requires macOS 13+):

```yaml
vmType: vz
video:
  display: "vz"
  vz:
    disableClipboard: false  # Default: enabled
```

To disable clipboard sharing:

```yaml
vmType: vz
video:
  display: "vz"
  vz:
    disableClipboard: true
```

**Requirements**:
- macOS 13.0 (Ventura) or later
- `spice-vdagent` installed in the guest OS

**Installing SPICE agent in guest**:

```bash
# Ubuntu/Debian
lima sudo apt-get install -y spice-vdagent

# Fedora/RHEL
lima sudo dnf install -y spice-vdagent

# Arch Linux
lima sudo pacman -S spice-vdagent
```

After installation, clipboard copy/paste will work bidirectionally between host and guest.

### Disable Display (Headless)

For servers or when you only need SSH access:

```yaml
vmType: vz
video:
  display: "none"
```

### Launching the GUI

Start the VM and open the display window:

```bash
# Start the VM
limactl start my-instance

# Open the GUI window
limactl show-gui my-instance
```

The GUI window will open with:
- 1920x1200 default resolution
- Native macOS window controls
- Metal-accelerated rendering
- Automatic HiDPI support on Retina displays

## Audio Configuration

### Enable Audio Output

Allow the guest VM to play audio on the host:

```yaml
vmType: vz
audio:
  device: "vz"
```

Or use the default:

```yaml
vmType: vz
audio:
  device: "default"
```

### Disable Audio

```yaml
vmType: vz
audio:
  device: "none"
```

### Full Configuration Example

```yaml
vmType: vz
video:
  display: "vz"
audio:
  device: "vz"
```

## Features

### Display Features

#### Native macOS Window
- Full macOS window with standard controls (minimize, maximize, close)
- Native menu bar integration
- Mission Control and Spaces support
- Command-Tab app switching

#### Metal Acceleration
- Hardware-accelerated graphics via Apple Metal
- Smooth rendering and video playback
- Support for graphics-intensive applications

#### HiDPI/Retina Support
- Automatic scaling on Retina displays
- Crisp text rendering
- Proper DPI awareness

### Audio Features

#### Audio Output (Playback)
- Stream guest audio to host speakers/headphones
- Low latency audio playback
- Multiple audio format support
- Synchronized with video

#### Audio Input (Microphone)
- Host microphone input to guest VM
- Useful for video conferencing, voice recording
- Requires macOS microphone permissions

## Usage

### Starting with GUI

```bash
# Create and start instance with display
limactl start --name=my-vz-vm template://ubuntu

# Launch the GUI window
limactl show-gui my-vz-vm
```

### Window Management

The VZ display window is a native macOS application window:

- **Minimize**: Click the yellow button or use Command-M
- **Zoom (green button)**: The zoom/maximize button is disabled in VZ windows due to Virtualization.framework limitations. To resize the window, drag the corners or edges manually.
- **Manual resize**: Drag window corners/edges to resize. On macOS 14+, the guest display will automatically adjust to match the new window size.
- **Close**: ⚠️ **Closing the window will STOP the VM** - this is by design in the VZ framework
- **Keep running**: To keep the VM running, minimize the window instead of closing it
- **Window state**: macOS automatically remembers window position and size between VM sessions

**Important**: Unlike QEMU+SPICE where closing the viewer leaves the VM running, **closing the VZ GUI window will stop the VM**. This is how Apple's Virtualization.framework is designed. If you want to access the VM without the GUI:

1. **Before starting**: Configure `video.display: "none"` for headless operation
2. **While running**: Minimize the window (don't close it) and use SSH
3. **To detach GUI**: VZ doesn't support this - use QEMU+SPICE instead if you need detachable GUI

### Keyboard and Mouse

- **Keyboard**: Full keyboard support with macOS key mappings
- **Mouse**: Native cursor integration with automatic grab/release
- **Clipboard**: Bidirectional clipboard sharing via SPICE agent (requires `spice-vdagent` in guest, macOS 13+)
- **Shortcuts**: Most macOS shortcuts work (Command-Tab, Command-Space, etc.)

## Comparison with QEMU

| Feature | VZ Driver | QEMU (with SPICE) |
|---------|-----------|-------------------|
| Native Window | ✅ Built-in | ❌ External viewer |
| Graphics Acceleration | ✅ Metal | ⚠️ OpenGL/Virgl |
| Audio Output | ✅ Built-in | ✅ Requires config |
| Audio Input | ✅ Built-in | ⚠️ Limited |
| Clipboard Sharing | ✅ Via SPICE agent | ✅ Via SPICE agent |
| Setup Complexity | Low | Medium |
| Performance | Excellent | Good |
| macOS Integration | Excellent | Fair |
| USB Redirection | ⚠️ Limited | ✅ Full |
| Remote Access | ❌ No | ✅ Network protocol |
| Cross-platform | macOS only | Multi-platform |

## Advantages of VZ Display

### Better Performance
- Native Apple Silicon integration
- Direct Metal rendering (no translation layer)
- Lower CPU usage
- Better battery life on MacBooks

### Simpler Setup
- No external viewer installation needed
- No network ports to configure
- No password management
- Works out of the box

### Better Integration
- Native macOS window behavior
- Automatic HiDPI handling
- System-level clipboard (with guest agent)
- Mission Control and Spaces integration

## Limitations

### VZ Display Limitations
- **No remote access**: Display only works locally (use SSH for remote)
- **Single display**: No multi-monitor support (as of macOS 13)
- **No recording**: No built-in screen recording API
- **macOS only**: Only works on macOS hosts

### When to Use QEMU Instead
- Need remote graphical access
- Running on Linux/Windows host
- Need USB device passthrough
- Require specific QEMU features
- Need multi-monitor support

## Troubleshooting

### GUI Window Doesn't Open

**Error**: `RunGUI is not supported for the given driver 'vz' and display 'none'`

**Solution**: Check your configuration has `video.display` set to `"vz"` or `"default"`, not `"none"`.

### No Audio in Guest

**Problem**: Guest plays audio but nothing heard on host

**Solution**:
1. Verify audio is enabled: `audio.device: "vz"`
2. Check macOS audio output device is working
3. Check guest OS audio settings
4. Ensure guest has audio drivers (usually automatic with modern Linux)

### Microphone Not Working

**Problem**: Guest can't access microphone

**Solution**:
1. Grant microphone permission to Lima in macOS System Settings → Privacy & Security → Microphone
2. Restart the VM after granting permissions
3. Check guest OS can see the audio input device

### Poor Graphics Performance

**Solution**:
1. Ensure you're on macOS 13.0 or later
2. Update to latest macOS for best performance
3. Check Activity Monitor for CPU/memory constraints
4. Try reducing guest resolution or workload

### Window is Too Small/Large

**Current Limitation**: The window size is fixed at 1920x1200

**Workaround**:
- Use macOS zoom feature (green button) to resize
- Guest can change its resolution (may appear letterboxed)

## Advanced Topics

### Programmatic GUI Control

For automation or custom workflows:

```bash
# Check if GUI can be launched
limactl info my-instance | grep -q "CanRunGUI: true"

# Launch GUI
limactl show-gui my-instance
```

### Guest Display Resolution

The guest VM can change its resolution dynamically. Most modern Linux distributions with proper graphics drivers will support:

- Automatic resolution detection
- Dynamic resolution changes
- HiDPI awareness (on supported guests)

### Audio Permissions

On first use, macOS may prompt for microphone permissions. To manage:

1. Open System Settings
2. Go to Privacy & Security → Microphone
3. Find "lima" or "limactl" in the list
4. Toggle permission on/off

## Future Enhancements

Potential improvements being considered:

- [ ] Configurable window size
- [ ] Multi-monitor support (when Apple adds API)
- [ ] Audio device selection
- [ ] Screen recording support
- [ ] Remote display protocol (VNC/SPICE server)
- [ ] GPU passthrough (when Apple adds API)

## Related Documentation

- [Video Configuration](../video/)
- [Audio Configuration](../audio/)
- [VZ Driver](../vmtype/vz/)
- [QEMU SPICE Display](../spice-display/) - Alternative for remote access
- [Apple Virtualization.framework](https://developer.apple.com/documentation/virtualization)
