/* SPDX-License-Identifier: GPL-2.0+ WITH Linux-syscall-note */

package iflink

const (
	IFNAMSIZ    = 16
	IFALIASZ    = 256
	ALTIFNAMSIZ = 128
)

//go:generate stringer -output=ziff_string.go -type=NetDeviceFlag -trimprefix=IFF_ .
type NetDeviceFlag uint32

const (
	IFF_UP NetDeviceFlag = 1 << iota
	IFF_BROADCAST
	IFF_DEBUG
	IFF_LOOPBACK
	IFF_POINTOPOINT
	IFF_NOTRAILERS
	IFF_RUNNING
	IFF_NOARP
	IFF_PROMISC
	IFF_ALLMULTI
	IFF_MASTER
	IFF_SLAVE
	IFF_MULTICAST
	IFF_PORTSEL
	IFF_AUTOMEDIA
	IFF_DYNAMIC
	IFF_LOWER_UP
	IFF_DORMANT
	IFF_ECHO
)

const IFF_VOLATILE = IFF_LOOPBACK |
	IFF_POINTOPOINT |
	IFF_BROADCAST |
	IFF_ECHO |
	IFF_MASTER |
	IFF_SLAVE |
	IFF_RUNNING |
	IFF_LOWER_UP |
	IFF_DORMANT

const (
	IF_GET_IFACE = 1 + iota
	IF_GET_PROTO
)

//go:generate stringer -output=zif_iface_string.go -type=IfIface -trimprefix=IF_IFACE_ .
type IfIface uint32

const (
	IF_IFACE_V35 IfIface = 0x1000 + iota
	IF_IFACE_V24
	IF_IFACE_X21
	IF_IFACE_T1
	IF_IFACE_E1
	IF_IFACE_SYNC_SERIAL
	IF_IFACE_X21D
)

//go:generate stringer -output=zif_proto_string.go -type=IfProto -trimprefix=IF_PROTO_ .
type IfProto uint32

const (
	IF_PROTO_HDLC IfProto = 0x2000 + iota
	IF_PROTO_PPP
	IF_PROTO_CISCO
	IF_PROTO_FR
	IF_PROTO_FR_ADD_PVC
	IF_PROTO_FR_DEL_PVC
	IF_PROTO_X25
	IF_PROTO_HDLC_ETH
	IF_PROTO_FR_ADD_ETH_PVC
	IF_PROTO_FR_DEL_ETH_PVC
	IF_PROTO_FR_PVC
	IF_PROTO_FR_ETH_PVC
	IF_PROTO_RAW
)

//go:generate stringer -output=zif_oper_string.go -type=IfOper -trimprefix=IF_OPER_ .
type IfOper uint8

const (
	IF_OPER_UNKNOWN IfOper = iota
	IF_OPER_NOTPRESENT
	IF_OPER_DOWN
	IF_OPER_LOWERLAYERDOWN
	IF_OPER_TESTING
	IF_OPER_DORMANT
	IF_OPER_UP
)

//go:generate stringer -output=zif_link_mode_string.go -type=IfLinkMode -trimprefix=IF_LINK_MODE_ .
type IfLinkMode uint8

const (
	IF_LINK_MODE_DEFAULT IfLinkMode = iota
	IF_LINK_MODE_DORMANT
	IF_LINK_MODE_TESTING
)

type IfMap struct {
	MemStart uint64
	MemEnd   uint64
	BaseAddr uint64
	IRQ      uint16
	DMA      uint8
	Port     uint8
}

type IfSettings struct {
	Type uint32
	Size uint32
	Ptr  uintptr
}
