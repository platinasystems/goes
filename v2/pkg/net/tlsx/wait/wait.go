// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package wait

import (
	"context"
	"strings"
	"text/template"

	"github.com/platinasystems/goes/v2/pkg/container/slice"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/os/program"
)

const Usage = `
usage: {{.}}
Wait until kill signal to hold container namespace open.
`

func Daemon(
	ctx context.Context,
	path []string,
	args ...string,
) error {
	usage := func() error {
		return template.Must(template.New("usage").Parse(Usage[1:])).
			Execute(style.Plain.Notice.Writer(),
				strings.Join(path, " "))
	}

	switch path[1] {
	case "complete":
		return nil
	case "help":
		path = slice.Cut[string](path, 1, 1)
		path[1] = "start" // replace "daemon"
		return usage()
	default:
		if !program.IsKoApp() {
			style.System()
		}
	}

	<-ctx.Done()
	return ctx.Err()
}
