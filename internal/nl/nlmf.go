// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nl

import "syscall"

const (
	NLM_F_ACK     uint16 = syscall.NLM_F_ACK
	NLM_F_APPEND  uint16 = syscall.NLM_F_APPEND
	NLM_F_ATOMIC  uint16 = syscall.NLM_F_ATOMIC
	NLM_F_CREATE  uint16 = syscall.NLM_F_CREATE
	NLM_F_DUMP    uint16 = syscall.NLM_F_DUMP
	NLM_F_ECHO    uint16 = syscall.NLM_F_ECHO
	NLM_F_EXCL    uint16 = syscall.NLM_F_EXCL
	NLM_F_MATCH   uint16 = syscall.NLM_F_MATCH
	NLM_F_MULTI   uint16 = syscall.NLM_F_MULTI
	NLM_F_REPLACE uint16 = syscall.NLM_F_REPLACE
	NLM_F_REQUEST uint16 = syscall.NLM_F_REQUEST
	NLM_F_ROOT    uint16 = syscall.NLM_F_ROOT
)
