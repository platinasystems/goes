// Created by cgo -godefs - DO NOT EDIT
// cgo -godefs -- -I./arch/x86/include -I./arch/x86/include/generated -I./include -I./arch/x86/include/uapi -I./arch/x86/include/generated/uapi -I./include/uapi -I./include/generated/uapi -include ./include/linux/kconfig.h ./drivers/net/ethernet/xeth/go/src/xeth/godefs.go

package xeth

const (
	IFNAMSIZ		= 0x10
	SizeofJumboFrame	= 0x2600
	SizeofMsgHdr		= 0x10
	SizeofBreakMsg		= 0x10
	SizeofStatMsg		= 0x30
	SizeofEthtoolFlagsMsg	= 0x28
	SizeofEthtoolDumpMsg	= 0x10
)

type Hdr struct {
	Z64	uint64
	Z32	uint32
	Z16	uint16
	Z8	uint8
	Op	uint8
}
type Stat struct {
	Index	uint64
	Count	uint64
}
