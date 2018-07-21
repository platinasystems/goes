// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"testing"
)

// Cleanup wraps a testing.Test or Benchmark with Program
type Cleanup struct {
	testing.TB
}

// Program runs a cleanup task to its End
func (cleanup Cleanup) Program(options ...interface{}) {
	if !cleanup.Failed() {
		cleanup.Helper()
	}
	p, err := Begin(cleanup.TB, options...)
	if err == nil {
		err = p.End()
	}
	if err != nil {
		cleanup.Log(err)
	}
}
