# SPICE Clipboard Sharing Test Results

## Test Environment
- Host OS: macOS (Ventura 13+)
- Lima Version: d936fb01.m (custom build with SPICE agent support)
- VM: ubuntu-desktop (VZ driver)
- Guest OS: Ubuntu 24.04 with Xubuntu desktop

## Implementation Summary

### Files Created/Modified

1. **pkg/driver/vz/spice_darwin.go** - New file
   - Implements `attachSpiceAgent()` function
   - Creates SpiceAgentPortAttachment for clipboard sharing
   - Configures VirtioConsoleDeviceConfiguration with SPICE agent port
   - Handles graceful degradation if SPICE agent unavailable

2. **pkg/limatype/lima_yaml.go** - Modified
   - Added `DisableClipboard *bool` field to `VZOptions` struct
   - Documented clipboard requirements (macOS 13+, spice-vdagent)

3. **pkg/driver/vz/vm_darwin.go** - Modified
   - Added `attachSpiceAgent()` call after display attachment
   - SPICE agent configured before folder mounts

4. **docs/vz-display-audio.md** - Updated
   - Added "Clipboard Sharing" section with configuration examples
   - Added guest installation instructions for spice-vdagent
   - Updated comparison table to show clipboard support
   - Updated "Keyboard and Mouse" section

## Configuration

The SPICE agent is **enabled by default** for VZ VMs with display:

```yaml
vmType: vz
video:
  display: "vz"
  vz:
    width: 1440
    height: 900
    # disableClipboard: false  # Default: enabled
```

To disable clipboard sharing:

```yaml
vmType: vz
video:
  display: "vz"
  vz:
    disableClipboard: true
```

## Build Verification

```bash
$ make limactl
✅ Build successful
✅ Code signed successfully
```

## Implementation Details

### SPICE Agent Integration

The implementation uses Apple Virtualization.framework's `SpiceAgentPortAttachment`:

1. **Port Creation**: Creates SPICE agent port with name "com.redhat.spice.0"
2. **Clipboard Enable**: Calls `SetSharesClipboard(true)` on the attachment
3. **Console Device**: Attaches to VirtioConsoleDeviceConfiguration
4. **VM Configuration**: Sets console device via `SetConsoleDevicesVirtualMachineConfiguration()`

### Error Handling

The implementation gracefully handles errors:
- If SPICE agent creation fails, warning is logged but VM continues
- If port attachment fails, clipboard just won't work (non-fatal)
- Requires macOS 13+, but code checks for API availability

### Guest Requirements

For clipboard to work, the guest needs:

```bash
# Ubuntu/Debian
sudo apt-get install -y spice-vdagent

# Fedora/RHEL
sudo dnf install -y spice-vdagent

# Arch Linux
sudo pacman -S spice-vdagent
```

The agent must be running in the guest OS to enable clipboard functionality.

## Testing Steps

### 1. Install SPICE Agent in Guest

```bash
# SSH into the running VM
limactl shell ubuntu-desktop

# Install spice-vdagent
sudo apt-get update
sudo apt-get install -y spice-vdagent

# Start the agent (should auto-start after install)
sudo systemctl enable spice-vdagent
sudo systemctl start spice-vdagent

# Verify it's running
systemctl status spice-vdagent
```

### 2. Test Clipboard Functionality

**Host to Guest**:
1. Copy text on macOS (Command-C)
2. Switch to VM window
3. Paste in guest application (Ctrl-V)

**Guest to Host**:
1. Copy text in guest application (Ctrl-C)
2. Switch to macOS application
3. Paste on macOS (Command-V)

### 3. Verify SPICE Agent Port

```bash
# In the guest VM, check for SPICE port
ls -la /dev/virtio-ports/

# Should see: com.redhat.spice.0
```

## Expected Results

✅ **Build Success**: limactl builds without errors
✅ **VM Starts**: Ubuntu desktop VM starts with VZ driver
✅ **SPICE Agent**: Console device created with SPICE agent port
✅ **Clipboard**: Bidirectional clipboard sharing works when spice-vdagent installed

## Current Status

- [x] Implementation complete
- [x] Code compiles successfully
- [x] Documentation updated
- [ ] Guest agent installation (pending manual test)
- [ ] Clipboard functionality test (pending guest agent)
- [ ] Verify console device in VM (pending inspection)

## Next Steps

1. **Install spice-vdagent** in the ubuntu-desktop VM
2. **Test clipboard** copy/paste both directions
3. **Verify console port** exists in guest at `/dev/virtio-ports/com.redhat.spice.0`
4. **Document results** of manual testing
5. **Consider unit tests** for SPICE agent attachment (if feasible)

## Notes

- SPICE agent is separate from SPICE protocol display (used in QEMU)
- VZ uses native Apple display, but can use SPICE agent for clipboard
- This provides best of both worlds: Metal graphics + clipboard sharing
- Implementation follows Apple's Virtualization.framework best practices
- Graceful degradation ensures VM works even if clipboard fails

## References

- [Apple Virtualization.framework - SpiceAgentPortAttachment](https://developer.apple.com/documentation/virtualization/vzspiceagentportattachment)
- [Code-Hex/vz v3.7.1 - SpiceAgentPortAttachment](https://pkg.go.dev/github.com/Code-Hex/vz/v3@v3.7.1#SpiceAgentPortAttachment)
- [SPICE vdagent](https://www.spice-space.org/spice-for-newbies.html)
