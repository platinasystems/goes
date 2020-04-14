// Copyright © 2018-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbg

import (
	"bytes"
	"os"
	"testing"
)

func Test(t *testing.T) {
	buf := new(bytes.Buffer)
	Writer(buf)
	for _, style := range []Style{
		NoOp,
		Plain,
		FileLine,
		Func,
	} {
		style.Log("printed")
		style.Logf("%s", "formatted")
		style.Log(nil, "not", "printed")
		style.Logf("%v %s %s", nil, "not", "printed")
		style.Log(os.ErrInvalid, "printed")
		style.Logf("%v %s", os.ErrInvalid, "formatted")
	}
	want := []byte(`
printed
formatted
invalid argument printed
invalid argument formatted
dbg_test.go:22: printed
dbg_test.go:23: formatted
dbg_test.go:26: invalid argument printed
dbg_test.go:27: invalid argument formatted
github.com/platinasystems/goes/external/dbg.Test() printed
github.com/platinasystems/goes/external/dbg.Test() formatted
github.com/platinasystems/goes/external/dbg.Test() invalid argument printed
github.com/platinasystems/goes/external/dbg.Test() invalid argument formatted
`)[1:]
	if bytes.Compare(want, buf.Bytes()) != 0 {
		t.Fatal("\ngot:\n" + buf.String() + "\nwant:\n" + string(want))
	} else if testing.Verbose() {
		os.Stdout.Write(buf.Bytes())
	}
	if NoOp.Log(nil) != nil {
		t.Fatal("not nil")
	}
	if NoOp.Log(os.ErrInvalid) != os.ErrInvalid {
		t.Fatal("not invalid")
	}
}
