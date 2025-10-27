# VZ Driver Enhancement Proposal: Advanced Display and Audio Configuration

## Current State

The VZ driver already has excellent display and audio support through Apple's Virtualization.framework:

- Native macOS GUI window with Metal acceleration
- Built-in audio output and input
- Simple configuration: `video.display: "vz"` and `audio.device: "vz"`

## Proposed Enhancements

### 1. Add VZOptions Type Definition

Similar to SPICEOptions for QEMU, add structured configuration for VZ-specific features:

```go
// In pkg/limatype/lima_yaml.go

type VZDisplayOptions struct {
    // Window width in pixels (default: 1920)
    Width *int `yaml:"width,omitempty" json:"width,omitempty" jsonschema:"nullable"`
    // Window height in pixels (default: 1200)
    Height *int `yaml:"height,omitempty" json:"height,omitempty" jsonschema:"nullable"`
    // Pixels per inch for HiDPI support (default: 72, Retina: 144)
    PixelsPerInch *int `yaml:"pixelsPerInch,omitempty" json:"pixelsPerInch,omitempty" jsonschema:"nullable"`
}

type VZAudioOptions struct {
    // Enable audio input (microphone) from host to guest
    InputEnabled *bool `yaml:"inputEnabled,omitempty" json:"inputEnabled,omitempty" jsonschema:"nullable"`
    // Enable audio output (speakers) from guest to host
    OutputEnabled *bool `yaml:"outputEnabled,omitempty" json:"outputEnabled,omitempty" jsonschema:"nullable"`
}

type Video struct {
    Display *string          `yaml:"display,omitempty" json:"display,omitempty" jsonschema:"nullable"`
    VNC     VNCOptions       `yaml:"vnc,omitempty" json:"vnc,omitempty"`
    SPICE   SPICEOptions     `yaml:"spice,omitempty" json:"spice,omitempty"`
    VZ      VZDisplayOptions `yaml:"vz,omitempty" json:"vz,omitempty"` // NEW
}

type Audio struct {
    Device *string        `yaml:"device,omitempty" json:"device,omitempty" jsonschema:"nullable"`
    VZ     VZAudioOptions `yaml:"vz,omitempty" json:"vz,omitempty"` // NEW
}
```

### 2. Update VZ VM Creation

Modify `pkg/driver/vz/vm_darwin.go`:

```go
func attachDisplay(inst *limatype.Instance, vmConfig *vz.VirtualMachineConfiguration) error {
    switch *inst.Config.Video.Display {
    case "vz", "default":
        graphicsDeviceConfiguration, err := vz.NewVirtioGraphicsDeviceConfiguration()
        if err != nil {
            return err
        }
        
        // Get configured dimensions or use defaults
        width := 1920
        height := 1200
        pixelsPerInch := 72
        
        if inst.Config.Video.VZ.Width != nil {
            width = *inst.Config.Video.VZ.Width
        }
        if inst.Config.Video.VZ.Height != nil {
            height = *inst.Config.Video.VZ.Height
        }
        if inst.Config.Video.VZ.PixelsPerInch != nil {
            pixelsPerInch = *inst.Config.Video.VZ.PixelsPerInch
        }
        
        scanoutConfiguration, err := vz.NewVirtioGraphicsScanoutConfiguration(
            int64(width), 
            int64(height),
        )
        if err != nil {
            return err
        }
        
        // Set pixels per inch for Retina displays
        scanoutConfiguration.SetPixelsPerInch(int64(pixelsPerInch))
        
        graphicsDeviceConfiguration.SetScanouts(scanoutConfiguration)
        vmConfig.SetGraphicsDevicesVirtualMachineConfiguration([]vz.GraphicsDeviceConfiguration{
            graphicsDeviceConfiguration,
        })
        return nil
    case "none":
        return nil
    default:
        return fmt.Errorf("unexpected video display %q", *inst.Config.Video.Display)
    }
}

func attachAudio(inst *limatype.Instance, config *vz.VirtualMachineConfiguration) error {
    switch *inst.Config.Audio.Device {
    case "vz", "default":
        // Check what's enabled (default: output only)
        inputEnabled := false
        outputEnabled := true
        
        if inst.Config.Audio.VZ.InputEnabled != nil {
            inputEnabled = *inst.Config.Audio.VZ.InputEnabled
        }
        if inst.Config.Audio.VZ.OutputEnabled != nil {
            outputEnabled = *inst.Config.Audio.VZ.OutputEnabled
        }
        
        var devices []vz.AudioDeviceConfiguration
        
        if outputEnabled {
            outputStream, err := vz.NewVirtioSoundDeviceHostOutputStreamConfiguration()
            if err != nil {
                return err
            }
            outputDevice, err := vz.NewVirtioSoundDeviceConfiguration()
            if err != nil {
                return err
            }
            outputDevice.SetStreams(outputStream)
            devices = append(devices, outputDevice)
        }
        
        if inputEnabled {
            inputStream, err := vz.NewVirtioSoundDeviceHostInputStreamConfiguration()
            if err != nil {
                return err
            }
            inputDevice, err := vz.NewVirtioSoundDeviceConfiguration()
            if err != nil {
                return err
            }
            inputDevice.SetStreams(inputStream)
            devices = append(devices, inputDevice)
        }
        
        if len(devices) > 0 {
            config.SetAudioDevicesVirtualMachineConfiguration(devices)
        }
        
        return nil
    case "", "none":
        return nil
    default:
        return fmt.Errorf("unexpected audio device %q", *inst.Config.Audio.Device)
    }
}
```

### 3. Update RunGUI to Use Configured Size

Modify `pkg/driver/vz/vz_driver_darwin.go`:

```go
func (l *LimaVzDriver) RunGUI() error {
    if l.canRunGUI() {
        width := 1920
        height := 1200
        
        if l.Instance.Config.Video.VZ.Width != nil {
            width = *l.Instance.Config.Video.VZ.Width
        }
        if l.Instance.Config.Video.VZ.Height != nil {
            height = *l.Instance.Config.Video.VZ.Height
        }
        
        return l.machine.StartGraphicApplication(width, height)
    }
    return fmt.Errorf("RunGUI is not supported for the given driver '%s' and display '%s'", 
        "vz", *l.Instance.Config.Video.Display)
}
```

## Example Configurations

### Basic VZ with Display and Audio

```yaml
vmType: vz
video:
  display: "vz"
audio:
  device: "vz"
```

### Custom Resolution for Retina Display

```yaml
vmType: vz
video:
  display: "vz"
  vz:
    width: 2560
    height: 1600
    pixelsPerInch: 144  # Retina
```

### Audio Output Only (No Microphone)

```yaml
vmType: vz
audio:
  device: "vz"
  vz:
    inputEnabled: false
    outputEnabled: true
```

### Full-Featured Configuration

```yaml
vmType: vz
video:
  display: "vz"
  vz:
    width: 3840
    height: 2160
    pixelsPerInch: 144
audio:
  device: "vz"
  vz:
    inputEnabled: true
    outputEnabled: true
cpus: 4
memory: "8GiB"
```

## Benefits

### 1. Better Display Control
- Users can specify resolution that matches their workflow
- Proper Retina support with pixelsPerInch
- Easier to optimize for different screen sizes

### 2. Better Audio Control
- Disable microphone input for privacy
- Disable audio entirely when not needed
- Clearer configuration vs. "all or nothing"

### 3. Consistency with QEMU
- Similar options structure as SPICEOptions
- Easier to understand for users who use both drivers
- Clear driver-specific configuration namespacing

### 4. Future Extensibility
- Easy to add more VZ-specific options later
- Examples: GPU settings, display count, color depth, etc.

## Implementation Steps

1. **Add type definitions** to `pkg/limatype/lima_yaml.go`
2. **Update display attachment** in `pkg/driver/vz/vm_darwin.go`
3. **Update audio attachment** in `pkg/driver/vz/vm_darwin.go`
4. **Update RunGUI** in `pkg/driver/vz/vz_driver_darwin.go`
5. **Update validation** to check for valid ranges
6. **Add documentation** with examples
7. **Add tests** for new configuration options

## Backward Compatibility

All new options are optional with sensible defaults:
- Default width: 1920
- Default height: 1200
- Default pixelsPerInch: 72
- Default inputEnabled: false
- Default outputEnabled: true

Existing configurations continue to work without changes.

## Comparison: VZ vs QEMU/SPICE

| Feature | VZ (Enhanced) | QEMU/SPICE |
|---------|---------------|------------|
| Display | Native macOS window | External viewer |
| Resolution Control | ✅ New feature | Via QEMU args |
| HiDPI Support | ✅ Native | Manual |
| Audio Output | ✅ Built-in | Requires config |
| Audio Input | ✅ Configurable | Limited |
| Setup | Simple | Complex |
| Performance | Excellent | Good |
| Remote Access | No | Yes |

## Notes

- VZ driver requires macOS 13.0+ (already enforced)
- Audio input requires microphone permission in macOS
- Display resolution is limited by guest OS capabilities
- StartGraphicApplication is a Code-Hex/vz library method
