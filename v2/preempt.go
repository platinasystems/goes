// Copyright © 2015-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import "context"

var preemptive = map[string]bool{
	"complete": true,
	"help":     true,
}

// Preemption returns "complete" or "help" if the context path is
// preempted by either; otherwise, this returns an empty string.
func Preemption(ctx context.Context) string {
	path := PathOf(ctx)
	if len(path) > 1 && preemptive[path[1]] {
		return path[1]
	}
	return ""
}

// Move the "complete" and "help" preeemptive command arguments to context.
func Preempt(ctx context.Context, args []string) (context.Context, []string) {
	for len(args) > 0 && preemptive[args[0]] {
		ctx = WithPath(ctx, args[0])
		args = args[1:]
	}
	return ctx, args
}
