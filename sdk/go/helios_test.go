package helios

import "testing"

func TestSmoke(t *testing.T) {
	dac := NewDAC()
	defer dac.Close()

	if dac == nil {
		t.Fatal("Failed to create HeliosDac instance")
	}

	// This should run without crashing.
	n := dac.OpenDevices()
	t.Logf("Found %d devices", n)

	// Check that we can call methods safely even if 0 devices
	// (Actual logic verification not required, just bindings)
}
