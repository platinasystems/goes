// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xeth

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/platinasystems/go/internal/test"
)

var alpha = flag.Bool("test.alpha", false, "this is a zero based alpha system")

func TestNames(t *testing.T) {
	assert := test.Assert{t}
	cleanup := test.Cleanup{t}
	b, err := ioutil.ReadFile("/proc/net/unix")
	assert.Nil(err)
	if bytes.Index(b, []byte("@platina-mk1/xeth")) >= 0 {
		assert.Program("rmmod", "platina-mk1")
	}
	assert.Program("modprobe", "platina-mk1")
	base := 1
	if *alpha {
		base = 0
	}
	for _, name := range []string{
		fmt.Sprintf("eth-%d-%d", base, base),
		fmt.Sprintf("xeth%d", base+1),
		fmt.Sprintf("xeth%d", base+2),
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
