// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main_test

import (
	"testing"

	"github.com/platinasystems/go/goes/cmd/ip"
	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/main/ip/test/link"
)

func Test(t *testing.T) {
	if test.Goes {
		test.Exec(ip.New().Main)
	}
	test.Suite{
		{"link", link.Suite},
	}.Run(t)
}
