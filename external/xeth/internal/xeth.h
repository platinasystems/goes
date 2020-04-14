/**
 * XETH side-band channel protocol.
 *
 * SPDX-License-Identifier: GPL-2.0
 * Copyright(c) 2018-2019 Platina Systems, Inc.
 *
 * Contact Information:
 * sw@platina.com
 * Platina Systems, 3180 Del La Cruz Blvd, Santa Clara, CA 95054
 */

#ifndef __XETH_UAPI_H
#define __XETH_UAPI_H

#ifdef IFNAMSIZE
# define XETH_IFNAMSIZ IFNAMSIZ
#else
# define XETH_IFNAMSIZ 16
#endif

#ifdef ETH_ALEN
# define XETH_ALEN ETH_ALEN
#else
# define XETH_ALEN 6
#endif

#ifdef VLAN_VID_MASK
# define XETH_VLAN_VID_MASK VLAN_VID_MASK
#else
# define XETH_VLAN_VID_MASK 0x0fff
#endif

#ifdef VLAN_N_VID
# define XETH_VLAN_N_VID VLAN_N_VID
#else
# define XETH_VLAN_N_VID 4096
#endif

enum xeth_msg_version {
	XETH_MSG_VERSION = 2,
};

enum {
	XETH_SIZEOF_JUMBO_FRAME = 9728,
};

enum xeth_vlan_ifla {
	XETH_VLAN_IFLA_UNSPEC,
	XETH_VLAN_IFLA_VID,
	XETH_VLAN_N_IFLA,
};

enum xeth_dev_kind {
	XETH_DEV_KIND_UNSPEC,
	XETH_DEV_KIND_PORT,
	XETH_DEV_KIND_VLAN,
	XETH_DEV_KIND_BRIDGE,
	XETH_DEV_KIND_LAG,
};

enum xeth_msg_kind {
	XETH_MSG_KIND_BREAK,
	XETH_MSG_KIND_LINK_STAT,
	XETH_MSG_KIND_ETHTOOL_STAT,
	XETH_MSG_KIND_ETHTOOL_FLAGS,
	XETH_MSG_KIND_ETHTOOL_SETTINGS,
	XETH_MSG_KIND_ETHTOOL_LINK_MODES_SUPPORTED,
	XETH_MSG_KIND_ETHTOOL_LINK_MODES_ADVERTISING,
	XETH_MSG_KIND_ETHTOOL_LINK_MODES_LP_ADVERTISING,
	XETH_MSG_KIND_DUMP_IFINFO,
	XETH_MSG_KIND_CARRIER,
	XETH_MSG_KIND_SPEED,
	XETH_MSG_KIND_IFINFO,
	XETH_MSG_KIND_IFA,
	XETH_MSG_KIND_IFA6,
	XETH_MSG_KIND_DUMP_FIBINFO,
	XETH_MSG_KIND_FIBENTRY,
	XETH_MSG_KIND_FIB6ENTRY,
	XETH_MSG_KIND_NEIGH_UPDATE,
	XETH_MSG_KIND_CHANGE_UPPER_XID,
};

enum xeth_msg_carrier_flag {
	XETH_CARRIER_OFF,
	XETH_CARRIER_ON,
};

enum xeth_msg_ifinfo_reason {
	XETH_IFINFO_REASON_NEW,
	XETH_IFINFO_REASON_DEL,
	XETH_IFINFO_REASON_UP,
	XETH_IFINFO_REASON_DOWN,
	XETH_IFINFO_REASON_DUMP,
	XETH_IFINFO_REASON_REG,
	XETH_IFINFO_REASON_UNREG,
};

struct xeth_msg_header {
	uint64_t z64;
	uint32_t z32;	
	uint16_t z16;
	uint8_t version;
	uint8_t kind;
};

struct xeth_msg {
	struct xeth_msg_header header;
};

struct xeth_msg_break {
	struct xeth_msg_header header;
};

struct xeth_msg_carrier {
	struct xeth_msg_header header;
	uint32_t xid;
	uint8_t flag;
	uint8_t pad[3];
};

struct xeth_msg_change_upper_xid {
	struct xeth_msg_header header;
	uint32_t upper;
	uint32_t lower;
	uint8_t linking;
	uint8_t pad[7];
};

struct xeth_msg_dump_fibinfo {
	struct xeth_msg_header header;
};

struct xeth_msg_dump_ifinfo {
	struct xeth_msg_header header;
};

struct xeth_msg_ethtool_flags {
	struct xeth_msg_header header;
	uint32_t xid;
	uint32_t flags;
};

struct xeth_msg_ethtool_settings {
	struct xeth_msg_header header;
	uint32_t xid;
	uint32_t speed;
	uint8_t duplex;
	uint8_t port;
	uint8_t phy_address;
	uint8_t autoneg;
	uint8_t mdio_support;
	uint8_t eth_tp_mdix;
	uint8_t eth_tp_mdix_ctrl;
	uint8_t pad;
};


struct xeth_msg_ethtool_link_modes {
	struct xeth_msg_header header;
	uint32_t xid;
	uint32_t reserved;
	uint64_t modes;
};

struct xeth_next_hop {
	int32_t ifindex;
	int32_t weight;
	uint32_t flags;
	__be32 gw;
	uint8_t scope;
	uint8_t pad[7];
};

struct xeth_msg_fibentry {
	struct xeth_msg_header header;
	uint64_t net;
	__be32 address;
	__be32 mask;
	uint8_t event;
	uint8_t nhs;
	uint8_t tos;
	uint8_t type;
	uint32_t table;
	struct xeth_next_hop nh[];
};

struct xeth_next_hop6 {
	int32_t ifindex;
	int32_t weight;
	uint32_t flags;
	uint32_t reserved;
	uint8_t gw[16];
};

struct xeth_msg_fib6entry {
	struct xeth_msg_header header;
	uint64_t net;
	uint8_t address[16];
	uint8_t length;
	uint8_t event;
	uint8_t nsiblings;
	uint8_t type;
	uint32_t table;
	struct xeth_next_hop6 nh;
	struct xeth_next_hop6 siblings[];
};

struct xeth_msg_ifa {
	struct xeth_msg_header header;
	uint32_t xid;
	uint32_t event;
	__be32 address;
	__be32 mask;
};

struct xeth_msg_ifa6 {
	struct xeth_msg_header header;
	uint32_t xid;
	uint32_t event;
	uint8_t address[16];
	uint8_t length;
	uint8_t pad[7];
};

struct xeth_msg_ifinfo {
	struct xeth_msg_header header;
	uint32_t xid;
	uint32_t reserved;
	uint8_t ifname[XETH_IFNAMSIZ];
	uint64_t net;
	int32_t ifindex;
	uint32_t flags;
	uint8_t addr[XETH_ALEN];
	uint8_t kind;
	uint8_t reason;
};

struct xeth_msg_neigh_update {
	struct xeth_msg_header header;
	uint64_t net;
	int32_t ifindex;
	uint8_t family;
	uint8_t len;
	uint16_t reserved;
	uint8_t dst[16];
	uint8_t lladdr[XETH_ALEN];
	uint8_t pad[8-XETH_ALEN];
};

struct xeth_msg_speed {
	struct xeth_msg_header header;
	uint32_t xid;
	uint32_t mbps;
};

struct xeth_msg_stat {
	struct xeth_msg_header header;
	uint32_t xid;
	uint32_t index;
	uint64_t count;
};

#endif /* __XETH_UAPI_H */
