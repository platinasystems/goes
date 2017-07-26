// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ethernet

import (
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/vnet"

	"fmt"
)

func (a *Address) String() string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", a[0], a[1], a[2], a[3], a[4], a[5])
}

func (a *Address) Parse(in *parse.Input) {
	var b [3]vnet.Uint16
	switch {
	case in.Parse("%x:%x:%x:%x:%x:%x", &a[0], &a[1], &a[2], &a[3], &a[4], &a[5]):
	case in.Parse("%x.%x.%x", &b[0], &b[1], &b[2]):
		a[0], a[1] = uint8(b[0]>>8), uint8(b[0])
		a[2], a[3] = uint8(b[1]>>8), uint8(b[1])
		a[4], a[5] = uint8(b[2]>>8), uint8(b[2])
	default:
		panic(parse.ErrInput)
	}
}

func (h *Header) String() (s string) {
	return fmt.Sprintf("%s: %s -> %s", h.GetType().String(), h.Src.String(), h.Dst.String())
}

func (h *Header) Parse(in *parse.Input) {
	if !in.ParseLoose("%v: %v -> %v", &h.Type, &h.Src, &h.Dst) {
		panic(parse.ErrInput)
	}
}

type ParseHeader struct {
	h Header
	v []VlanHeader
}

func (h *ParseHeader) Parse(in *parse.Input) (innerType Type) {
	h.h.Parse(in)
	for !in.End() {
		var vh VlanHeader
		if in.Parse("%v", &vh) {
			h.v = append(h.v, vh)
		} else {
			break
		}
	}

	innerType = h.h.Type
	if len(h.v) > 0 {
		h.h.Type = h.v[0].Type
		for i := range h.v {
			t := innerType
			if i+1 < len(h.v) {
				t = h.v[i+1].Type
			}
			h.v[i].Type = t
		}
	}
	return
}
func (h *ParseHeader) Sizeof() uint { return SizeofHeader + uint(len(h.v))*SizeofVlanHeader }
func (h *ParseHeader) Write(b []byte) {
	h.h.Write(b)
	for i := range h.v {
		h.v[i].Write(b[SizeofHeader+i*SizeofVlanHeader:])
	}
}

func (h *VlanHeader) Parse(sup_in *parse.Input) {
	var (
		in  parse.Input
		id  vnet.Uint16
		pri vnet.Uint16
	)
	if sup_in.Parse("vlan %v", &in) {
		var (
			tag vnet.Uint16
			tp  Type
		)
		tp = TYPE_VLAN.FromHost()
		for !in.End() {
			switch {
			case in.Parse("%v", &id):
				if id > 0xfff {
					panic(parse.ErrInput)
				}
				tag = (tag &^ 0xfff) | id
			case in.Parse("cfi"):
				tag |= 1 << 12
			case in.Parse("pri%*ority %d", &pri):
				if pri > 8 {
					panic(parse.ErrInput)
				}
				tag = (tag &^ (7 << 13)) | pri<<13
			case in.Parse("tpid %v", &tp):
			default:
				panic(parse.ErrInput)
			}
		}
		h.Type = tp
		h.Tag = VlanTag(tag).FromHost()
	} else {
		panic(parse.ErrInput)
	}
}

func (h *VlanHeader) String() (s string) {
	if h.Type.ToHost() != TYPE_VLAN {
		s = h.Type.ToHost().String() + ": "
	}
	s += fmt.Sprintf("vlan %d", h.Tag.Id())
	return
}

func (v *VlanTag) String() string { return fmt.Sprintf("0x%04x", vnet.Uint16(*v).ToHost()) }

func (v *Address) MaskedString(r vnet.MaskedStringer) (s string) {
	m := r.(*Address)
	s = v.String() + "/" + m.String()
	return
}

func (v Type) MaskedString(r vnet.MaskedStringer) (s string) {
	m := r.(Type)
	if m == 0xffff {
		s = v.String()
	} else {
		s += fmt.Sprintf("0x%x/%x", v.ToHost(), m.ToHost())
	}
	return
}

func (v VlanTag) MaskedString(r vnet.MaskedStringer) (s string) {
	m := r.(VlanTag)
	s = v.String()
	if m != 0 {
		s += fmt.Sprintf("/%s", m.String())
	}
	return
}
