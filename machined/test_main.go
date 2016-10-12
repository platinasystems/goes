// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// +build tests

package machined

// Test machined's recover from an InfoProvider crash
func (*test) Main(...string) error {
	_ = []string{}[0]
	return nil
}
