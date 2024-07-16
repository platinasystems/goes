// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package iflink

import (
	_ "embed"
	"strings"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/integer"
)

//go:embed iff.txt
var IffText string
var IffLines = sync.OnceValue(func() []string {
	return strings.Split(IffText, "\n")
})

func IffNames[T integer.Int | integer.Uint](i T) string {
	return integer.Names(i, IffLines())
}

//go:embed iface.txt
var IfaceText string
var IfaceLines = sync.OnceValue(func() []string {
	return strings.Split(IfaceText, "\n")
})

func IfaceName[T integer.Int | integer.Uint](i T) string {
	integer.Sub(&i, IfaceBase)
	return integer.Name(i, IfaceLines())
}

//go:embed ifproto.txt
var IfProtoText string
var IfProtoLines = sync.OnceValue(func() []string {
	return strings.Split(IfProtoText, "\n")
})

func IfProtoName[T integer.Int | integer.Uint](i T) string {
	integer.Sub(&i, IfProtoBase)
	return integer.Name(i, IfProtoLines())
}

//go:embed ifoper.txt
var IfOperText string
var IfOperLines = sync.OnceValue(func() []string {
	return strings.Split(IfOperText, "\n")
})

func IfOperName[T integer.Int | integer.Uint](i T) string {
	return integer.Name(i, IfOperLines())
}

const INVALID_IF_OPER = -1

func IfOperByName(name string) int8 {
	for i, line := range IfOperLines() {
		if name == line {
			return int8(i)
		}
	}
	return INVALID_IF_OPER
}

//go:embed iflinkmode.txt
var IfLinkModeText string
var IfLinkModeLines = sync.OnceValue(func() []string {
	return strings.Split(IfLinkModeText, "\n")
})

func IfLinkModeName[T integer.Int | integer.Uint](i T) string {
	return integer.Name(i, IfLinkModeLines())
}

const INVALID_IF_LINK_MODE = -1

func IfLinkModeByName(name string) int8 {
	for i, line := range IfLinkModeLines() {
		if name == line {
			return int8(i)
		}
	}
	return INVALID_IF_LINK_MODE
}
