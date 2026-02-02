# Helios DAC Go Bindings

This package provides Go bindings for the Helios DAC C++ SDK. It uses CGO to wrap the underlying C++ library, allowing Go applications to control Helios laser DAC controllers seamlessly via USB or Ethernet.

## Overview

The bindings are designed to be a thin, efficient wrapper around the official C++ SDK. It exposes all core functionality required to discover devices, manage connection states, and stream point data in standard, high-resolution, or extended formats.

## Usage

### Bazel Integration

To use this library in your Bazel project, add the target as a dependency:

```starlark
go_binary(
    name = "my_laser_app",
    srcs = ["main.go"],
    deps = [
        "//third_party/helios_dac/go:helios",
    ],
)
```

### Basic Example

```go
package main

import (
 "fmt"
 "time"

 "github.com/tidaltech/koi/third_party/helios_dac/go/helios"
)

func main() {
  // Initialize the DAC manager
  dac := helios.NewDAC()
  defer dac.Close()

  // Scan for devices
  numDevices := dac.OpenDevices()
  fmt.Printf("Found %d devices\n", numDevices)

  if numDevices == 0 {
    return
  }

   // Create a simple point (center of projection)
   // X, Y are 12-bit (0-4095). Colors are 8-bit (0-255).
   points := []helios.Point{
     {X: 2048, Y: 2048, R: 255, G: 0, B: 0, I: 255},
   }

   // Send frame to first device (index 0)
   // PPS: 30000, Flags: 0
   dac.WriteFrame(0, 30000, 0, points)

   // Wait to ensure transmission (in a real loop, you'd manage timing)
   time.Sleep(100 * time.Millisecond)

   // Stop output
   dac.Stop(0)

   // Close connections
   dac.CloseDevices()
}
```

For a complete, runnable example, see `examples/simple/main.go`:
```bash
bazel run //third_party/helios_dac/go/examples/simple
```

## Functionality Comparison

The Go bindings provide 1:1 parity with the features of the C++ SDK. The primary difference is how method overloading is handled; since Go does not support overloading, distinct method names are used for different frame types.

| Feature Category | C++ SDK Method | Go SDK Method | Notes |
| :--- | :--- | :--- | :--- |
| **Lifecycle** | `HeliosDac()` / `~HeliosDac()` | `NewDAC()` / `Close()` | `Close()` must be called to free C++ resources. |
| **Discovery** | `OpenDevices()` | `OpenDevices()` | Also supports `OnlyUsb` and `OnlyNetwork` variants. |
| | `CloseDevices()` | `CloseDevices()` | |
| **Data Types** | `HeliosPoint` | `Point` | 12-bit XY (in uint16), 8-bit Color. |
| | `HeliosPointHighRes` | `PointHighRes` | 12-bit XY, 16-bit Color. |
| | `HeliosPointExt` | `PointExt` | 16-bit Color + Intensity + User fields. |
| **Frame Output** | `WriteFrame(..., HeliosPoint*)` | `WriteFrame(...)` | |
| | `WriteFrame(..., HeliosPointHighRes*)` | `WriteFrameHighResolution(...)` | Explicit naming for type safety. |
| | `WriteFrame(..., HeliosPointExt*)` | `WriteFrameExtended(...)` | |
| **Control** | `Stop(i)` | `Stop(i)` | Blocks for ~100ms. |
| | `SetShutter(i, bool)` | `SetShutter(i, bool)` | |
| | `SetName(i, name)` | `SetName(i, string)` | Handles C-string conversion automatically. |
| **Status/Info** | `GetStatus(i)` | `GetStatus(i)` | Returns 1 if ready for next frame. |
| | `GetName(i)` | `GetName(i)` | |
| | `GetFirmwareVersion(i)` | `GetFirmwareVersion(i)` | |
| | `GetIsUsb(i)` | `GetIsUsb(i)` | |

## Performance

The performance overhead of using these Go bindings compared to the native C++ SDK is negligible for standard laser operations.

1. **Low API Overhead**: CGO calls typically incur an overhead of ~50-100ns. Since `WriteFrame` is called once per frame (e.g., 60 times/sec) rather than per-point, this total overhead is < 6µs per second (0.0006% of CPU time).
2. **Zero-Copy Transmission**: The Go structs are memory-aligned with the C++ SDK. Data is passed by pointer, avoiding expensive marshalling or copying.
3. **Hardware Bound**: The bottlenecks are the DAC communication (USB/Network) and physical scanner speeds (typically 30k-60k points per second), which are orders of magnitude slower than the Go/CGO layer.

## Implementation Details

* **CGO Wrapper**: The bindings use a C shim (`wrapper.cpp` / `wrapper.h`) to bridge the C++ class methods to C-compatible functions that CGO can call.
* **Struct Layout**: Go structs are manually defined to match the memory layout of the C++ structs exactly. This allows for zero-copy casting in the C wrapper layer, making frame transmission highly efficient.
* **Thread Safety**: The underlying C++ SDK claims thread safety for device operations. However, CGO calls block the calling Go goroutine. For high-performance rendering loops, ensure your frame generation logic does not bottleneck on the `WriteFrame` call.
