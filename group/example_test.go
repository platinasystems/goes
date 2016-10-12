// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package group

import "fmt"

func ExampleAdm() {
	group := Parse()
	fmt.Println("adm.Gid:", group["adm"].Gid())
	// Output:
	// adm.Gid: 4
}

func ExampleFoobar() {
	group := Parse()
	fmt.Println("foobar.Gid:", group["foobar"].Gid())
	// Output:
	// foobar.Gid: 0
}
