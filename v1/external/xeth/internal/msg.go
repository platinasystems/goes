// Copyright © 2018-2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package internal

import (
	"fmt"
	"reflect"
	"unsafe"
)

type MsgKind uint8

func (h *MsgHeader) Set(kind uint8) {
	h.Z64 = 0
	h.Z32 = 0
	h.Z16 = 0
	h.Version = MsgVersion
	h.Kind = kind
}

func (h *MsgHeader) Validate(buf []byte) error {
	if len(buf) < SizeofMsg {
		return fmt.Errorf("msg too small %v", buf)
	}
	if h.Z64 != 0 || h.Z32 != 0 || h.Z16 != 0 {
		return fmt.Errorf("msg has non-zero header fields  %+v", *h)
	}
	if h.Version != MsgVersion {
		return fmt.Errorf("msg version %d, expect %d",
			h.Version, MsgVersion)
	}
	var exact, min int
	switch h.Kind {
	case MsgKindFibEntry:
		min = SizeofMsgFibEntry
	case MsgKindFib6Entry:
		min = SizeofMsgFib6Entry
	case MsgKindEthtoolLinkModesSupported,
		MsgKindEthtoolLinkModesAdvertising,
		MsgKindEthtoolLinkModesLPAdvertising:
		min = SizeofMsgEthtoolLinkModes
	case MsgKindBreak:
		exact = SizeofMsgBreak
	case MsgKindChangeUpperXid:
		exact = SizeofMsgChangeUpperXid
	case MsgKindEthtoolFlags:
		exact = SizeofMsgEthtoolFlags
	case MsgKindEthtoolSettings:
		exact = SizeofMsgEthtoolSettings
	case MsgKindIfa:
		exact = SizeofMsgIfa
	case MsgKindIfa6:
		exact = SizeofMsgIfa6
	case MsgKindIfInfo:
		exact = SizeofMsgIfInfo
	case MsgKindNeighUpdate:
		exact = SizeofMsgNeighUpdate
	case MsgKindNetNsAdd:
		exact = SizeofMsgNetNs
	case MsgKindNetNsDel:
		exact = SizeofMsgNetNs
	default:
		return fmt.Errorf("msg kind %d unsupported", h.Kind)
	}
	if min > 0 && len(buf) < min {
		return fmt.Errorf("msg length %d, expect >= %d", len(buf), min)
	}
	if exact > 0 && len(buf) < exact {
		return fmt.Errorf("msg length %d, expect %d", len(buf), exact)
	}
	return nil
}

func (msg *MsgFibEntry) NextHops() []NextHop {
	ptr := unsafe.Pointer(msg)
	nhs := int(msg.Nhs)
	return *(*[]NextHop)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(ptr) + uintptr(SizeofMsgFibEntry),
		Len:  nhs,
		Cap:  nhs,
	}))
}

func (msg *MsgFib6Entry) Siblings() []NextHop6 {
	nsiblings := int(msg.Nsiblings)
	if nsiblings == 0 {
		return []NextHop6{}
	}
	ptr := unsafe.Pointer(msg)
	return *(*[]NextHop6)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(ptr) + uintptr(SizeofMsgFib6Entry),
		Len:  nsiblings,
		Cap:  nsiblings,
	}))
}
