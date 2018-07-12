// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ethtool

import (
	"io/ioutil"
	"time"

	"github.com/platinasystems/go/internal/test"
	"gopkg.in/yaml.v2"
)

var Map map[string][]string

func Init(assert test.Assert) {
	assert.Helper()
	Map = make(map[string][]string)
	b, err := ioutil.ReadFile("testdata/ethtool.yaml")
	if err != nil {
		return
	}
	err = yaml.Unmarshal(b, Map)
	assert.Nil(err)
	for ifname, args := range Map {
		cmd := make([]interface{}, 4+len(args))
		cmd[0] = 2 * time.Second
		cmd[1] = "ethtool"
		cmd[2] = "-s"
		cmd[3] = ifname
		for i, arg := range args {
			cmd[4+i] = arg
		}
		assert.Program(cmd...)
	}
}
