// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package accumulate

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func Test(t *testing.T) {
	const (
		s0 = "The quick brown fox jumped "
		s1 = "over the lazy dog's back"
		n  = len(s0) + len(s1)
	)
	w := New(ioutil.Discard)
	defer w.Fini()
	fmt.Fprint(w, s0)
	fmt.Fprint(w, s1)
	if int(w.N) != n {
		t.Fatalf("accumulated %d, expected %d", w.N, n)
	}
}
