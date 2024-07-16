// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build darwin || freebsd || netbsd

package xnet

import (
	_ "embed"
	"strings"

	"golang.org/x/sys/unix"
)

const (
	SIOCGIFCAP = unix.SIOCGIFCAP
	SIOCSIFCAP = unix.SIOCSIFCAP
)

//go:embed ifcap.txt
var IFCAPText string

func IFCAPStrings() []string {
	return strings.Split(IFCAPText, "\n")
}
