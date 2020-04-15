// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package semaphore

import "sync"

type Nothing struct{}

type Semaphore chan Nothing

type OneTime struct {
	Semaphore
	sync.Once
}

var Stop = OneTime{
	Semaphore: make(Semaphore),
}

func (ch Semaphore) Signal() {
	ch <- Nothing{}
}

func (ch Semaphore) Wait() bool {
	_, ok := <-ch
	return ok
}

func (ot *OneTime) Signal() {
	ot.Do(func() { close(ot.Semaphore) })
}
