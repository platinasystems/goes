// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// +build !noten

package reboot

func (*reboot) Apropos() string {
	return "reboot system"
}

func (*reboot) Man() string {
	return `NAME
	reboot - reboots system

SYNOPSIS
	reboot

DESCRIPTION
	Reboots system.`
}
