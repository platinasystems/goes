// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
)

// https://en.wikipedia.org/wiki/ICMPv6
type ICMP6 struct {
	Type uint8
	Code uint8
	Sum  big.Uint16
}

func (icmp6 *ICMP6) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "icmp6: ", ICMP6TypeName(icmp6.Type), ", ")
	switch icmp6.Type {
	case ICMP6TypeDestinationUnreachable:
		fmt.Fprint(w, ICMP6UnreachableCodeName(icmp6.Code))
	case ICMP6TypeTimeExceeded:
		fmt.Fprint(w, ICMP6TimeExceededCodeName(icmp6.Code))
	case ICMP6TypeInvalidParameter:
		fmt.Fprint(w, ICMP6InvalidParameterCodeName(icmp6.Code))
	case ICMP6TypeRouterRenumbering:
		fmt.Fprint(w, ICMP6RouterRenumberingCodeName(icmp6.Code))
	default:
		fmt.Fprint(w, icmp6.Code)
	}
}

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
