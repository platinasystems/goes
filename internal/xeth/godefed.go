// Created by cgo -godefs - DO NOT EDIT
// cgo -godefs -- -I./arch/x86/include -I./arch/x86/include/generated -I./include -I./arch/x86/include/uapi -I./arch/x86/include/generated/uapi -I./include/uapi -I./include/generated/uapi -include ./include/linux/kconfig.h ./drivers/net/ethernet/xeth/go/src/xeth/godefs.go

package xeth

const (
	IFNAMSIZ			= 0x10
	SizeofJumboFrame		= 0x2600
	SizeofMsg			= 0x10
	SizeofMsgBreak			= 0x10
	SizeofMsgStat			= 0x30
	SizeofMsgEthtoolFlags		= 0x28
	SizeofMsgEthtoolSettings	= 0x48
	SizeofMsgDumpIfinfo		= 0x10
	SizeofMsgCarrier		= 0x28
	SizeofMsgSpeed			= 0x28
	SizeofMsgIfindex		= 0x30
	SizeofMsgIfa			= 0x30
)

type Msg struct {
	Z64	uint64
	Z32	uint32
	Z16	uint16
	Z8	uint8
	Kind	uint8
}
type Ifmsg struct {
	Z64	uint64
	Z32	uint32
	Z16	uint16
	Z8	uint8
	Kind	uint8
	Ifname	[16]uint8
}
type MsgBreak struct {
	Z64	uint64
	Z32	uint32
	Z16	uint16
	Z8	uint8
	Kind	uint8
}
type MsgStat struct {
	Z64	uint64
	Z32	uint32
	Z16	uint16
	Z8	uint8
	Kind	uint8
	Ifname	[16]uint8
	Index	uint64
	Count	uint64
}
type MsgEthtoolFlags struct {
	Z64		uint64
	Z32		uint32
	Z16		uint16
	Z8		uint8
	Kind		uint8
	Ifname		[16]uint8
	Flags		uint32
	Pad_cgo_0	[4]byte
}
type MsgEthtoolSettings struct {
	Z64				uint64
	Z32				uint32
	Z16				uint16
	Z8				uint8
	Kind				uint8
	Ifname				[16]uint8
	Speed				uint32
	Duplex				uint8
	Port				uint8
	Phy_address			uint8
	Autoneg				uint8
	Mdio_support			uint8
	Eth_tp_mdix			uint8
	Eth_tp_mdix_ctrl		uint8
	Link_mode_masks_nwords		int8
	Link_modes_supported		[2]uint32
	Link_modes_advertising		[2]uint32
	Link_modes_lp_advertising	[2]uint32
	Pad_cgo_0			[4]byte
}
type MsgCarrier struct {
	Z64		uint64
	Z32		uint32
	Z16		uint16
	Z8		uint8
	Kind		uint8
	Ifname		[16]uint8
	Flag		uint8
	Pad_cgo_0	[7]byte
}
type MsgSpeed struct {
	Z64		uint64
	Z32		uint32
	Z16		uint16
	Z8		uint8
	Kind		uint8
	Ifname		[16]uint8
	Mbps		uint32
	Pad_cgo_0	[4]byte
}
type MsgIfindex struct {
	Z64	uint64
	Z32	uint32
	Z16	uint16
	Z8	uint8
	Kind	uint8
	Ifname	[16]uint8
	Ifindex	uint64
	Net	uint64
}
type MsgIfa struct {
	Z64		uint64
	Z32		uint32
	Z16		uint16
	Z8		uint8
	Kind		uint8
	Ifname		[16]uint8
	Event		uint32
	Address		uint32
	Mask		uint32
	Pad_cgo_0	[4]byte
}
