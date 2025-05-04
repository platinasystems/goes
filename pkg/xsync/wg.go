//go:build !go1.25

package xsync

import "sync"

// This wrapper implements go1.25's WaitGroup.Go
type WaitGroup struct{ sync.WaitGroup }

func (wg *WaitGroup) Go(f func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		f()
	}()
}
