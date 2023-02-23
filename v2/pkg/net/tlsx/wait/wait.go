// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package wait

import (
	"context"
	"text/template"

	"github.com/platinasystems/goes/v2/pkg/container/slice"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/os/program"
)

const Usage = `
usage: {{.}} wait
Wait until kill signal to hold container namespace open.
`

func Daemon(
	ctx context.Context,
	path []string,
	args ...string,
) error {
	if !program.IsKoApp() {
		style.System()
	}

	usage := func() error {
		return template.Must(template.New("usage").Parse(Usage[1:])).
			Execute(style.Plain.Notice.Writer(), path[0])
	}

	switch path[1] {
	case "complete":
		return nil
	case "help":
		path = slice.Cut[string](path, 1, 1)
		return usage()
	}

	<-ctx.Done()
	return ctx.Err()
}
