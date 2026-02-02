# Helios Go Examples

This directory contains examples of how to use the Helios DAC Go bindings to generate laser patterns.

## 1. Simple (`simple/`)

**Pattern**: A horizontal scanner line that moves vertically down the projection area.

```text
    Frame 1        Frame N/2       Frame N
  +---------+    +---------+    +---------+
  | <=====> |    |         |    |         |
  |         |    | <=====> |    |         |
  |         |    |         |    | <=====> |
  +---------+    +---------+    +---------+
   Y = Top        Y = Mid        Y = Bottom
```

**Run usage**:

```bash
bazel run //sdk/go/examples/simple
```

## 2. Advanced Pattern (`advanced_pattern/`)

**Pattern**: A geometric triangle with dwell points at the corners to ensure sharp edges.

```text
       (2048, 3500)
           /\
          /  \
         /    \
        /      \
       /        \
      /__________\
(1000,1000)    (3096,1000)
```

**Features demonstrated**:

* **Vector Graphics**: Drawing lines between coordinates.
* **Blanking**: Turning the laser off (`RGBI=0`) while moving to the start position to avoid "travel lines".
* **Dwell**: Repeating points at corners to allow the physical galvo mirrors to settle, creating sharp corners instead of rounded curves.
* **Timing**: Calculating frame duration based on point count and PPS.

**Run usage**:

```bash
bazel run //sdk/go/examples/advanced_pattern
```

## 3. Concurrent (`concurrent/`)

**Pattern**: A vertical green line that scans horizontally left and right (Sine wave).

```text
    Frame 1        Frame N/2       Frame N
  +---------+    +---------+    +---------+
  |    |    |    | |       |    |       | |
  |    |    | -> | |       | -> |       | |
  |    |    |    | |       |    |       | |
  +---------+    +---------+    +---------+
    Center          Left           Right
```

**Features demonstrated**:

* **Concurrency**: Separating frame generation (CPU work) from frame transmission (IO work) using Go channels.
* **OS Thread Locking**: Using `runtime.LockOSThread()` in the output goroutine to ensure consistent timing and prevent OS scheduler jitter, which is critical for smooth laser projection.
* **Double Buffering**: Using a buffered channel to minimize blocking between the generator and the writer.

**Run usage**:

```bash
bazel run //sdk/go/examples/concurrent
```
