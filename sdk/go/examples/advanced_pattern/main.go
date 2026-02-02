// Example: Advanced Pattern (Vector Graphics)
//
// This example demonstrates how to draw vector shapes (a triangle) with proper
// laser control techniques to ensure sharp corners and no visible travel lines.
//
// Concepts shown:
// - Blanking: Turning off the laser (RGB=0) while moving between shapes (Vector move).
// - Dwell: Holding the laser at a point for multiple samples to let physical mirrors settle.
// - Timing: Calculating point counts based on duration and PPS (Points Per Second).
// - Interpolation: Generating lines between two points.
package main

import (
	"fmt"
	"math"
	"time"

	"github.com/Grix/helios_dac/sdk/go/helios"
)

// Config calculation constants
const (
	MaxXY    = 4095
	MaxColor = 255
	PPS      = 30000 // Points per second
)

// GenerateLine creates a line of points from start to end with a specific color.
// The number of points is determined by the duration and PPS.
func GenerateLine(start, end helios.Point, duration time.Duration, on bool) []helios.Point {
	numPoints := int(float64(PPS) * duration.Seconds())
	if numPoints < 1 {
		numPoints = 1
	}

	points := make([]helios.Point, numPoints)

	for i := 0; i < numPoints; i++ {
		t := float64(i) / float64(numPoints-1) // 0.0 to 1.0

		// Linear interpolation for position
		x := float64(start.X) + t*(float64(end.X)-float64(start.X))
		y := float64(start.Y) + t*(float64(end.Y)-float64(start.Y))

		p := helios.Point{
			X: uint16(math.Round(x)),
			Y: uint16(math.Round(y)),
		}

		if on {
			p.R = end.R
			p.G = end.G
			p.B = end.B
			p.I = end.I
		}
		// Else (blanking), colors stay 0

		points[i] = p
	}
	return points
}

// GenerateDwell creates a set of points at a static location.
// Useful for sharp corners or waiting.
func GenerateDwell(pos helios.Point, duration time.Duration, on bool) []helios.Point {
	numPoints := int(float64(PPS) * duration.Seconds())
	if numPoints < 1 {
		numPoints = 1
	}

	points := make([]helios.Point, numPoints)
	for i := 0; i < numPoints; i++ {
		p := pos
		if !on {
			p.R, p.G, p.B, p.I = 0, 0, 0, 0
		}
		points[i] = p
	}
	return points
}

func main() {
	dac := helios.NewDAC()
	defer dac.Close()

	fmt.Println("Scanning for devices...")
	if dac.OpenDevices() == 0 {
		fmt.Println("No devices found. Exiting.")
		return
	}

	// Define our shape: A triangle with dwell at corners and blanking moves
	// 1. Move (Blanked) to Bottom-Left
	// 2. Line to Top-Center
	// 3. Dwell at Top-Center
	// 4. Line to Bottom-Right
	// 5. Line to Bottom-Left

	pLowLeft := helios.Point{X: 1000, Y: 1000, G: 255, I: 255}
	pTopCenter := helios.Point{X: 2048, Y: 3500, G: 255, I: 255}
	pLowRight := helios.Point{X: 3096, Y: 1000, G: 255, I: 255}

	var frame []helios.Point

	// Part 1: Blanking move from logic origin (0,0) to start (5ms)
	// Important: To prevent trails, turn laser off while moving to start position
	frame = append(frame, GenerateLine(helios.Point{X: 0, Y: 0}, pLowLeft, 5*time.Millisecond, false)...)

	// Part 2: Draw Left Side (Line) - 20ms
	frame = append(frame, GenerateLine(pLowLeft, pTopCenter, 20*time.Millisecond, true)...)

	// Part 3: Dwell at Top Corner (Accentuate the point) - 2ms
	frame = append(frame, GenerateDwell(pTopCenter, 2*time.Millisecond, true)...)

	// Part 4: Draw Right Side - 20ms
	frame = append(frame, GenerateLine(pTopCenter, pLowRight, 20*time.Millisecond, true)...)

	// Part 5: Dwell Bottom Right - 1ms
	frame = append(frame, GenerateDwell(pLowRight, 1*time.Millisecond, true)...)

	// Part 6: Draw Bottom Side - 20ms
	frame = append(frame, GenerateLine(pLowRight, pLowLeft, 20*time.Millisecond, true)...)

	fmt.Printf("Generated frame with %d points.\n", len(frame))

	// Frame Time Calculation
	// Total Points: ~2000 (approx)
	// PPS: 30000
	// Frame Time = 2000 / 30000 = 0.066s = 66ms = ~15 FPS
	frameDuration := time.Duration(len(frame)) * time.Second / PPS
	fmt.Printf("Expected frame duration: %v\n", frameDuration)

	fmt.Println("Outputting pattern... (Ctrl+C to stop)")

	// Output Loop:
	// To prevent buffer underrun, we simply need to write the next frame
	// as soon as the status is ready. The hardware buffer is small, so we keep feeding it.

	ticker := time.NewTicker(frameDuration) // Optional: Try to sync to frame time roughly
	defer ticker.Stop()

	for {
		// In a real high-perf loop, we might not sleep, but just poll GetStatus.
		for i := 0; i < dac.GetStatus(0); i++ {
			// Check if ready (GetStatus returns 1 if ready)
			// Actually typical Helios usage is:
			// if (status == 1) SendFrame()

			status := dac.GetStatus(0)
			if status == 1 {
				// Send the pre-calculated frame
				dac.WriteFrame(0, PPS, 0, frame)
			} else {
				// Prevent CPU spin if not ready
				time.Sleep(1 * time.Millisecond)
			}
		}

		// Note on Underrun:
		// If 'frameDuration' is very small (< 10ms) and your PC can't keep up,
		// the laser might flicker.
		// If 'frameDuration' is large (> 100ms), the refresh rate is low (flicker).
		// Ideal frames are often limited to ~1000-2000 points for smooth 30-60FPS.
	}
}
