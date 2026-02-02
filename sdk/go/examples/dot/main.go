package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"time"

	"github.com/Grix/helios_dac/sdk/go/helios"
)

const (
	// Resolution
	galvoBitDepth  = 12
	galvoFullScale = 1 << galvoBitDepth // 4096
	galvoMaxCoord  = galvoFullScale - 1 // 4095

	// Hardware specs for dynamic latency calculation based on galvanometer physics.
	stepResponseSmallAngle = 250 * time.Microsecond  // 0.1° step
	stepResponseLargeAngle = 1000 * time.Microsecond // 40° optical step

	// Default PPS (Points Per Second).
	defaultPPS = 50000
)

func main() {
	var x, y, radius, pps int
	flag.IntVar(&x, "x", 2048, "X coordinate of the dot center (0-4095)")
	flag.IntVar(&y, "y", 2048, "Y coordinate of the dot center (0-4095)")
	flag.IntVar(&radius, "radius", 84, "Radius of the dot in galvo units (approx 30mm)")
	flag.IntVar(&pps, "pps", defaultPPS, "Points per second")
	flag.Parse()

	if x < 0 || x > galvoMaxCoord || y < 0 || y > galvoMaxCoord {
		log.Fatalf("Coordinates out of bounds: (%d, %d). Must be 0-%d", x, y, galvoMaxCoord)
	}

	fmt.Printf("Generating simple circle at (%d, %d) radius %d...\n", x, y, radius)

	// Target 15ms frame time to fit within shorter exposure times of a 30fps camera.
	targetFrameTime := 15 * time.Millisecond
	pointBudget := int(targetFrameTime.Seconds() * float64(pps))

	// Allocate budget. We need to reserve some for flyback.
	// Flyback from radius to center is small distance (radius).
	// Let's reserve 20% for flyback and safety.
	featureBudget := int(float64(pointBudget) * 0.8)

	// 1. Generate Feature (Center -> Ring Start -> Ring)
	points := getFeaturePoints(float64(x), float64(y), radius, featureBudget, pps)

	// 2. Flyback (Ring End -> Center)
	// We must return to center because the feature generator expects to start settling at center.
	if len(points) > 0 {
		lastPt := points[len(points)-1]
		// Determine where we ended.
		// Construct travel points from Last Point -> Center (x, y)
		flyback := getTravelPoints(float64(lastPt.X), float64(lastPt.Y), float64(x), float64(y), pps)
		points = append(points, flyback...)
	}

	// OPTIMIZATION: Fill the buffer.
	// Replicate the frame to reduce USB overhead and ensure continuous playback.
	// We target ~2000 points (approx 40ms at 50kpps) which is well within Helios usually 4096+ point buffer.
	// This ensures the DAC doesn't run dry between updates.
	const targetBufferPoints = 2000
	singleLoop := make([]helios.Point, len(points))
	copy(singleLoop, points)

	for len(points) < targetBufferPoints {
		points = append(points, singleLoop...)
	}

	frameDuration := time.Duration(float64(len(points)) / float64(pps) * float64(time.Second))
	fmt.Printf("Generated %d points (~%s per frame)\n", len(points), frameDuration)

	dac := helios.NewDAC()
	defer dac.Close()

	fmt.Println("Scanning for devices...")
	numDevices := dac.OpenDevices()
	fmt.Printf("Found %d DACs\n", numDevices)

	if numDevices == 0 {
		return
	}

	// Setup interrupt handler
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	fmt.Println("Projecting dot... Press Ctrl-C to stop.")

	// Playback Rate Limiting
	// Enforce min replay interval to avoid buffer underruns/partial frames
	minReplayInterval := time.Duration(float64(frameDuration) * 0.9)
	lastWriteTime := time.Time{}

	ticker := time.NewTicker(1 * time.Millisecond)
	defer ticker.Stop()

	running := true
	for running {
		select {
		case <-stop:
			running = false
		case <-ticker.C:
			if time.Since(lastWriteTime) < minReplayInterval {
				continue
			}

			// Poll all devices
			anyWritten := false
			for i := 0; i < numDevices; i++ {
				if dac.GetStatus(i) == 1 {
					dac.WriteFrame(i, pps, 0, points)
					anyWritten = true
				}
			}
			if anyWritten {
				lastWriteTime = time.Now()
			}
		}
	}

	for i := 0; i < numDevices; i++ {
		dac.Stop(i)
	}
	dac.CloseDevices()
}

// getTravelPoints generates blanking points to move the laser from start to end coordinates.
func getTravelPoints(startX, startY, endX, endY float64, pps int) []helios.Point {
	// Calculate Euclidean distance for the jump
	dist := math.Hypot(endX-startX, endY-startY)

	ratio := dist / float64(galvoFullScale)
	if ratio > 1.0 {
		ratio = 1.0
	}

	// Dynamic Latency: Interpolate settling time between small (300µs) and large (1000µs) moves.
	reqTime := stepResponseSmallAngle + time.Duration(float64(stepResponseLargeAngle-stepResponseSmallAngle)*ratio)

	travelPoints := int(math.Ceil(reqTime.Seconds() * float64(pps)))
	if travelPoints < 1 {
		travelPoints = 1
	}

	var points []helios.Point
	for k := 1; k <= travelPoints; k++ {
		t := float64(k) / float64(travelPoints)
		// SmoothStep (S-curve) interpolation to minimize mechanical shock/jerk.
		alpha := t * t * (3.0 - 2.0*t)

		ix := startX + (endX-startX)*alpha
		iy := startY + (endY-startY)*alpha
		points = append(points, helios.Point{
			X: uint16(ix), Y: uint16(iy), R: 0, G: 0, B: 0, I: 0,
		})
	}

	// Settling Dwell: 150µs dead time to ensure absolute stability before laser enable.
	settleTime := 150 * time.Microsecond
	settlePoints := int(math.Ceil(settleTime.Seconds() * float64(pps)))
	if settlePoints < 1 {
		settlePoints = 1
	}

	for k := 0; k < settlePoints; k++ {
		points = append(points, helios.Point{
			X: uint16(endX), Y: uint16(endY), R: 0, G: 0, B: 0, I: 0,
		})
	}

	return points
}

// getFeaturePoints generates the visible ring pattern.
func getFeaturePoints(cx, cy float64, dotRadius int, pointBudget int, pps int) []helios.Point {
	var points []helios.Point

	// Dynamic Production: Circle Only
	availableForDrawing := pointBudget - 1
	if availableForDrawing < 10 {
		return points
	}

	// 1. Move from Center (blanked) to Ring Start (Angle 0).
	// We assume the laser is historically at Center (cx, cy).
	ringStart := helios.Point{X: uint16(cx + float64(dotRadius)), Y: uint16(cy), R: 0, G: 0, B: 0, I: 0}
	travel := getTravelPoints(cx, cy, float64(ringStart.X), float64(ringStart.Y), pps)
	points = append(points, travel...)

	// 2. Draw Ring
	// Use remaining budget for the ring
	ringPts := availableForDrawing - len(travel)
	if ringPts < 10 {
		ringPts = 10
	}

	for i := 1; i <= ringPts; i++ {
		t := float64(i) / float64(ringPts)

		theta := 2.0 * math.Pi * t

		px := uint16(cx + float64(dotRadius)*math.Cos(theta))
		py := uint16(cy + float64(dotRadius)*math.Sin(theta))

		points = append(points, helios.Point{X: px, Y: py, R: 255, G: 255, B: 255, I: 255})
	}

	return points
}
