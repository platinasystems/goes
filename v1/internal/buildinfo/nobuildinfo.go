// Copyright © 2019-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build !go1.12

package buildinfo

import "fmt"

type BuildInfo struct{}

func New() BuildInfo {
	return BuildInfo{}
}

func (bi BuildInfo) Format(f fmt.State, c rune) {
	f.Write(unavailable)
}

func (bi BuildInfo) Version() string {
	return Unavailable
}
