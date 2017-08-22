// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build elog_gtk

package loop

import (
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/elib/elog/elogview"

	"sync"
)

var elogviewWaitGroup sync.WaitGroup

func (l *Loop) ViewEventLog(v *elog.View) {
	wg := &elogviewWaitGroup
	wg.Wait()
	wg.Add(1)
	go func() {
		elogview.View(v, elogview.Config{Width: 1200, Height: 750})
		wg.Done()
	}()
}
