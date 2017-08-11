// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import (
	"fmt"
	"syscall"
)

const (
	RT_TABLE_UNSPEC  uint32 = syscall.RT_TABLE_UNSPEC
	RT_TABLE_COMPAT  uint32 = syscall.RT_TABLE_COMPAT
	RT_TABLE_DEFAULT uint32 = syscall.RT_TABLE_DEFAULT
	RT_TABLE_MAIN    uint32 = syscall.RT_TABLE_MAIN
	RT_TABLE_LOCAL   uint32 = syscall.RT_TABLE_LOCAL
	RT_TABLE_MAX     uint32 = syscall.RT_TABLE_MAX
)

var RtTableByName = map[string]uint32{
	"unspec":  RT_TABLE_UNSPEC,
	"compat":  RT_TABLE_COMPAT,
	"default": RT_TABLE_DEFAULT,
	"main":    RT_TABLE_MAIN,
	"local":   RT_TABLE_LOCAL,
	"max":     RT_TABLE_MAX,
}

func RtTableName(t uint32) string {
	var s string
	// very sparse so use cascaded conditions instead of map
	switch {
	case t == RT_TABLE_UNSPEC:
		s = "unspec"
	case t == RT_TABLE_COMPAT:
		s = "compat"
	case t == RT_TABLE_DEFAULT:
		s = "default"
	case t == RT_TABLE_MAIN:
		s = "main"
	case t == RT_TABLE_LOCAL:
		s = "local"
	default:
		s = fmt.Sprint(t)
	}
	return s
}
