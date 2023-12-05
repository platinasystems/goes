// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package usage

import (
	"fmt"
	"strings"
	"text/template"
)

type usageError string

func InError(err error) bool {
	_, ok := err.(usageError)
	return ok
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

func (e usageError) Error() string { return string(e) }
