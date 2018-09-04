// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netport

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/platinasystems/go/internal/test"
	"gopkg.in/yaml.v2"
)

var PortByNetPort map[string]string
var NetPortByPort map[string]string

func Init(assert test.Assert) {
	assert.Helper()
	PortByNetPort = make(map[string]string)
	NetPortByPort = make(map[string]string)
	b, err := ioutil.ReadFile("testdata/netport.yaml")
	assert.Nil(err)
	err = yaml.Unmarshal(b, PortByNetPort)
	assert.Nil(err)
	for netport, port := range PortByNetPort {
		NetPortByPort[port] = netport
		_, err = os.Stat(filepath.Join("/sys/class/net", port))
		if err != nil {
			assert.Fatal(err)
		}
	}
}
