// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"strings"
	"text/template"

	"github.com/platinasystems/goes/v2/pkg/context/parameter"
)

var TemplateFuncs = template.FuncMap{
	"branch": SprintContextBranch,
	"flags":  SprintContextFlags,
	"root":   InitEnabledTemplateFunc,
}

func init() {
	TemplateFuncs["root"] = SprintContextRootKeys
}

func InitEnabledTemplateFunc(context.Context, ...string) string {
	return ""
}

func ContextTemplateFuncs(ctx context.Context) template.FuncMap {
	return parameter.Value(ctx, &TemplateFuncs)
}

func TemplateFuncsContext(
	ctx context.Context,
	funcs template.FuncMap,
) context.Context {
	return parameter.Context(ctx, &TemplateFuncs, funcs)
}

// Parse and execute a text/template with context data and TemplateFuncs; then
// return as string.
func SprintTemplate(ctx context.Context, tt string) string {
	var sb strings.Builder
	branch := ContextBranch(ctx)
	funcs := ContextTemplateFuncs(ctx)
	t, err := template.New(branch[len(branch)-1]).Funcs(funcs).
		Parse(strings.TrimSpace(tt))
	if err != nil {
		return err.Error()
	}
	err = t.Execute(&sb, ctx)
	if err != nil {
		return err.Error()
	}
	return sb.String()
}
