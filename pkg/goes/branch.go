// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"fmt"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/context/parameter"
	"github.com/platinasystems/goes/v2/pkg/os/program"
)

var branchParameter = []string{program.Base()}

func ContextBranch(ctx context.Context) []string {
	return parameter.Value(ctx, &branchParameter)
}

func AppendBranchContext(ctx context.Context, s string) context.Context {
	return parameter.Append(ctx, &branchParameter, s)
}

func BranchContext(ctx context.Context, branch []string) context.Context {
	return parameter.Context(ctx, &branchParameter, branch)
}

func SprintContextBranch(ctx context.Context, begend ...int) string {
	branch := ContextBranch(ctx)
	beg, end := 0, len(branch)
	if len(begend) > 0 {
		if i := begend[0]; i > end {
			beg = end
		} else if i >= 0 {
			beg = i
		}
		if len(begend) > 1 {
			if i := begend[1]; i > beg && i < end {
				end = i
			}
		}
	}
	return strings.Join(branch[beg:end], " ")
}

func Mark(ctx context.Context, err error) error {
	if err != nil {
		err = MarkError{ContextBranch(ctx), err}
	}
	return err
}

type MarkError struct {
	branch []string
	err    error
}

func IsMarked(err error) bool {
	_, ok := err.(MarkError)
	return ok
}

func (e MarkError) Error() string {
	var sb strings.Builder
	for _, s := range e.branch {
		fmt.Fprint(&sb, s, ":")
	}
	eee := e.err.Error()
	colon := strings.IndexRune(eee, ':')
	space := strings.IndexRune(eee, ' ')
	if colon < 0 || (space > 0 && space < colon) {
		fmt.Fprint(&sb, " ")
	}
	fmt.Fprint(&sb, eee)
	return sb.String()
}

func (m MarkError) Unwrap() error { return m.err }
