// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package indent

import (
	"fmt"
	"os"
)

func Example() {
	w := New(os.Stdout, "\t")
	fmt.Fprintln(w, "level 0")
	Increase(w)
	fmt.Fprintln(w, "level 1")
	Increase(w)
	fmt.Fprintln(w, "level 2")
	Increase(w)
	fmt.Fprintln(w, "level 3")
	Decrease(w)
	fmt.Fprintln(w, "back to level 2")
	Decrease(w)
	fmt.Fprintln(w, "back to level 1")
	Increase(w)
	fmt.Fprintln(w, "another level 2")
	Decrease(w)
	Decrease(w)
	fmt.Fprintln(w, "back level 0")
	Decrease(w)
	fmt.Fprintln(w, "caught -1")
	Increase(w)
	fmt.Fprintln(w, "level 1")
	// Output:
	// level 0
	// 	level 1
	// 		level 2
	// 			level 3
	// 		back to level 2
	// 	back to level 1
	// 		another level 2
	// back level 0
	// caught -1
	// 	level 1
}
