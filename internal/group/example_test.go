// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package group

import "fmt"

func ExampleParse_adm() {
	group := Parse()
	fmt.Println("adm.Gid:", group["adm"].Gid())
	// Output:
	// adm.Gid: 4
}

func ExampleParse_foobar() {
	group := Parse()
	fmt.Println("foobar.Gid:", group["foobar"].Gid())
	// Output:
	// foobar.Gid: 0
}
