// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

//go:generate stringer -type OpCode -trimprefix OpCode
type OpCode uint16

const (
	OpCodeQuery OpCode = iota
	OpCodeIQuery
	OpCodeStatus
)
