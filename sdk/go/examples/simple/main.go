// Example: Simple
//
// This example demonstrates the most basic usage of the Helios DAC Go bindings.
// It generates a pre-calculated animation of a horizontal line scanning vertically.
//
// Concepts shown:
// - Initializing the DAC
// - Discovering devices
// - Basic frame generation loop
// - Single-threaded synchronous writing
package main

import (
	"fmt"
	"time"

	"github.com/Grix/helios_dac/sdk/go/helios"
)

func main() {
	const (
		numFramesInLoop   = 20
		numPointsPerFrame = 1500
		pointsPerSecond   = 40000

		// Coordinate and color constants
		maxValue    = 0xFFFF // Max 16-bit value (65535)
		colorNormal = 0xD0FF // A nice lavender/purple color intensity
	)

	fmt.Println("Generating frames...")
	// Generate frames
	frames := make([][]helios.PointHighRes, numFramesInLoop)
	for i := 0; i < numFramesInLoop; i++ {
		frames[i] = make([]helios.PointHighRes, numPointsPerFrame)
		y := uint16(i * maxValue / numFramesInLoop)
		for j := 0; j < numPointsPerFrame; j++ {
			var x uint16
			if j < numPointsPerFrame/2 {
				x = uint16(j * maxValue / (numPointsPerFrame / 2))
			} else {
				x = uint16(maxValue - ((j - (numPointsPerFrame / 2)) * maxValue / (numPointsPerFrame / 2)))
			}

			frames[i][j] = helios.PointHighRes{
				X: x,
				Y: y,
				R: colorNormal, // From main.cpp
				G: maxValue,
				B: colorNormal,
				// I: maxValue, // Not supported in PointHighRes
			}
		}
	}

	dac := helios.NewDAC()
	defer dac.Close()

	fmt.Println("Scanning for devices...")
	numDevices := dac.OpenDevices()
	fmt.Printf("Found %d DACs:\n", numDevices)

	if numDevices == 0 {
		fmt.Println("No DACs found (exiting example)")
		return
	}

	for j := 0; j < numDevices; j++ {
		name := dac.GetName(j)
		fmt.Printf("- %s: FW: %d\n", name, dac.GetFirmwareVersion(j))
	}

	fmt.Println("Outputting animation... (Press Ctrl+C to stop)")

	// Output loop
	frameIdx := 0
	for {
		for j := 0; j < numDevices; j++ {
			// Poll status
			attempts := 0
			for attempts < 1024 {
				status := dac.GetStatus(j)
				if status == 1 {
					dac.WriteFrameHighResolution(j, pointsPerSecond, 0, frames[frameIdx%numFramesInLoop])
					break
				} else if status < 0 {
					fmt.Printf("Error polling device %d: %d\n", j, status)
					break
				}
				attempts++
			}
		}
		frameIdx++

		// Prevent tight loop if devices are not ready
		if numDevices > 0 {
			// In main.cpp there is no sleep in the write loop, just polling.
			// But valid status polling loop breaks immediately once ready.
			// If not ready, it retries.
		} else {
			time.Sleep(100 * time.Millisecond)
		}
	}
}
