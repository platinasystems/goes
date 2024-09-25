// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

import "encoding/binary"

// See https://en.wikipedia.org/wiki/ICMPv6
type ICMP6 struct {
	Type uint8
	Code uint8
	Sum  uint16
}

const ICMP6TypeIndex = 0
const ICMP6SumIndex = 1 + 1
const ICMP6Size = 1 + 1 + 2

const (
	ICMP6TypeDestinationUnreachable                = 1
	ICMP6TypePacketTooBig                          = 2
	ICMP6TypeTimeExceeded                          = 3
	ICMP6TypeInvalidParameter                      = 4
	ICMP6TypeEchoRequest                           = 128
	ICMP6TypeEchoReply                             = 129
	ICMP6TypeMulticastListenerQuery                = 130
	ICMP6TypeMulticastListenerReport               = 131
	ICMP6TypeMulticastListenerDone                 = 132
	ICMP6TypeRouterSolicitation                    = 133
	ICMP6TypeRouterAdvertisement                   = 134
	ICMP6TypeNeighborSolicitation                  = 135
	ICMP6TypeNeighborAdvertisement                 = 136
	ICMP6TypeRedirectMessage                       = 137
	ICMP6TypeRouterRenumbering                     = 138
	ICMP6TypeNodeInformationQuery                  = 139
	ICMP6TypeNodeInformationResponse               = 140
	ICMP6TypeInverseNeighborDiscoverySolicitation  = 141
	ICMP6TypeInverseNeighborDiscoveryAdvertisement = 142
	ICMP6TypeMulticastListenerDiscovery            = 143
	ICMP6TypeHomeAgentAddressDiscoveryRequest      = 144
	ICMP6TypeHomeAgentAddressDiscoveryReply        = 145
	ICMP6TypeMobilePrefixSolicitation              = 146
	ICMP6TypeMobilePrefixAdvertisement             = 147
	ICMP6TypeCertificationPathSolicitation         = 148
	ICMP6TypeCertificationPathAdvertisement        = 149
	ICMP6TypeMulticastRouterAdvertisement          = 151
	ICMP6TypeMulticastRouterSolicitation           = 152
	ICMP6TypeMulticastRouterTermination            = 153
	ICMP6TypeRPLControlMessage                     = 155
)

const (
	ICMP6UnreachableCodeNoRouteToDestination = iota
	ICMP6UnreachableCodeCommunicationAdministrativelyProhibited
	ICMP6UnreachableCodeBeyondScopeOfSourceAddress
	ICMP6UnreachableCodeAddressUnreachable
	ICMP6UnreachableCodePortUnreachable
	ICMP6UnreachableCodeSourceAddressFailedIngressOrEgressPolicy
	ICMP6UnreachableCodeRejectRouteToDestination
	ICMP6UnreachableCodeErrorInSourceRoutingHeader
)

const (
	ICMP6TimeExceededCodeTransmitHopLimit = iota
	ICMP6TimeExceededCodeFragmentReassembly
)

const (
	ICMP6InvalidParameterCodeErroneousHeader = iota
	ICMP6InvalidParameterCodeUnrecognizedNextHeader
	ICMP6InvalidParameterCodeUnrecognizedOption
)

const (
	ICMP6RouterRenumberingCodeCommand = iota
	ICMP6RouterRenumberingCodeResult
)

// See https://www.rfc-editor.org/rfc/rfc4443.html#page-13
type ICMP6EchoRequest struct {
	ICMP6
	Identifier,
	Sequence uint16
}

type ICMP6EchoReply struct {
	ICMP6
	Identifier,
	Sequence uint16
}

// See https://datatracker.ietf.org/doc/html/rfc4861#section-4.1
type ICMP6RouterSolicitation struct {
	ICMP6
	_ uint32
}

// See https://datatracker.ietf.org/doc/html/rfc4861#section-4.2
type ICMP6RouterAdvertisement struct {
	ICMP6
	CurHopLimit,
	Flags uint8
	RouterLifetime uint16
	ReachableTime,
	RetransTimer uint32
}

const (
	_ = 1 << iota
	_ // 0x02
	_ // 0x04
	_ // 0x08
	_ // 0x10
	_ // 0x20
	ICMP6RouterAdvertisementOtherConfiguration
	ICMP6RouterAdvertisementManagedAddress
)

// See https://datatracker.ietf.org/doc/html/rfc4861#section-4.3
type ICMP6NeighborSolicitation struct {
	ICMP6
	_ uint32

	TargetAddress [IPv6len]byte
}

// https://datatracker.ietf.org/doc/html/rfc4861#section-4.4
type ICMP6NeighborAdvertisement struct {
	ICMP6
	Flags,
	_ uint8
	_ uint16

	TargetAddress [IPv6len]byte
}

const (
	_ = 1 << iota
	_ // 0x02
	_ // 0x04
	_ // 0x08
	_ // 0x10
	ICMP6NeighborAdvertisementOverride
	ICMP6NeighborAdvertisementSolicited
	ICMP6NeighborAdvertisementRouter
)

// https://datatracker.ietf.org/doc/html/rfc4861#section-4.5
type ICMP6RedirectMessage struct {
	ICMP6
	_ uint32
	TargetAddress,
	DestinationAddress [IPv6len]byte
}

// https://datatracker.ietf.org/doc/html/rfc4861#section-4.6
type ICMP6Option struct {
	Type,
	Length uint8
}

const ICMP6OptionSize = 2
const ICMP6OptionLengthIndex = 1

const (
	ICMP6OptionTypeSourceLinkLayerAddress = 1 + iota
	ICMP6OptionTypeTargetLinkLayerAddress
	ICMP6OptionTypePrefixInformation
	ICMP6OptionTypeRedirectedHeader
	ICMP6OptionTypeMTU
)

const (
	ICMP6OptionTypeRDNSS = 25
	ICMP6OptionTypeDNSSL = 31
)

type ICMP6OptionHeader interface {
	ICMP6SourceLinkLayerAddress |
		ICMP6TargetLinkLayerAddress |
		ICMP6PrefixInformation |
		ICMP6RedirectedHeader |
		ICMP6MTU |
		ICMP6RDNSS |
		ICMP6DNSSL
}

func ParseICMP6Option[H ICMP6OptionHeader](b []byte) (
	header H, data, next []byte, err error,
) {
	n, err := binary.Decode(b, binary.BigEndian, &header)
	if err == nil {
		i := int(b[1])
		data = b[n:i]
		next = b[i:]
	}
	return
}

// https://datatracker.ietf.org/doc/html/rfc4861#section-4.6.1
// Note: “Link-Layer Address” is the variable length data that follows this
// empty header.
type ICMP6SourceLinkLayerAddress struct{ ICMP6Option }
type ICMP6TargetLinkLayerAddress struct{ ICMP6Option }

const ICMP6SourceLinkLayerAddressSize = ICMP6OptionSize
const ICMP6TargetLinkLayerAddressSize = ICMP6OptionSize

// https://datatracker.ietf.org/doc/html/rfc4861#section-4.6.2
type ICMP6PrefixInformation struct {
	ICMP6Option
	PrefixLength,
	Flags uint8
	ValidLifetime,
	PreferredLifetime,
	_ uint32
	Prefix [IPv6len]byte
}

const (
	_ = 1 << iota
	_ // 0x02
	_ // 0x04
	_ // 0x08
	_ // 0x10
	_ // 0x20
	ICMP6PrefixAutonomousAddressConfiguration
	ICMP6PrefixOnLink
)

// https://datatracker.ietf.org/doc/html/rfc4861#section-4.6.3
// Note: “IP header + data” is the variable length data that follows these
// reserved fields.
type ICMP6RedirectedHeader struct {
	ICMP6Option
	_ uint16
	_ uint32
}

// https://datatracker.ietf.org/doc/html/rfc4861#section-4.6.4
type ICMP6MTU struct {
	ICMP6Option
	_   uint16
	MTU uint32
}

// https://datatracker.ietf.org/doc/html/rfc8106#section-5.1
// Note: “Addresses of IPv6 Recursive DNS Servers” are the 16-byte
// arrays that follow this ICMP6RDNSS header.
type ICMP6RDNSS struct {
	ICMP6Option
	_        uint16
	Lifetime uint32
}

// https://datatracker.ietf.org/doc/html/rfc8106#section-5.2
// Note: “Domain Names of DNS Search List” are the null terminated strings
// that follow this ICMP6RDNSSSize header.
type ICMP6DNSSL struct {
	ICMP6Option
	_        uint16
	Lifetime uint32
}
