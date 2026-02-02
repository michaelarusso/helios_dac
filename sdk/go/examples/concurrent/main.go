// Example: Concurrent Generator/Writer
//
// This example demonstrates a high-performance architecture for laser generation.
// It separates the frame generation (CPU intensive) from the frame output (IO sensitive).
//
// Concepts shown:
// - Concurrency: Using Go channels to pipe frames from generator to writer.
// - Double Buffering: Using a buffered channel to decouple generation frame rate from output.
// - Performance: Using runtime.LockOSThread() to reduce OS scheduler jitter on the output loop.
// - Dynamic Generation: Calculating frames on-the-fly based on wall-clock time.
package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/Grix/helios_dac/sdk/go/helios"
)

const (
	PPS       = 30000
	Center    = 2048
	ScanRange = 1000 // How far left/right to move
	ScanSpeed = 2.0  // Radians per second
	FrameRate = 30   // Target FPS for generation
)

// generateFrames continually calculates new frames based on time and sends them to the channel.
func generateFrames(ctx context.Context, framesChan chan<- []helios.Point) {
	ticker := time.NewTicker(time.Second / FrameRate)
	defer ticker.Stop()

	startTime := time.Now()

	fmt.Println("Generator: Started")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Generator: Stopping")
			return
		case t := <-ticker.C:
			// Calculate animation state based on elapsed time
			elapsed := t.Sub(startTime).Seconds()

			// Pattern: A vertical line that scans left and right (sine wave)
			xOffset := math.Sin(elapsed*ScanSpeed) * float64(ScanRange)
			lineX := float64(Center) + xOffset

			// Line dimensions (Vertical)
			yTop := float64(Center + ScanRange)
			yBottom := float64(Center - ScanRange)

			// Simple line generation (fixed point count)
			// Frame duration ~ 1/30s = 33ms. @ 30k PPS ~ 1000 points.
			numPoints := PPS / FrameRate
			frame := make([]helios.Point, numPoints)

			for i := 0; i < numPoints; i++ {
				progress := float64(i) / float64(numPoints-1)

				// X is constant for the whole frame (vertical line)
				x := lineX

				// Y interpolates from Bottom to Top
				y := yBottom + (yTop-yBottom)*progress

				frame[i] = helios.Point{
					X: uint16(x),
					Y: uint16(y),
					R: 0,
					G: 255, // Green Line
					B: 0,
					I: 255,
				}
			}

			// Non-blocking send if possible, otherwise block until writer is ready.
			// Ideally the buffer size prevents blocking unless writer falls behind.
			select {
			case framesChan <- frame:
			case <-ctx.Done():
				return
			}
		}
	}
}

// outputLoop manages the DAC communication.
// It is locked to an OS thread to reduce scheduler jitter for the timing-sensitive output loop.
func outputLoop(ctx context.Context, dac *helios.DAC, framesChan <-chan []helios.Point) {
	// 1. Lock this goroutine to an OS thread.
	// This ensures that the Go scheduler doesn't migrate this execution context
	// across different OS threads, improving cache locality and reducing latency spikes.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	fmt.Println("Writer: Started (Locked to OS Thread)")

	// We use the last received frame if 'framesChan' is empty but the DAC needs data immediately.
	// This behaves like a "hold" buffer.
	var currentFrame []helios.Point

	// Initial wait for first frame
	select {
	case f := <-framesChan:
		currentFrame = f
	case <-ctx.Done():
		return
	}

	for {
		// Check for cancellation
		select {
		case <-ctx.Done():
			fmt.Println("Writer: Stopping")
			// Stop output before exiting
			// Note: dac.Stop() stops the stream.
			// Iterate all devices to stop them.
			// (Assuming 1 device for simplicity, or we can scan again)
			// Ideally we just exit, and the defer Close() in main handles cleanup.
			return
		default:
		}

		// Try to get a newer frame if one is available (drain buffer to get latest)
		// This ensures we always output the freshest animation state.
	DrainLoop:
		for {
			select {
			case f := <-framesChan:
				currentFrame = f
			default:
				// Channel empty, stop draining
				break DrainLoop
			}
		}

		// Write to all ready devices
		// In a real production app, you might track devices more robustly.
		// For this example, we'll blindly write to index 0 if it's ready.
		// (Assuming at least one device exists)

		status := dac.GetStatus(0)
		if status == 1 {
			dac.WriteFrame(0, PPS, 0, currentFrame)
		} else {
			// If not ready, sleep briefly to yield CPU
			// High-performance loops might busy-wait or sleep progressively.
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func main() {
	// Initialize DAC
	dac := helios.NewDAC()
	defer dac.Close()

	if dac.OpenDevices() == 0 {
		fmt.Println("No devices found. Exiting.")
		return
	}

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel buffer size of 2 acts as a double-buffer.
	// Generator produces next frame while Writer consumes current.
	framesChan := make(chan []helios.Point, 2)

	// Start Generator
	go generateFrames(ctx, framesChan)

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutdown signal received...")
		cancel() // Stop goroutines
	}()

	// Start Writer
	// Run on main thread to avoid libusb threading issues on some platforms
	outputLoop(ctx, dac, framesChan)

	// Give slight time for cleanup printouts
	time.Sleep(100 * time.Millisecond)

	dac.Stop(0)
	dac.CloseDevices()
	fmt.Println("Devices closed. Bye!")
}
