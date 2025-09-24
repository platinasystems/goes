// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build linux

package goes_util

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"io"
	"os"

	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xmain"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

const SystemdTmpl = `
[Unit]
Description={{.Description}}
After={{.After}}
Wants={{.Wants}}

[Service]
Type=simple
ExecStart={{.Program}}{{range .Args}} {{.}}{{end}}
KillMode=control-group
StandardInput={{.StdIn}}
StandardOutput={{.StdOut}}
StandardError={{.StdErr}}
Restart=on-success
RestartPreventExitStatus=1
SuccessExitStatus=TEMPFAIL SIGKILL SIGTERM

[Install]
WantedBy={{.WantedBy}}
`

func init() { Shows["systemd"] = ShowSystemd }

func ShowSystemd(ctx context.Context, args []string) error {
	const oCreate = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	var m xflag.FileMode = 0664
	var w io.Writer

	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [args...]
Generate systemd config for program's named service.

{{flags .}}`)
	data := struct {
		Program,
		Description,
		After,
		Wants,
		StdIn,
		StdOut,
		StdErr,
		WantedBy string
		Args []string
	}{
		Program:     xprogram.Path(),
		Description: fmt.Sprint(xmain.PackageName(), " service"),
		After:       "network-online.target",
		Wants:       "network-online.target",
		StdIn:       "null",
		StdOut:      "journal",
		StdErr:      "inherit",
		WantedBy:    "multi-user.target",
	}

	flag.CommandLine.StringVar(&data.Description, "description",
		data.Description, "")
	flag.CommandLine.StringVar(&data.After, "after", data.After, "")
	flag.CommandLine.StringVar(&data.Wants, "wants", data.Wants, "")
	flag.CommandLine.StringVar(&data.StdIn, "stdin", data.StdIn, "")
	flag.CommandLine.StringVar(&data.StdOut, "stdout", data.StdOut, "")
	flag.CommandLine.StringVar(&data.StdErr, "stderr", data.StdErr, "")
	flag.CommandLine.StringVar(&data.WantedBy, "wanted-by",
		data.WantedBy, "")

	o := flag.CommandLine.String("o", "-", `
Writes config to the named file or standard output if "-".`[1:])
	flag.CommandLine.Var(&m, "m", "Output file mode.")

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	data.Args = flag.CommandLine.Args()

	if *o == "-" {
		w = os.Stdout
	} else if f, err := os.OpenFile(*o, oCreate, m.Mode()); err != nil {
		return err
	} else {
		defer f.Close()
		w = f
	}

	tt, err := template.New("systemd").Parse(SystemdTmpl[1:])
	if err != nil {
		return err
	}

	return tt.Execute(w, data)
}
