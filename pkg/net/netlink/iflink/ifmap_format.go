// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package iflink

import "fmt"

func (p *IfMap) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "mem_start:%#x,", p.MemStart)
	fmt.Fprint(w, "mem_end:%#x,", p.MemEnd)
	fmt.Fprint(w, "base_addr:%#x,", p.BaseAddr)
	fmt.Fprint(w, "irq:%#x,", p.IRQ)
	fmt.Fprint(w, "dma:%#x,", p.DMA)
	fmt.Fprint(w, "port:%d,", p.Port)
}
