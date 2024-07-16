// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xflag

import (
	"flag"
	"fmt"
	"strings"
	"text/template"
)

var UsageFuncs = template.FuncMap{
	"flags": SprintDefaults,
}

// These results are passed to the usage template execution.
var UsageData = func(flags *flag.FlagSet) any {
	return flags
}

// [flag.FlagSet.Init] wht new name but current [flag.ErrorHandling].
func Rename(flags *flag.FlagSet, name string) {
	flags.Init(name, flags.ErrorHandling())
}

// Redirect [flag.FlagSet.PrintDefaults] to string.
func SprintDefaults(flags *flag.FlagSet) string {
	save := flags.Output()
	defer flags.SetOutput(save)
	var sb strings.Builder
	flags.SetOutput(&sb)
	flags.PrintDefaults()
	return sb.String()
}

// Assign [flag.FlagSet.Usage] to a closure that creates a new [text/template]
// with [UsageFuncs]; parses “tmpl”; then [text/template.Template.Execute]'s
// to [flag.FlagSet.Output] with [UsageData] results.
func UsageTemplate(flags *flag.FlagSet, tmpl string) {
	flags.Usage = func() {
		w := flags.Output()
		tmpl = strings.TrimLeft(tmpl, " \t\n")
		tt, err := template.New("usage").Funcs(UsageFuncs).Parse(tmpl)
		if err == nil {
			err = tt.Execute(w, UsageData(flags))
		}
		if err != nil {
			name := strings.Replace(flags.Name(), " ", ":", -1)
			fmt.Fprint(w, name, ":usage:", err, "\n")
		}
	}
}
