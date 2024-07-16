// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !darwin && !freebsd && !netbsd

package xnet

const (
	SIOCGIFCAP = 0
	SIOCSIFCAP = 0
)

func IFCAPStrings() []string {
	return []string{}
}
