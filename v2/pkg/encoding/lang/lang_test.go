// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package lang

import (
	"fmt"
	"strings"
	"testing"
)

func Test(t *testing.T) {
	greetings := Alt{
		EnUS: "hello",
		FrFR: "bonjour",
		JaJP: "こんにちは",
		ZhCN: "你好",
	}
	w := new(strings.Builder)
	ut := func(
		t *testing.T,
		lang LANG,
		want string,
	) {
		t.Helper()
		w.Reset()
		Precedence = []LANG{lang}
		fmt.Fprint(w, greetings)
		if got := w.String(); got != want {
			t.Errorf("%q != %q", got, want)
		} else {
			t.Log(got)
		}
	}
	for k, s := range greetings {
		t.Run(string(k), func(t *testing.T) {
			ut(t, k, s)
		})
	}
}
