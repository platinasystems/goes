// Created by cgo -godefs - DO NOT EDIT
// cgo -godefs -- -I./arch/x86/include -I./arch/x86/include/generated -I./include -I./arch/x86/include/uapi -I./arch/x86/include/generated/uapi -I./include/uapi -I./include/generated/uapi -include ./include/linux/kconfig.h drivers/net/ethernet/xeth/godefs.go

package xeth

const (
	IFNAMSIZ			= 0x10
	ETH_ALEN			= 0x6
	SizeofJumboFrame		= 0x2600
	SizeofMsg			= 0x10
	SizeofMsgBreak			= 0x10
	SizeofMsgCarrier		= 0x18
	SizeofMsgDumpFibinfo		= 0x10
	SizeofMsgDumpIfinfo		= 0x10
	SizeofMsgEthtoolFlags		= 0x18
	SizeofMsgEthtoolSettings	= 0x38
	SizeofMsgIfa			= 0x20
	SizeofMsgIfinfo			= 0x48
	SizeofNextHop			= 0x18
	SizeofMsgFibentry		= 0x28
	SizeofMsgNeighUpdate		= 0x38
	SizeofMsgSpeed			= 0x18
	SizeofMsgStat			= 0x28
)

type Msg struct {
	Z64	uint64
	Z32	uint32
	Z16	uint16
	Z8	uint8
	Kind	uint8
}

type MsgBreak struct {
	Z64	uint64
	Z32	uint32
	Z16	uint16
	Z8	uint8
	Kind	uint8
}

type MsgCarrier struct {
	Z64	uint64
	Z32	uint32
	Z16	uint16
	Z8	uint8
	Kind	uint8
	Ifindex	int32
	Flag	uint8
	Pad	[3]uint8
}

type MsgEthtoolFlags struct {
	Z64	uint64
	Z32	uint32
	Z16	uint16
	Z8	uint8
	Kind	uint8
	Ifindex	int32
	Flags	uint32
}

type MsgEthtoolSettings struct {
	Z64				uint64
	Z32				uint32
	Z16				uint16
	Z8				uint8
	Kind				uint8
	Ifindex				int32
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
}

type NextHop struct {
	Ifindex	int32
	Weight	int32
	Flags	uint32
	Gw	uint32
	Scope	uint8
	Pad	[7]uint8
}

type MsgFibentry struct {
	Z64	uint64
	Z32	uint32
	Z16	uint16
	Z8	uint8
	Kind	uint8
	Net	uint64
	Address	uint32
	Mask	uint32
	Event	uint8
	Nhs	uint8
	Tos	uint8
	Type	uint8
	Id	uint32
}

type MsgIfa struct {
	Z64	uint64
	Z32	uint32
	Z16	uint16
	Z8	uint8
	Kind	uint8
	Ifindex	int32
	Event	uint32
	Address	uint32
	Mask	uint32
}

type MsgIfinfo struct {
	Z64		uint64
	Z32		uint32
	Z16		uint16
	Z8		uint8
	Kind		uint8
	Ifname		[16]uint8
	Net		uint64
	Ifindex		int32
	Iflinkindex	int32
	Flags		uint32
	Id		uint16
	Addr		[6]uint8
	Portindex	int16
	Subportindex	int8
	Devtype		uint8
	Portid		int16
	Reason		uint8
	Pad		[5]uint8
}

type MsgNeighUpdate struct {
	Z64	uint64
	Z32	uint32
	Z16	uint16
	Z8	uint8
	Kind	uint8
	Net	uint64
	Ifindex	int32
	Family	uint8
	Len	uint8
	Pad0	[2]uint8
	Dst	[16]uint8
	Lladdr	[6]uint8
	Pad	[2]uint8
}

type MsgSpeed struct {
	Z64	uint64
	Z32	uint32
	Z16	uint16
	Z8	uint8
	Kind	uint8
	Ifindex	int32
	Mbps	uint32
}

type MsgStat struct {
	Z64	uint64
	Z32	uint32
	Z16	uint16
	Z8	uint8
	Kind	uint8
	Ifindex	int32
	Pad	[4]uint8
	Index	uint64
	Count	uint64
}
