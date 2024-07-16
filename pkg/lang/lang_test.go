// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package lang

import "testing"

func TestLang(t *testing.T) {
	greetings := Alt{
		EnUS: "hello",
		FrFR: "bonjour",
		JaJP: "こんにちは",
		ZhCN: "你好",
	}
	for lang, want := range greetings {
		t.Run(string(lang), func(t *testing.T) {
			Precedence = []LANG{lang}
			if got := greetings.String(); got != want {
				t.Fatalf("%q != %q", got, want)
			}
		})
	}
}
