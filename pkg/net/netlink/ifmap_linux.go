// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netlink

import "fmt"

// uapi/linux/if_link.h: struct rtnl_link_ifmap
type RtnlLinkIfmap struct {
	MemStart uint64
	MemEnd   uint64
	BaseAddr uint64
	IRQ      uint16
	DMA      uint8
	Port     uint8
}

func (p *RtnlLinkIfmap) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "mem_start:%#x,", p.MemStart)
	fmt.Fprint(w, "mem_end:%#x,", p.MemEnd)
	fmt.Fprint(w, "base_addr:%#x,", p.BaseAddr)
	fmt.Fprint(w, "irq:%#x,", p.IRQ)
	fmt.Fprint(w, "dma:%#x,", p.DMA)
	fmt.Fprint(w, "port:%d,", p.Port)
}
