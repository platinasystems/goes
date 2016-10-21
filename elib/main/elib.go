// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/platinasystems/go/elib"

	"fmt"
)

func main() {
	if true {
		elib.HeapTest()
	}
	if false {
		elib.SparseTest()
	}
	if false {
		elib.FibHeapTest()
	}
	if false {
		{
			b := elib.Bitmap(0)
			b = b.Set(64)
			fmt.Printf("%s %+v\n", b, elib.Bitmaps)
			b = b.Free()
			fmt.Printf("%s %+v\n", b, elib.Bitmaps)
			b = elib.Bitmap((1 << 3) | (1 << 13))
			fmt.Printf("%s %+v\n", b, elib.Bitmaps)
			b = elib.Bitmaps.Set(b, 128)
			fmt.Printf("%s %+v\n", b, elib.Bitmaps)

			c := b.Dup().Set(12)
			fmt.Printf("%s %+v\n", c, elib.Bitmaps)

			d := c.Dup()
			for x := ^uint(0); d.Next(&x); {
				fmt.Printf("%d\n", x)
			}

			c = c.Free()
			fmt.Printf("%s %+v\n", c, elib.Bitmaps)
			d.Free()
			b.Free()
			fmt.Printf("%s %+v\n", c, elib.Bitmaps)
		}

		{
			zero := elib.Bitmap(0)
			c := zero.Set(1).Set(2)
			c = c.Orx(128)
			d := zero.Orx(256).Or(c)
			fmt.Printf("%s %s %+v\n", c, d, elib.Bitmaps)
			c.Free()

			d = d.AndNotx(256)
			d = d.AndNotx(128)
			fmt.Printf("%s %+v\n", d, elib.Bitmaps)
		}
	}
}
