// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netport

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/platinasystems/go/internal/test"
	"gopkg.in/yaml.v2"
)

var Map map[string]string

func Init(assert test.Assert) {
	assert.Helper()
	Map = make(map[string]string)
	b, err := ioutil.ReadFile("testdata/netport.yaml")
	assert.Nil(err)
	err = yaml.Unmarshal(b, Map)
	assert.Nil(err)
	for _, port := range Map {
		_, err = os.Stat(filepath.Join("/sys/class/net", port))
		if err == nil {
			continue
		}
		assert.Program(2*time.Second, test.Self{},
			"ip", "link", "add", port, "type", "platina-mk1")
	}
}
