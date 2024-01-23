package frame

import "golang.org/x/sys/unix"

const (
	IPPROTO_ICMP   = unix.IPPROTO_ICMP
	IPPROTO_ICMPV6 = unix.IPPROTO_ICMPV6
	IPPROTO_TCP    = unix.IPPROTO_TCP
	IPPROTO_UDP    = unix.IPPROTO_UDP
)
