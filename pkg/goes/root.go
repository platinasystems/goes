// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/context/parameter"
)

var Root = map[string]any{
	"command":  ExternalCommand,
	"complete": InitEnabledRootFunc,
	"daemon":   Daemons,
	"help":     InitEnabledRootFunc,
	"show":     Show,
	"start":    InitEnabledRootFunc,
}

func ContextRoot(ctx context.Context) map[string]any {
	return parameter.Value(ctx, &Root)
}

func RootContext(ctx context.Context, root map[string]any) context.Context {
	return parameter.Context(ctx, &Root, root)
}

// The InitEnabled functions work around initialization cycle loop detection.
func InitEnabledRootFunc(context.Context, []string) error {
	return ErrDisabled
}

func SprintContextRootKeys(ctx context.Context, filter ...string) string {
	var sb strings.Builder
	m := ContextRoot(ctx)
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		ok := true
		for _, x := range filter {
			if k == x {
				ok = false
			}
		}
		if ok {
			fmt.Fprintln(&sb, " ", k)
		}
	}
	return sb.String()
}

func init() {
	Root["complete"] = Complete
	Root["help"] = Help
	Root["start"] = Start
}
