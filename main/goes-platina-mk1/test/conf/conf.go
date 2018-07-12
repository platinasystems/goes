// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package conf

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
	"text/template"

	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/internal/test/netport"
)

func New(t *testing.T, fn string) []byte {
	assert := test.Assert{t}
	assert.Helper()
	text, err := ioutil.ReadFile(fn)
	assert.Nil(err)
	name := strings.TrimSuffix(fn, ".tmpl")
	tmpl, err := template.New(name).Parse(string(text))
	assert.Nil(err)
	buf := new(bytes.Buffer)
	assert.Nil(tmpl.Execute(buf, netport.Map))
	return buf.Bytes()
}
