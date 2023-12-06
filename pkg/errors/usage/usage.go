// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package usage

import (
	"flag"
	"fmt"
	"io"
	"strings"
	"text/template"
)

// Set Help flag instead of PrintDefaults output during flag.Parse
var Help = flag.Bool("help", false, "Print usage.")

func init() {
	flag.BoolVar(Help, "h", *Help, "aka. -help.")
}

// If args[0] is a string containing "{{", it is parsed as a text/Template that
// is executed with args[1] as data; otherwise, the Sprint of each arg is
// concatenated into a usage error.
func Error(args ...any) error {
	var sb strings.Builder
	if len(args) == 0 {
		return usageError("unspecified usage")
	}
	if s, ok := args[0].(string); ok && strings.Index(s, "{{") >= 0 {
		t, err := template.New("usage").Parse(s)
		if err != nil {
			return usageError(err.Error())
		}
		var data any
		if len(args) > 1 {
			data = args[1]
		}
		if err = t.Execute(&sb, data); err != nil {
			return usageError(err.Error())
		}
	} else {
		fmt.Fprint(&sb, args...)
	}
	return usageError(sb.String())
}

// Returns true if err is a usage error.
func In(err error) bool {
	_, ok := err.(usageError)
	return ok
}

// Make a flag.FlagSet that doesn't output during Parse but instead sets a
// the above Help flag.
func NewFlags(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, 0)
	flags.SetOutput(io.Discard)
	flags.Usage = func() {}
	flags.BoolVar(Help, "help", *Help, "Print usage.")
	flags.BoolVar(Help, "h", *Help, "aka. -help.")
	return flags
}

type usageError string

func (e usageError) Error() string { return string(e) }
