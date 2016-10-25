// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// +build !noten

package gpio

func (*gpio) Apropos() string {
	return "manipulate GPIO pins"
}

func (*gpio) Man() string {
	return `NAME
	gpio - Manipulate GPIO pins

SYNOPSIS
	gpio

DESCRIPTION
	Manipulate GPIO pins`
}
