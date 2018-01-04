// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package cmd

import "github.com/platinasystems/go/goes/lang"

type Cmd interface {
	Apropos() lang.Alt
	Main(...string) error
	// String returns the coammand name.
	String() string
	Usage() string
	/* Optional
	Aka() string
	Close() error
	Complete(...string) []string
	Goes(*goes.Goes)
	Help(...string) string
	Kind() Kind
	Man() lang.Alt
	*/
}
