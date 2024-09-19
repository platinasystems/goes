// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

import "unsafe"

// See https://en.wikipedia.org/wiki/ICMPv6
type ICMP6 struct {
	Type uint8
	Code uint8
	Sum  uint16
}

const ICMP6SumIndex = 2
const ICMP6Size = int(unsafe.Sizeof(ICMP6{}))

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
	Identifier,
	Sequence uint16
}

const ICMP6EchoRequestSize = int(unsafe.Sizeof(ICMP6EchoRequest{}))

type ICMP6EchoReply = ICMP6EchoRequest

const ICMP6EchoReplySize = ICMP6EchoRequestSize

// See https://datatracker.ietf.org/doc/html/rfc4861#section-4.1
type ICMP6RouterSolicitation struct {
	_ uint32
}

const ICMP6RouterSolicitationSize = int(unsafe.
	Sizeof(ICMP6RouterSolicitation{}))

// See https://datatracker.ietf.org/doc/html/rfc4861#section-4.2
type ICMP6RouterAdvertisement struct {
	CurHopLimit,
	Flags uint8
	Lifetime uint16
	ReachableTime,
	RetransTimer uint32
}

const ICMP6RouterAdvertisementSize = int(unsafe.
	Sizeof(ICMP6RouterAdvertisement{}))

const (
	_ = 1 << iota
	_ // 0x02
	_ // 0x04
	_ // 0x06
	_ // 0x08
	_ // 0x10
	_ // 0x20
	_ // 0x40
	ICMP6RouterAdvertisementOtherConfiguration
	ICMP6RouterAdvertisementManagedAddress
)

// See https://datatracker.ietf.org/doc/html/rfc4861#section-4.3
type ICMP6NeighborSolicitation struct {
	_ uint32

	TargetAddress [IPv6len]byte
}

const ICMP6NeighborSolicitationSize = int(unsafe.
	Sizeof(ICMP6NeighborSolicitation{}))

// https://datatracker.ietf.org/doc/html/rfc4861#section-4.4
type ICMP6NeighborAdvertisement struct {
	Flags,
	_ uint8

	_ uint16

	TargetAddress [IPv6len]byte
}

const ICMP6NeighborAdvertisementSize = int(unsafe.
	Sizeof(ICMP6NeighborAdvertisement{}))

const (
	_ = 1 << iota
	_ // 0x02
	_ // 0x04
	_ // 0x06
	_ // 0x08
	_ // 0x10
	_ // 0x20
	ICMP6NeighborAdvertisementOverride
	ICMP6NeighborAdvertisementSolicited
	ICMP6NeighborAdvertisementRouter
)

// https://datatracker.ietf.org/doc/html/rfc4861#section-4.5
type ICMP6RedirectMessage struct {
	_ uint32
	TargetAddress,
	DestinationAddress [IPv6len]byte
}

const ICMP6RedirectMessageSize = int(unsafe.Sizeof(ICMP6RedirectMessage{}))

// https://datatracker.ietf.org/doc/html/rfc4861#section-4.6
type ICMP6Option struct {
	Type,
	Length uint8
}

const ICMP6OptionSize = int(unsafe.Sizeof(ICMP6Option{}))

const (
	ICMP6OptionTypeSourceLinkLayerAddress = 1 + iota
	ICMP6OptionTypeTargetLinkLayerAddress
	ICMP6OptionTypePrefixInformation
	ICMP6OptionTypeRedirectedHeader
	ICMP6OptionTypeMTU
)

// https://datatracker.ietf.org/doc/html/rfc4861#section-4.6.2
type ICMP6Prefix struct {
	PrefixLength,
	Flags uint8
	ValidLifetime,
	PreferredLifetime,
	_ uint32
	Prefix [IPv6len]byte
}

const ICMP6PrefixSize = int(unsafe.Sizeof(ICMP6Prefix{}))

// https://datatracker.ietf.org/doc/html/rfc4861#section-4.6.4
type ICMP6MTU struct {
	_   uint16
	MTU uint32
}

const ICMP6MTUSize = int(unsafe.Sizeof(ICMP6MTU{}))
