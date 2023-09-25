// Copyright © 2023 Platina Systems, Inc. All rights reserved.
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
	Type ICMPType
	Code uint8
	Sum  big.Uint16
}

func (icmp *ICMP) Format(w fmt.State, verb rune) {
	var code any
	fmt.Fprint(w, "icmp: ")
	switch icmp.Type {
	case ICMPTypeDestinationUnreachable:
		code = ICMPUnreachableCode(icmp.Code)
	case ICMPTypeRedirectMessage:
		code = ICMPRedirectCode(icmp.Code)
	case ICMPTypeTimeExceeded:
		code = ICMPTimeExceededCode(icmp.Code)
	case ICMPTypeInvalidParamter:
		code = ICMPInvalidParameterCode(icmp.Code)
	case ICMPTypeExtendedEchoReply:
		code = ICMPExtendedEchoReplyCode(icmp.Code)
	default:
		code = icmp.Code
	}
	fmt.Fprint(w, icmp.Type, ", ", code)
}

func (icmp *ICMP) IP() net.IP {
	ip := *(*[net.IPv4len]byte)(Data(icmp))
	return net.IP(ip[:])
}

type ICMPType uint8

//go:generate stringer -output=zicmp_type_string.go -type=ICMPType -trimprefix=ICMPType .
const (
	ICMPTypeEchoReply ICMPType = iota
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

const ICMPTypeTraceRoute = ICMPType(30)
const ICMPTypeExtendedEchoRequest = ICMPType(42)
const ICMPTypeExtendedEchoReply = ICMPType(42)

type ICMPUnreachableCode uint8

//go:generate stringer -output=zicmp_unreachable_string.go -type=ICMPUnreachableCode -trimprefix=ICMPUnreachableCode .
const (
	ICMPUnreachableCodeNetworkUnreachable ICMPUnreachableCode = iota
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

type ICMPRedirectCode uint8

//go:generate stringer -output=zicmp_redirect_string.go -type=ICMPRedirectCode -trimprefix=ICMPRedirectCode .
const (
	ICMPRedirectCodeNetwork ICMPRedirectCode = iota
	ICMPRedirectCodeHost
	ICMPRedirectCodeTOSNetwork
	ICMPRedirectCodeTOSHost
)

type ICMPTimeExceededCode uint8

//go:generate stringer -output=zicmp_time_exceeded_string.go -type=ICMPTimeExceededCode -trimprefix=ICMPTimeExceededCode .
const (
	ICMPTimeExceededCodeTTL ICMPTimeExceededCode = iota
	ICMPTimeExceededCodeFragmentReassembly
)

type ICMPInvalidParameterCode uint8

//go:generate stringer -output=zicmp_invalid_parameter_string.go -type=ICMPInvalidParameterCode -trimprefix=ICMPInvalidParameterCode .
const (
	ICMPInvalidParameterCodePointer ICMPInvalidParameterCode = iota
	ICMPInvalidParameterCodeMissingOption
	ICMPInvalidParameterCodeLength
)

type ICMPExtendedEchoReplyCode uint8

//go:generate stringer -output=zicmp_extended_echo_reply_string.go -type=ICMPExtendedEchoReplyCode -trimprefix=ICMPExtendedEchoReplyCode .
const (
	ICMPExtendedEchoReplyCodeNoError ICMPExtendedEchoReplyCode = iota
	ICMPExtendedEchoReplyCodeMalformedQuery
	ICMPExtendedEchoReplyCodeNoSuchInterface
	ICMPExtendedEchoReplyCodeNoSuchTableEntry
	ICMPExtendedEchoReplyCodeAmbiguousInterface
)
