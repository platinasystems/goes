// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
)

// https://en.wikipedia.org/wiki/ICMPv6
type ICMP6 struct {
	Type ICMP6Type
	Code uint8
	Sum  big.Uint16
}

func (icmp6 *ICMP6) Format(w fmt.State, verb rune) {
	var code any
	fmt.Fprint(w, "icmp6: ")
	switch icmp6.Type {
	case ICMP6TypeDestinationUnreachable:
		code = ICMP6UnreachableCode(icmp6.Code)
	case ICMP6TypeTimeExceeded:
		code = ICMP6TimeExceededCode(icmp6.Code)
	case ICMP6TypeInvalidParameter:
		code = ICMP6InvalidParameterCode(icmp6.Code)
	case ICMP6TypeRouterRenumbering:
		code = ICMP6RouterRenumberingCode(icmp6.Code)
	default:
		code = icmp6.Code
	}
	fmt.Fprint(w, icmp6.Type, ", ", code)
}

type ICMP6Type uint8

//go:generate stringer -output=zicmp6_type_string.go -type=ICMP6Type -trimprefix=ICMP6Type .
const (
	ICMP6TypeDestinationUnreachable                ICMP6Type = 1
	ICMP6TypePacketTooBig                          ICMP6Type = 2
	ICMP6TypeTimeExceeded                          ICMP6Type = 3
	ICMP6TypeInvalidParameter                      ICMP6Type = 4
	ICMP6TypeEchoRequest                           ICMP6Type = 128
	ICMP6TypeEchoReply                             ICMP6Type = 129
	ICMP6TypeMulticastListenerQuery                ICMP6Type = 130
	ICMP6TypeMulticastListenerReport               ICMP6Type = 131
	ICMP6TypeMulticastListenerDone                 ICMP6Type = 132
	ICMP6TypeRouterSolicitation                    ICMP6Type = 133
	ICMP6TypeRouterAdvertisement                   ICMP6Type = 134
	ICMP6TypeNeighborSolicitation                  ICMP6Type = 135
	ICMP6TypeNeighborAdvertisement                 ICMP6Type = 136
	ICMP6TypeRedirectMessage                       ICMP6Type = 137
	ICMP6TypeRouterRenumbering                     ICMP6Type = 138
	ICMP6TypeNodeInformationQuery                  ICMP6Type = 139
	ICMP6TypeNodeInformationResponse               ICMP6Type = 140
	ICMP6TypeInverseNeighborDiscoverySolicitation  ICMP6Type = 141
	ICMP6TypeInverseNeighborDiscoveryAdvertisement ICMP6Type = 142
	ICMP6TypeMulticastListenerDiscovery            ICMP6Type = 143
	ICMP6TypeHomeAgentAddressDiscoveryRequest      ICMP6Type = 144
	ICMP6TypeHomeAgentAddressDiscoveryReply        ICMP6Type = 145
	ICMP6TypeMobilePrefixSolicitation              ICMP6Type = 146
	ICMP6TypeMobilePrefixAdvertisement             ICMP6Type = 147
	ICMP6TypeCertificationPathSolicitation         ICMP6Type = 148
	ICMP6TypeCertificationPathAdvertisement        ICMP6Type = 149
	ICMP6TypeMulticastRouterAdvertisement          ICMP6Type = 151
	ICMP6TypeMulticastRouterSolicitation           ICMP6Type = 152
	ICMP6TypeMulticastRouterTermination            ICMP6Type = 153
	ICMP6TypeRPLControlMessage                     ICMP6Type = 155
)

type ICMP6UnreachableCode uint8

//go:generate stringer -output=zicmp6_unreachable_string.go -type=ICMP6UnreachableCode -trimprefix=ICMP6UnreachableCode .
const (
	ICMP6UnreachableCodeNoRouteToDestination ICMP6UnreachableCode = iota
	ICMP6UnreachableCodeCommunicationAdministrativelyProhibited
	ICMP6UnreachableCodeBeyondScopeOfSourceAddress
	ICMP6UnreachableCodeAddressUnreachable
	ICMP6UnreachableCodePortUnreachable
	ICMP6UnreachableCodeSourceAddressFailedIngressOrEgressPolicy
	ICMP6UnreachableCodeRejectRouteToDestination
	ICMP6UnreachableCodeErrorInSourceRoutingHeader
)

type ICMP6TimeExceededCode uint8

//go:generate stringer -output=zicmp6_time_exceeded_string.go -type=ICMP6TimeExceededCode -trimprefix=ICMP6TimeExceededCode .
const (
	ICMP6TimeExceededCodeTransmitHopLimit ICMP6TimeExceededCode = iota
	ICMP6TimeExceededCodeFragmentReassembly
)

type ICMP6InvalidParameterCode uint8

//go:generate stringer -output=zicmp6_invalid_parameter_string.go -type=ICMP6InvalidParameterCode -trimprefix=ICMP6InvalidParameterCode .
const (
	ICMP6InvalidParameterCodeErroneousHeader ICMP6InvalidParameterCode = iota
	ICMP6InvalidParameterCodeUnrecognizedNextHeader
	ICMP6InvalidParameterCodeUnrecognizedOption
)

type ICMP6RouterRenumberingCode uint8

//go:generate stringer -output=zicmp6_router_renumbering_string.go -type=ICMP6RouterRenumberingCode -trimprefix=ICMP6RouterRenumberingCode .
const (
	ICMP6RouterRenumberingCodeCommand ICMP6RouterRenumberingCode = iota
	ICMP6RouterRenumberingCodeResult
)
