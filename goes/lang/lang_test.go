// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package lang

import "testing"

var hello = Alt{
	EnUS: "hello",
	FrFR: "bonjour",
	JaJP: "こんにちは",
	ZhCN: "你好",
}

func Test(t *testing.T) {
	Lang = ""
	t.Log("default:", hello)
	for lang, expect := range hello {
		Lang = lang
		if s := hello.String(); s != expect {
			t.Fatalf("%q != %q", s, expect)
		} else {
			t.Logf("%s: %s", lang, s)
		}
	}
}
