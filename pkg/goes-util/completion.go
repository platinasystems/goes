// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes_util

import (
	_ "embed"
	"os"
	"path/filepath"
	"strings"

	"text/template"
)

//go:embed completion.bash.tmpl
var bashTmpl string

//go:embed completion.zsh.tmpl
var zshTmpl string

type CompletionTemplate string

var ShowCompletion = map[string]any{
	"bash": CompletionTemplate(bashTmpl),
	"zsh":  CompletionTemplate(zshTmpl),
}

func (ct CompletionTemplate) String() string {
	var sb strings.Builder
	tt, err := template.New("completion").Parse(string(ct))
	if err != nil {
		return err.Error()
	}
	exe, err := os.Executable()
	if err != nil {
		return err.Error()
	}
	err = tt.Execute(&sb, filepath.Base(exe))
	if err != nil {
		return err.Error()
	}
	return sb.String()
}
