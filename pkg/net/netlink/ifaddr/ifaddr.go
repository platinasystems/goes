/* SPDX-License-Identifier: GPL-2.0 WITH Linux-syscall-note */

package ifaddr

type Msg struct {
	Family    uint8
	PrefixLen uint8
	Flags     uint8
	Scope     uint8
	Index     uint32
}

type Ifa uint16

const (
	IFA_UNSPEC Ifa = iota
	IFA_ADDRESS
	IFA_LOCAL
	IFA_LABEL
	IFA_BROADCAST
	IFA_ANYCAST
	IFA_CACHEINFO
	IFA_MULTICAST
	IFA_FLAGS
	IFA_RT_PRIORITY
	IFA_TARGET_NETNSID
	IFA_PROTO
	IFA_CNT
)
const IFA_MAX Ifa = IFA_CNT - 1

type Attr struct {
	Len  uint16
	Type Ifa
}

const (
	IFA_F_SECONDARY = 1 << iota
	IFA_F_NODAD
	IFA_F_OPTIMISTIC
	IFA_F_DADFAILED
	IFA_F_HOMEADDRESS
	IFA_F_DEPRECATED
	IFA_F_TENTATIVE
	IFA_F_PERMANENT
	IFA_F_MANAGETEMPADDR
	IFA_F_NOPREFIXROUTE
	IFA_F_MCAUTOJOIN
	IFA_F_STABLE_PRIVACY
)
const IFA_F_TEMPORARY = IFA_F_SECONDARY

type Cacheinfo struct {
	Prefered uint32
	Valid    uint32
	Cstamp   uint32
	Tstamp   uint32
}

/* Replaced by netlink.Extract[ifaddr.IfAddrMsg]
#define IFA_RTA(r)  ((struct rtattr*)(((char*)(r)) + NLMSG_ALIGN(sizeof(struct ifaddrmsg))))
#define IFA_PAYLOAD(n) NLMSG_PAYLOAD(n,sizeof(struct ifaddrmsg))
*/

const (
	IFAPROT_UNSPEC = iota
	IFAPROT_KERNEL_LO
	IFAPROT_KERNEL_RA
	IFAPROT_KERNEL_LL
)
