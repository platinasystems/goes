// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package cmd

import (
	"strings"

	"github.com/platinasystems/go/goes/lang"
)

var Helpers = map[string]struct{}{
	"apropos":  struct{}{},
	"complete": struct{}{},
	"help":     struct{}{},
	"man":      struct{}{},
	"usage":    struct{}{},
}

// Swap hyphen prefaced helper flags with command, so,
//
//	COMMAND -[-]HELPER [ARGS]...
//
// becomes
//
//	HELPER COMMAND [ARGS]...
//
// and
//
//	-[-]HELPER [ARGS]...
//
// becomes
//
//	HELPER [ARGS]...
func Swap(args []string) {
	n := len(args)
	if n > 0 && strings.HasPrefix(args[0], "-") {
		opt := strings.TrimLeft(args[0], "-")
		if _, found := Helpers[opt]; found {
			args[0] = opt
		}
	} else if n > 1 && strings.HasPrefix(args[1], "-") {
		opt := strings.TrimLeft(args[1], "-")
		if _, found := Helpers[opt]; found {
			args[1] = args[0]
			args[0] = opt
		}
	}
}

type Cmd interface {
	Apropos() lang.Alt
	Main(...string) error
	// String returns the coammand name.
	String() string
	Usage() string
	/* Optional
	Close() error
	Complete(...string) []string
	Goese(*goes.Goes)
	Help(...string) string
	Kind() Kind
	Man() lang.Alt
	*/
}
