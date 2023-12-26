// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import "context"

// Parse and execute a text/template with context data and TemplateFuncs; then
// return as usage error.
func Usage(ctx context.Context, tt string) error {
	return usageError(SprintTemplate(ctx, tt))
}

func IsUsage(err error) bool {
	_, ok := err.(usageError)
	return ok
}

type usageError string

func (e usageError) Error() string { return string(e) }
