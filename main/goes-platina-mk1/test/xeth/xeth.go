// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xeth

import (
	"fmt"
	"testing"

	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/mk1"
)

var Suite = test.Suite{
	Name: "xeth",
	Tests: test.Tests{
		&test.Unit{"bad-names", badNames},
		&test.Unit{"good-names", goodNames},
	},
}

func badNames(t *testing.T) {
	assert := test.Assert{t}
	base := mk1.Base()
	for _, name := range []string{
		"eth-",
		"eth-n-0",
		fmt.Sprintf("eth_%d-%d", base+3, base),
		fmt.Sprintf("eth-%d-%d", base+3, base+4),
		fmt.Sprintf("eth-%d-%d", base+33, base),
		"xeth",
		fmt.Sprintf("xeth%d", base+33),
		fmt.Sprintf("xeth%d_%d", base+3, base),
		fmt.Sprintf("xeth%d-%d", base+3, base+4),
		"xethbr.",
		"xethbr.n",
		"xethbr.0",
		"xethbr.4095",
	} {
		assert.ProgramErr(true, test.Self{},
			"ip", "link", "add", name, "type", "platina-mk1")
	}
}

func goodNames(t *testing.T) {
	assert := test.Assert{t}
	cleanup := test.Cleanup{t}
	base := mk1.Base()
	for _, name := range []string{
		"xethbr.100",
		"xethbr.101",
		fmt.Sprintf("xeth%d.100u", base+1),
		fmt.Sprintf("xeth%d.100u", base+2),
		fmt.Sprintf("xeth%d.100t", base+2),
	} {
		assert.Program(test.Self{},
			"ip", "link", "add", name, "type", "platina-mk1")
		defer cleanup.Program(test.Self{},
			"ip", "link", "del", name)
	}
}
