// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"
	"net"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
)

// https://en.wikipedia.org/wiki/Internet_Control_Message_Protocol
type ICMP struct {
	Type uint8
	Code uint8
	Sum  big.Uint16
}

func (icmp *ICMP) Format(w fmt.State, verb rune) {
	typename := ICMPTypeName(icmp.Type)
	if len(typename) == 0 {
		typename = "unknown"
	}
	fmt.Fprint(w, "icmp: ", typename, ", ")
	switch icmp.Type {
	case ICMPTypeDestinationUnreachable:
		fmt.Fprint(w, ICMPUnreachableCodeName(icmp.Code))
	case ICMPTypeRedirectMessage:
		fmt.Fprint(w, ICMPRedirectCodeName(icmp.Code))
	case ICMPTypeTimeExceeded:
		fmt.Fprint(w, ICMPTimeExceededCodeName(icmp.Code))
	case ICMPTypeInvalidParamter:
		fmt.Fprint(w, ICMPInvalidParameterCodeName(icmp.Code))
	case ICMPTypeExtendedEchoReply:
		fmt.Fprint(w, ICMPExtendedEchoReplyCodeName(icmp.Code))
	default:
		fmt.Fprint(w, icmp.Code)
	}
}

func (icmp *ICMP) IP() net.IP {
	ip := *(*[net.IPv4len]byte)(Data(icmp))
	return net.IP(ip[:])
}

const (
	ICMPTypeEchoReply = iota
	_
	_
	ICMPTypeDestinationUnreachable
	ICMPTypeSourceQuench
	ICMPTypeRedirectMessage
	_
	_
	ICMPTypeEchoRequest
	ICMPTypeRouterAdvertisement
	ICMPTypeRouterSolicitation
	ICMPTypeTimeExceeded
	ICMPTypeInvalidParamter
	ICMPTypeTimestampRequest
	ICMPTypeTimestampReply
	ICMPTypeInformationRequest
	ICMPTypeInformationReply
	ICMPTypeAddressMaskRequest
	ICMPTypeAddressMaskReeply
)

const (
	ICMPTypeTraceRoute          = 30
	ICMPTypeExtendedEchoRequest = 42
	ICMPTypeExtendedEchoReply   = 42
)

const (
	ICMPUnreachableCodeNetworkUnreachable = iota
	ICMPUnreachableCodeHostUnreachable
	ICMPUnreachableCodeProtocolUnreachable
	ICMPUnreachableCodePortUnreachable
	ICMPUnreachableCodeFragmentationRequired
	ICMPUnreachableCodeSourceRouteFailed
	ICMPUnreachableCodeNetworkUnknown
	ICMPUnreachableCodeHostUnknown
	ICMPUnreachableCodeSourceHostIsolated
	ICMPUnreachableCodeNetworkAdministrativelyProhibited
	ICMPUnreachableCodeHostAdministrativelyProhibited
	ICMPUnreachableCodeTPSNetworkUnreachable
	ICMPUnreachableCodeTPSHostUnreachable
	ICMPUnreachableCodeCommunicationAdministrativelyProhibited
	ICMPUnreachableCodeHostPrecedenceViolation
	ICMPUnreachableCodePrecedenceCutoff
)

const (
	ICMPRedirectCodeNetwork = iota
	ICMPRedirectCodeHost
	ICMPRedirectCodeTOSNetwork
	ICMPRedirectCodeTOSHost
)

const (
	ICMPTimeExceededCodeTTL = iota
	ICMPTimeExceededCodeFragmentReassembly
)

const (
	ICMPInvalidParameterCodePointer = iota
	ICMPInvalidParameterCodeMissingOption
	ICMPInvalidParameterCodeLength
)

const (
	ICMPExtendedEchoReplyCodeNoError = iota
	ICMPExtendedEchoReplyCodeMalformedQuery
	ICMPExtendedEchoReplyCodeNoSuchInterface
	ICMPExtendedEchoReplyCodeNoSuchTableEntry
	ICMPExtendedEchoReplyCodeAmbiguousInterface
)
