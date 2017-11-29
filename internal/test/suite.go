// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import "testing"

// Suite of tests
type Suite []struct {
	Name string
	Func func(*testing.T)
}

// Run test suite
func (suite Suite) Run(t *testing.T) {
	for _, x := range suite {
		if t.Failed() {
			break
		}
		t.Run(x.Name, x.Func)
	}
}
