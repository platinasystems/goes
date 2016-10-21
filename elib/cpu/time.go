// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cpu

import (
	"math"
	"sync"
	"time"
)

const (
	Second      float64 = 1
	Minute              = 60 * Second
	Hour                = 60 * Minute
	Day                 = 24 * Hour
	Microsecond         = 1e-6 * Second
	Millisecond         = 1e-3 * Second
)

type Time uint64

var (
	// Ticks per second of event timer (and inverse).
	cyclesPerSec, secsPerCycle float64
	cyclesOnce                 sync.Once
)

func (dt Time) Seconds() float64 {
	estimateOnce()
	return float64(dt) * secsPerCycle
}

func (t *Time) Cycles(dt float64) {
	estimateOnce()
	*t = Time(dt * cyclesPerSec)
}

func measureCPUCyclesPerSec(wait float64) (freq float64) {
	var t0 [2]Time
	var t1 [2]int64
	t1[0] = time.Now().UnixNano()
	t0[0] = TimeNow()
	time.Sleep(time.Duration(1e9 * wait))
	t1[1] = time.Now().UnixNano()
	t0[1] = TimeNow()
	freq = 1e9 * float64(t0[1]-t0[0]) / float64(t1[1]-t1[0])
	return
}

func round(x, unit float64) float64 {
	return unit * math.Floor(.5+x/unit)
}

func estimateOnce() {
	cyclesOnce.Do(func() {
		go estimateFrequency(1e-4, 1e6, 5e5)
	})
	// Wait until estimateFrequency is done.
	for secsPerCycle == 0 {
		time.Sleep(10 * time.Microsecond)
	}
}

func estimateFrequency(dt, unit, tolerance float64) {
	var sum, sum2, ave, rms, n float64
	for n = float64(1); true; n++ {
		f := measureCPUCyclesPerSec(dt)
		sum += f
		sum2 += f * f
		ave = sum / n
		rms = math.Sqrt((sum2/n - ave*ave) / n)
		if n >= 16 && rms < tolerance {
			break
		}
	}

	cyclesPerSec = round(ave, unit)
	secsPerCycle = 1 / cyclesPerSec
	return
}

// Aid for timing blocks of code.
type Timing [2]Time

// Number of cpu clock cycles per operation.
func (t *Timing) ClocksPer(ops uint) float64 { return float64(t[1]-t[0]) / float64(ops) }

// Number of operations per second.
func (t *Timing) PerSecond(ops uint) float64 { return float64(ops) / (t[1] - t[0]).Seconds() }
