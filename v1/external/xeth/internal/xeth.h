/* SPDX-License-Identifier: GPL-2.0 WITH Linux-syscall-note */
/**
 * XETH side-band channel protocol.
 *
 * Copyright(c) 2018-2019 Platina Systems, Inc.
 *
 * Contact Information:
 * sw@platina.com
 * Platina Systems, 3180 Del La Cruz Blvd, Santa Clara, CA 95054
 */

#ifndef __XETH_UAPI_H
#define __XETH_UAPI_H

#include <linux/types.h>

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

enum xeth_encap {
	XETH_ENCAP_VLAN = 0,
	XETH_ENCAP_VPLS,
};

enum xeth_encap_vid_bit {
      XETH_ENCAP_VLAN_VID_BIT = 12,
      XETH_ENCAP_VPLS_VID_BIT = 20,
};

enum xeth_encap_vid_mask {
      XETH_ENCAP_VLAN_VID_MASK = (1 << XETH_ENCAP_VLAN_VID_BIT) - 1,
      XETH_ENCAP_VPLS_VID_MASK = (1 << XETH_ENCAP_VPLS_VID_BIT) - 1,
};

enum xeth_port_ifla {
	XETH_PORT_IFLA_UNSPEC,
	XETH_PORT_IFLA_XID,	/* u32 */
	XETH_PORT_N_IFLA,
};

enum xeth_vlan_ifla {
	XETH_VLAN_IFLA_UNSPEC,
	XETH_VLAN_IFLA_VID,	/* u16 */
	XETH_VLAN_N_IFLA,
};

enum xeth_lb_ifla {
	XETH_LB_IFLA_UNSPEC,
	XETH_LB_IFLA_CHANNEL,	/* u8 */
	XETH_LB_N_IFLA,
};

enum xeth_mux_ifla {
	XETH_MUX_IFLA_UNSPEC,
	XETH_MUX_IFLA_ENCAP,	/* u8 */
	XETH_MUX_N_IFLA,
};

enum xeth_dev_kind {
	XETH_DEV_KIND_UNSPEC,
	XETH_DEV_KIND_PORT,
	XETH_DEV_KIND_VLAN,
	XETH_DEV_KIND_BRIDGE,
	XETH_DEV_KIND_LAG,
	XETH_DEV_KIND_LB,
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
	XETH_MSG_KIND_NETNS_ADD,
	XETH_MSG_KIND_NETNS_DEL,
};

enum xeth_link_stat {
	XETH_LINK_STAT_RX_PACKETS,
	XETH_LINK_STAT_TX_PACKETS,
	XETH_LINK_STAT_RX_BYTES,
	XETH_LINK_STAT_TX_BYTES,
	XETH_LINK_STAT_RX_ERRORS,
	XETH_LINK_STAT_TX_ERRORS,
	XETH_LINK_STAT_RX_DROPPED,
	XETH_LINK_STAT_TX_DROPPED,
	XETH_LINK_STAT_MULTICAST,
	XETH_LINK_STAT_COLLISIONS,
	XETH_LINK_STAT_RX_LENGTH_ERRORS,
	XETH_LINK_STAT_RX_OVER_ERRORS,
	XETH_LINK_STAT_RX_CRC_ERRORS,
	XETH_LINK_STAT_RX_FRAME_ERRORS,
	XETH_LINK_STAT_RX_FIFO_ERRORS,
	XETH_LINK_STAT_RX_MISSED_ERRORS,
	XETH_LINK_STAT_TX_ABORTED_ERRORS,
	XETH_LINK_STAT_TX_CARRIER_ERRORS,
	XETH_LINK_STAT_TX_FIFO_ERRORS,
	XETH_LINK_STAT_TX_HEARTBEAT_ERRORS,
	XETH_LINK_STAT_TX_WINDOW_ERRORS,
	XETH_LINK_STAT_RX_COMPRESSED,
	XETH_LINK_STAT_TX_COMPRESSED,
	XETH_LINK_STAT_RX_NOHANDLER,
	XETH_N_LINK_STAT,
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
	XETH_IFINFO_REASON_FEATURES,
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

static inline bool xeth_is_msg(void *data)
{
	struct xeth_msg *msg = data;

	return	msg->header.z64 == 0 &&
		msg->header.z32 == 0 &&
		msg->header.z16 == 0;
}

static inline enum xeth_msg_kind xeth_msg_kind(void *data)
{
	struct xeth_msg *msg = data;
	return msg->header.kind;
}

static inline bool xeth_msg_version_match(void *data)
{
	struct xeth_msg *msg = data;
	return msg->header.version == XETH_MSG_VERSION;
}

static inline void xeth_msg_init(void *data, enum xeth_msg_kind kind)
{
	struct xeth_msg *msg = data;
	msg->header.z64 = 0;
	msg->header.z32 = 0;
	msg->header.z16 = 0;
	msg->header.version = XETH_MSG_VERSION;
	msg->header.kind = kind;
}

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
	/* @kdata: kind specific data
	 * 	vlan: { XETH_ENCAP_VLAN or XETH_ENCAP_VPLS }
	 * 	loopback: CHANNEL )
	 */
	uint32_t kdata;
	uint8_t ifname[XETH_IFNAMSIZ];
	uint64_t net;
	int32_t ifindex;
	uint32_t flags;
	uint8_t addr[XETH_ALEN];
	uint8_t kind;
	uint8_t reason;
	uint64_t features;
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

struct xeth_msg_netns {
	struct xeth_msg_header header;
	uint64_t net;
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
