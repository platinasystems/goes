// Copyright © 2015-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netlink

const (
	NLM_F_REQUEST = 1 << iota
	NLM_F_MULTI
	NLM_F_ACK
	NLM_F_ECHO
	NLM_F_DUMP_INTR
	NLM_F_DUMP_FILTERED
)

// Modifiers to GET request
const (
	NLM_F_ROOT = 0x100 << iota
	NLM_F_MATCH
	NLM_F_ATOMIC
)

const NLM_F_DUMP = NLM_F_ROOT | NLM_F_MATCH

// Modifiers to NEW request
const (
	NLM_F_REPLACE = 0x100 << iota
	NLM_F_EXCL
	NLM_F_CREATE
	NLM_F_APPEND
)

// Modifiers to DELETE request
const (
	NLM_F_NONREC = 0x10 << iota
	NLM_F_BULK
)

// Flags for ACK message
const (
	NLM_F_CAPPED = 0x100 << iota
	NLM_F_ACK_TLVS
)
