// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	_ "embed"
	"strings"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/integer"
)

//go:embed icmp_type.txt
var ICMPTypeText string

var ICMPTypeLines = sync.OnceValue(func() []string {
	return strings.Split(ICMPTypeText, "\n")
})

func ICMPTypeName[T integer.Int | integer.Uint](i T) string {
	return integer.Name(i, ICMPTypeLines())
}

//go:embed icmp_unreachable_code.txt
var ICMPUnreachableCodeText string

var ICMPUnreachableCodeLines = sync.OnceValue(func() []string {
	return strings.Split(ICMPUnreachableCodeText, "\n")
})

func ICMPUnreachableCodeName[T integer.Int | integer.Uint](i T) string {
	return integer.Name(i, ICMPUnreachableCodeLines())
}

//go:embed icmp_redirect_code.txt
var ICMPRedirectCodeText string

var ICMPRedirectCodeLines = sync.OnceValue(func() []string {
	return strings.Split(ICMPRedirectCodeText, "\n")
})

func ICMPRedirectCodeName[T integer.Int | integer.Uint](i T) string {
	return integer.Name(i, ICMPRedirectCodeLines())
}

//go:embed icmp_time_exceeded_code.txt
var ICMPTimeExceededCodeText string

var ICMPTimeExceededCodeLines = sync.OnceValue(func() []string {
	return strings.Split(ICMPTimeExceededCodeText, "\n")
})

func ICMPTimeExceededCodeName[T integer.Int | integer.Uint](i T) string {
	return integer.Name(i, ICMPTimeExceededCodeLines())
}

//go:embed icmp_invalid_parameter_code.txt
var ICMPInvalidParameterCodeText string

var ICMPInvalidParameterCodeLines = sync.OnceValue(func() []string {
	return strings.Split(ICMPInvalidParameterCodeText, "\n")
})

func ICMPInvalidParameterCodeName[T integer.Int | integer.Uint](i T) string {
	return integer.Name(i, ICMPInvalidParameterCodeLines())
}

//go:embed icmp_extended_echo_reply_code.txt
var ICMPExtendedEchoReplyCodeText string

var ICMPExtendedEchoReplyCodeLines = sync.OnceValue(func() []string {
	return strings.Split(ICMPExtendedEchoReplyCodeText, "\n")
})

func ICMPExtendedEchoReplyCodeName[T integer.Int | integer.Uint](i T) string {
	return integer.Name(i, ICMPExtendedEchoReplyCodeLines())
}

//go:embed icmp6_type.txt
var ICMP6TypeText string

var ICMP6TypeLines = sync.OnceValue(func() []string {
	return strings.Split(ICMP6TypeText, "\n")
})

func ICMP6TypeName[T integer.Int | integer.Uint](i T) string {
	return integer.Name(i, ICMP6TypeLines())
}

//go:embed icmp6_unreachable_code.txt
var ICMP6UnreachableCodeText string

var ICMP6UnreachableCodeLines = sync.OnceValue(func() []string {
	return strings.Split(ICMP6UnreachableCodeText, "\n")
})

func ICMP6UnreachableCodeName[T integer.Int | integer.Uint](i T) string {
	return integer.Name(i, ICMP6UnreachableCodeLines())
}

//go:embed icmp6_time_exceeded_code.txt
var ICMP6TimeExceededCodeText string

var ICMP6TimeExceededCodeLines = sync.OnceValue(func() []string {
	return strings.Split(ICMP6TimeExceededCodeText, "\n")
})

func ICMP6TimeExceededCodeName[T integer.Int | integer.Uint](i T) string {
	return integer.Name(i, ICMP6TimeExceededCodeLines())
}

//go:embed icmp6_invalid_parameter_code.txt
var ICMP6InvalidParameterCodeText string

var ICMP6InvalidParameterCodeLines = sync.OnceValue(func() []string {
	return strings.Split(ICMP6InvalidParameterCodeText, "\n")
})

func ICMP6InvalidParameterCodeName[T integer.Int | integer.Uint](i T) string {
	return integer.Name(i, ICMP6InvalidParameterCodeLines())
}

//go:embed icmp6_router_renumbering_code.txt
var ICMP6RouterRenumberingCodeText string

var ICMP6RouterRenumberingCodeLines = sync.OnceValue(func() []string {
	return strings.Split(ICMP6RouterRenumberingCodeText, "\n")
})

func ICMP6RouterRenumberingCodeName[T integer.Int | integer.Uint](i T) string {
	return integer.Name(i, ICMP6RouterRenumberingCodeLines())
}
