package helios

/*
#include "wrapper.h"
#include <stdlib.h>
*/
import "C"

import (
	"unsafe"
)

// HeliosDac is a wrapper around the C++ HeliosDac class.
type DAC struct {
	handle C.HeliosDacHandle
}

// Point corresponds to the standard point structure (8-bit colors, 12-bit XY).
// X, Y: 12-bit coordinates (Range: 0 - 4095). 0 is 0V/Bottom/Left, 4095 is MaxV/Top/Right.
// R, G, B, I: 8-bit color components (Range: 0 - 255).
// Intensity (I) is optional/redundant if RGB are used, but should be set to 255 for full brightness.
type Point struct {
	X, Y       uint16
	R, G, B, I uint8
}

// PointHighRes corresponds to the high-resolution point structure (16-bit colors, 12-bit XY).
// X, Y: 12-bit coordinates (Range: 0 - 4095).
// R, G, B: 16-bit color components (Range: 0 - 65535).
type PointHighRes struct {
	X, Y    uint16
	R, G, B uint16
}

// PointExt corresponds to the extended point structure (all fields 16-bit).
// X, Y: 12-bit coordinates (Range: 0 - 4095).
// R, G, B, I: 16-bit color/intensity components (Range: 0 - 65535).
// User1-4: 16-bit user defined values for accessory ports (Range: 0 - 65535).
type PointExt struct {
	X, Y                       uint16
	R, G, B, I                 uint16
	User1, User2, User3, User4 uint16
}

// New creates a new HeliosDac instance.
func NewDAC() *DAC {
	return &DAC{
		handle: C.HeliosDac_New(),
	}
}

// Close releases the underlying C++ instance.
func (d *DAC) Close() {
	if d.handle != nil {
		C.HeliosDac_Delete(d.handle)
		d.handle = nil
	}
}

// OpenDevices scans for and opens connected devices.
// Returns the number of devices found.
func (d *DAC) OpenDevices() int {
	return int(C.HeliosDac_OpenDevices(d.handle))
}

// OpenDevicesOnlyUsb scans for and opens only USB devices.
func (d *DAC) OpenDevicesOnlyUsb() int {
	return int(C.HeliosDac_OpenDevicesOnlyUsb(d.handle))
}

// OpenDevicesOnlyNetwork scans for and opens only network devices.
func (d *DAC) OpenDevicesOnlyNetwork() int {
	return int(C.HeliosDac_OpenDevicesOnlyNetwork(d.handle))
}

// ReScanDevices scans for new devices (preserves existing connections).
func (d *DAC) ReScanDevices() int {
	return int(C.HeliosDac_ReScanDevices(d.handle))
}

// ReScanDevicesOnlyUsb scans for new USB devices.
func (d *DAC) ReScanDevicesOnlyUsb() int {
	return int(C.HeliosDac_ReScanDevicesOnlyUsb(d.handle))
}

// ReScanDevicesOnlyNetwork scans for new network devices.
func (d *DAC) ReScanDevicesOnlyNetwork() int {
	return int(C.HeliosDac_ReScanDevicesOnlyNetwork(d.handle))
}

// CloseDevices closes all opened devices.
func (d *DAC) CloseDevices() {
	C.HeliosDac_CloseDevices(d.handle)
}

// GetStatus returns the status of the device.
// 1 means ready for next frame.
func (d *DAC) GetStatus(deviceIndex int) int {
	return int(C.HeliosDac_GetStatus(d.handle, C.int(deviceIndex)))
}

// WriteFrame sends a standard frame (8-bit colors, 12-bit XY) to the device.
func (d *DAC) WriteFrame(deviceIndex int, pps int, flags int, points []Point) int {
	if len(points) == 0 {
		return 0
	}
	return int(C.HeliosDac_WriteFrame(
		d.handle,
		C.int(deviceIndex),
		C.int(pps),
		C.int(flags),
		(*C.WrapperHeliosPoint)(unsafe.Pointer(&points[0])),
		C.int(len(points)),
	))
}

// WriteFrameHighResolution sends a high-resolution frame to the device.
// Uses 16-bit XY and RGB. Intensity is ignored.
func (d *DAC) WriteFrameHighResolution(deviceIndex int, pps int, flags int, points []PointHighRes) int {
	if len(points) == 0 {
		return 0
	}
	return int(C.HeliosDac_WriteFrameHighResolution(
		d.handle,
		C.int(deviceIndex),
		C.int(pps),
		C.int(flags),
		(*C.WrapperHeliosPointHighRes)(unsafe.Pointer(&points[0])),
		C.int(len(points)),
	))
}

// WriteFrameExtended sends an extended frame to the device.
// Uses all fields including Intensity and User fields.
func (d *DAC) WriteFrameExtended(deviceIndex int, pps int, flags int, points []PointExt) int {
	if len(points) == 0 {
		return 0
	}
	return int(C.HeliosDac_WriteFrameExtended(
		d.handle,
		C.int(deviceIndex),
		C.int(pps),
		C.int(flags),
		(*C.WrapperHeliosPointExt)(unsafe.Pointer(&points[0])),
		C.int(len(points)),
	))
}

// GetName retrieves the name of the device.
func (d *DAC) GetName(deviceIndex int) string {
	buf := make([]byte, 32)
	C.HeliosDac_GetName(d.handle, C.int(deviceIndex), (*C.char)(unsafe.Pointer(&buf[0])), C.int(len(buf)))
	return C.GoString((*C.char)(unsafe.Pointer(&buf[0])))
}

// GetFirmwareVersion retrieves the firmware version.
func (d *DAC) GetFirmwareVersion(deviceIndex int) int {
	return int(C.HeliosDac_GetFirmwareVersion(d.handle, C.int(deviceIndex)))
}

// GetSupportsHigherResolutions checks if the device supports high resolution data.
func (d *DAC) GetSupportsHigherResolutions(deviceIndex int) int {
	return int(C.HeliosDac_GetSupportsHigherResolutions(d.handle, C.int(deviceIndex)))
}

// GetIsUsb checks if the device is connected via USB.
func (d *DAC) GetIsUsb(deviceIndex int) bool {
	return bool(C.HeliosDac_GetIsUsb(d.handle, C.int(deviceIndex)))
}

// GetIsClosed checks if the device is closed.
func (d *DAC) GetIsClosed(deviceIndex int) bool {
	return bool(C.HeliosDac_GetIsClosed(d.handle, C.int(deviceIndex)))
}

// SetName sets the name of the device.
func (d *DAC) SetName(deviceIndex int, name string) int {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return int(C.HeliosDac_SetName(d.handle, C.int(deviceIndex), cName))
}

// Stop stops output of DAC until new frame is written.
// Blocks for 100ms.
func (d *DAC) Stop(deviceIndex int) int {
	return int(C.HeliosDac_Stop(d.handle, C.int(deviceIndex)))
}

// SetShutter sets the shutter level of the DAC.
// true = open, false = closed.
func (d *DAC) SetShutter(deviceIndex int, level bool) int {
	return int(C.HeliosDac_SetShutter(d.handle, C.int(deviceIndex), C.bool(level)))
}

// EraseFirmware erases the firmware of the DAC.
// Advanced use only.
func (d *DAC) EraseFirmware(deviceIndex int) int {
	return int(C.HeliosDac_EraseFirmware(d.handle, C.int(deviceIndex)))
}

// SetLibusbDebugLogLevel sets the debug log level for libusb.
func (d *DAC) SetLibusbDebugLogLevel(logLevel int) int {
	return int(C.HeliosDac_SetLibusbDebugLogLevel(d.handle, C.int(logLevel)))
}
