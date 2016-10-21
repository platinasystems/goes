// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elib

import (
	"testing"
)

func TestHeap(t *testing.T) {
	c := testHeap{
		iterations: 1000,
		nObjects:   10,
		verbose:    0,
	}
	err := runHeapTest(&c)
	if err != nil {
		t.Error(err)
	}
}
