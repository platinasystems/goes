// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

// https://en.wikipedia.org/wiki/Internet_Control_Message_Protocol
type ICMP struct {
	Type uint8
	Code uint8
	Sum  uint16
}

const ICMPSumIndex = 1 + 1
const ICMPSize = 1 + 1 + 2

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
