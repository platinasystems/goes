// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package style

import (
	"fmt"
	"sort"
	"strings"
	"text/template"
)

var UsageTemplateFuncs = template.FuncMap{
	"keys": MapKeys,
	"join": strings.Join,
}

func Usage(text string, data any) error {
	t, err := template.New("usage").Funcs(UsageTemplateFuncs).Parse(text)
	if err != nil {
		return err
	}
	return t.Execute(Plain.Notice.Writer(), data)
}

func MapKeys(m map[string]any) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		if k != "daemon" && !strings.HasPrefix(k, "_") {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var sb strings.Builder
	for _, k := range keys {
		fmt.Fprintln(&sb, " ", k)
	}
	return sb.String()
}
