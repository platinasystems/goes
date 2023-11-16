// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !darwin && !freebsd && !netbsd

package netioctl

const (
	SIOCGIFCAP = 0
	SIOCSIFCAP = 0
)

type IFCAP struct{}

type IFCAPS struct{ Req, Cur IFCAP }
