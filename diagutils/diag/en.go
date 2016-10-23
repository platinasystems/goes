// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// +build !noten

package diag

func (diag *diag) Apropos() string {
	return "run diagnostics"
}

func (diag *diag) Man() string {
	return `NAME
	diag - run diagnostics

SYNOPSIS
	diag

DESCRIPTION
	Runs diagnostic tests to validate BMC functionality and interfaces

EXAMPLES
	diag
`
}
