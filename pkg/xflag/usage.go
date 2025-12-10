// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xflag

import (
	"flag"
	"fmt"
	"strings"
	"text/template"
)

// These results are passed to the usage template execution.
var UsageData = func(flags *flag.FlagSet) any {
	return flags
}

// Each usage template includes these functions.
var UsageFuncs = template.FuncMap{
	"flags": SprintDefaults,
}

// Redirect [FlagSet.PrintDefaults] to string.
func SprintDefaults(flags *flag.FlagSet) string {
	save := flags.Output()
	defer flags.SetOutput(save)
	var sb strings.Builder
	flags.SetOutput(&sb)
	flags.PrintDefaults()
	return sb.String()
}

// [TemplateUsageIn] [flag.CommandLine]
func TemplateUsage(tmpl string) {
	TemplateUsageIn(flag.CommandLine, tmpl)
}

// Assign [FlagSet.Usage] to a closure that creates a new [text/template]
// with [UsageFuncs]; parses “tmpl”; then [text/template.Template.Execute]'s
// to [FlagSet.Output] with [UsageData].
func TemplateUsageIn(flags *flag.FlagSet, tmpl string) {
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
