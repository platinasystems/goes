// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package semaphore

import (
	"testing"
	"time"
)

func TestSemaphore(t *testing.T) {
	ch := make(Semaphore)
	go func() {
		defer close(ch)
		for i := 0; i < 3; i++ {
			ch.Signal()
		}
	}()
	for i := 1; ch.Wait(); i++ {
		t.Log("signal", i)
	}
}

func TestStop(t *testing.T) {
	go func() {
		Stop.Signal()
		Stop.Signal()
	}()
	for i := 1; Stop.Wait(); i++ {
		if i > 1 {
			t.Fatal(i)
		}
	}
}

func TestSelect(t *testing.T) {
	ch := make(Semaphore)
	go func() {
		defer close(ch)
		ch.Signal()
	}()
	dl := time.NewTimer(time.Second)
	for {
		select {
		case <-dl.C:
			t.Fatal("timeout")
		case <-ch:
			if !dl.Stop() {
				<-dl.C
			}
			return
		}
	}
}
