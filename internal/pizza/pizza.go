// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Slice args at boundary. For example,
//
//	pl := pizza.New("|")
//	pl.Slice("ls", "-lR", "|", "more")
//	// pl.Slices == [][]string{
//	// 	[]string{"ls", "-lR"},
//	// 	[]string{"more"},
//	// }
//	pl.Reset()
//	pl.Slice("ls", "-lR", "|")
//	// pl.Slices == [][]string{
//	// 	[]string{"ls", "-lR"},
//	// }
//	if pl.More {
//		pl.Slice("more")
//	}
//	// pl.Slices == [][]string{
//	// 	[]string{"ls", "-lR"},
//	// 	[]string{"more"},
//	// }
package pizza

type Pizza struct {
	Boundary string
	// More is true if the last arg of the last Slice was Boundary.
	More   bool
	Slices [][]string
}

func New(boundary string) *Pizza { return &Pizza{Boundary: boundary} }

func (p *Pizza) Slice(args ...string) {
	p.More = false
	p.Slices = append(p.Slices, make([]string, 0, len(args)))
	for len(args) > 0 {
		if args[0] == p.Boundary {
			args = args[1:]
			if len(args) == 0 {
				p.More = true
				break
			}
			p.Slices = append(p.Slices,
				make([]string, 0, len(args)))
		} else {
			i := len(p.Slices) - 1
			p.Slices[i] = append(p.Slices[i], args[0])
			args = args[1:]
		}
	}
}

func (p *Pizza) Reset() {
	p.More = false
	for i := 0; i < len(p.Slices); i++ {
		p.Slices[0] = p.Slices[0][:0]
	}
	p.Slices = p.Slices[:0]
}
