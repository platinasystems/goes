// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package accumulate

import (
	"io/ioutil"
	"testing"
)

func Test(t *testing.T) {
	var n int64
	acc := New(ioutil.Discard)
	defer acc.Fini()

	for _, s := range []string{
		"The quick brown fox jumped ",
		"over the lazy dog's back",
	} {
		acc.WriteString(s)
		n += int64(len(s))
	}
	if err := acc.Error(); err != nil {
		t.Fatal(err)
	}
	if sum := acc.Total(); sum != n {
		t.Fatalf("accumulated %d, expected %d", sum, n)
	}
}
