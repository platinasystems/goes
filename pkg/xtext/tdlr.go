// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xtext

import "fmt"

// [TopDownLeftRight.Format] prints Text in the style of “ls -C”.
type TopDownLeftRight struct {
	PageSize
	Text []string
}

func (tdlr TopDownLeftRight) Format(w fmt.State, verb rune) {
	var width uint = 8
	for i := 0; i < len(tdlr.Text); {
		if n := len(tdlr.Text[i]); n > 24 {
			fmt.Println(tdlr.Text[i])
			if i < len(tdlr.Text)-1 {
				copy(tdlr.Text[i:], tdlr.Text[i+1:])
			}
			tdlr.Text = tdlr.Text[:len(tdlr.Text)-1]
		} else {
			i++
			if uint(n) >= width {
				width += 8
			}
		}
	}

	n := uint(len(tdlr.Text))
	tcols := (tdlr.PageSize.Width - 1) / width
	tlines := n / tcols
	if tlines*tcols < n {
		tlines += 1
	}

	for line := uint(0); line < tlines; line++ {
		for tcol := uint(0); tcol < tcols; tcol++ {
			if k := line + (tcol * tlines); k < n {
				fmt.Printf("%-*s", width, tdlr.Text[k])
			}
		}
		fmt.Println()
	}
}
