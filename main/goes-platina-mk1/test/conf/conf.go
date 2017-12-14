// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package conf

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/platinasystems/go/internal/test"
	"github.com/platinasystems/go/main/goes-platina-mk1/test/port2port"
)

func New(t *testing.T, name, text string) []byte {
	assert := test.Assert{t}
	assert.Helper()
	tmpl, err := template.New(name).Parse(text)
	assert.Nil(err)
	buf := new(bytes.Buffer)
	assert.Nil(tmpl.Execute(buf, port2port.Conf))
	return buf.Bytes()
}
