// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package x509keys

import (
	"sync"
	"text/template"
)

var Template = sync.OnceValues(func() (*template.Template, error) {
	return template.New("signatures").Parse(`{{/*
*/}}{{ $n := len .}}{{if eq $n 0}}# none
{{else}}{{range .}}{{if .}}{{printf "# %T" .}}
{{end}}{{end}}{{end}}`)
})
