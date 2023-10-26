// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/platinasystems/goes/v2/pkg/context/help"
	"github.com/platinasystems/goes/v2/pkg/flag"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/path/restricted"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
	"golang.org/x/term"
)

// These are merged into the Root of the command tree.
var Integral = map[string]any{
	"append":   IntegralOutput,
	"command":  IntegralCommand,
	"complete": IntegralComplete,
	"help":     IntegralHelp,
	"input":    IntegralInput,
	"output":   IntegralOutput,
	"standby":  IntegralStandby,
	"start":    IntegralStart,
	"tee":      IntegralOutput,
	"timeout":  IntegralTimeout,
}

func IntegralCommand(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	const usage = `{{$path := join .Path " "}}{{/*
*/}}usage: {{$path}} [<options>] <command> [<args>]
Run external command.
{{print .Flags}}`
	fs, hFlag := flag.New()
	pFlag := fs.Bool("p", false, "Restricted path search.")
	vFlag := fs.Bool("v", false, "Report path found.")
	vvFlag := fs.Bool("V", false, "More verbose report.")
	if complete.Parameter.Value(ctx) {
		style.Completions(args, fs.FlagSet)
		return nil
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	if help.Parameter.Value(ctx) || *hFlag {
		return style.Usage(usage, struct {
			Path  []string
			Flags fmt.Formatter
		}{path, fs})
	}
	args = fs.Args()
	if len(args) == 0 {
		return ErrIncomplete
	}

	lookpath := exec.LookPath
	if *pFlag {
		lookpath = restricted.LookPath
	}
	full, err := lookpath(args[0])
	if err != nil {
		return err
	}
	if *vFlag {
		fmt.Fprintln(w, full)
		return nil
	}
	if *vvFlag {
		fmt.Fprintln(w, args[0], "is", full)
		return nil
	}
	stderr := new(strings.Builder)
	cmd := exec.CommandContext(ctx, full, args[1:]...)
	cmd.Stdin = r
	cmd.Stdout = w
	cmd.Stderr = stderr
	if method, ok := r.(interface{ Fd() uintptr }); ok {
		if fd := int(method.Fd()); term.IsTerminal(fd) {
			cmd.Stderr = w
			switch runtime.GOOS {
			case "linux":
				cmd.SysProcAttr = &syscall.SysProcAttr{
					Setsid:  true,
					Setctty: true,
				}
			case "darwin":
				cmd.SysProcAttr = &syscall.SysProcAttr{
					Setsid: true,
					// FIXME can't Setctty on darwin
					// Setctty: true,
					// Ctty:    0,
				}
			}
		}
	}
	if err = cmd.Start(); err == nil {
		err = cmd.Wait()
	}
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok && stderr.Len() > 0 {
			err = errors.New(stderr.String())
		}
	}
	return err
}

func IntegralComplete(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	m map[string]any,
	args ...string,
) error {
	ctx = complete.Parameter.With(ctx, true)
	path = path[:len(path)-1]
	if len(args) == 0 {
		style.Completions(args, m)
		return nil
	}
	return Select(ctx, r, w, path, m, args...)
}

func IntegralHelp(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	m map[string]any,
	args ...string,
) error {
	const usage = `{{/*
*/}}usage: {{join .Path " "}} <command|object> [<args>]

Command/Objects
{{keys .Map}}`
	path = path[:len(path)-1]
	if len(args) == 0 {
		return style.Usage(usage, struct {
			Path []string
			Map  map[string]any
		}{path, m})
	}
	return do(Select, help.Parameter.With(ctx, true), r, w, path, m, args...)
}

func IntegralInput(
	ctx context.Context,
	w io.Writer,
	path []string,
	m map[string]any,
	args ...string,
) error {
	const usage = `{{/*
*/}}usage: {{join . " "}} <file> <command|object> [<args>]

Perform command or set object with input from named file.
`
	if complete.Parameter.Value(ctx) {
		if len(args) < 2 {
			style.Completions(args, "*")
			return nil
		}
		return do(Select, ctx, nil, w, path, m, args[1:]...)
	}
	if help.Parameter.Value(ctx) {
		if len(args) < 2 {
			return style.Usage(usage, path)
		}
		return do(Select, ctx, nil, w, path, m, args[1:]...)
	}
	if len(args) < 2 {
		return ErrIncomplete
	}
	f, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer f.Close()
	return do(Select, ctx, f, w, path, m, args[1:]...)
}

func IntegralOutput(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	m map[string]any,
	args ...string,
) error {
	const usage = `{{/*
*/}}usage: {{join .Path " "}} [-m <mode>] <file> <command|object> [<args>]

Perform command or print object with output directed to named file.
{{print .Flags}}`
	cmd := path[len(path)-1]
	fs, hFlag := flag.New()
	aFlag := fs.Bool("a", cmd == "append", "Append output to named file.")
	mFlag := fs.Uint("m", 0666, "File mode (default 0666).")
	tFlag := fs.Bool("t", cmd == "tee", "Tee output to named file.")
	if complete.Parameter.Value(ctx) {
		if len(args) < 2 {
			style.Completions(args, "*")
			return nil
		}
		return do(Select, ctx, r, w, path, m, args[1:]...)
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	if help.Parameter.Value(ctx) || *hFlag {
		if len(args) < 2 {
			return style.Usage(usage, struct {
				Path  []string
				Flags fmt.Formatter
			}{path, fs})
		}
		return do(Select, ctx, r, w, path, m, args[1:]...)
	}
	if len(args) < 2 {
		return ErrIncomplete
	}
	fflags := os.O_RDWR | os.O_CREATE
	if *aFlag {
		fflags |= os.O_APPEND
	} else {
		fflags |= os.O_TRUNC
	}
	fmode := os.FileMode(*mFlag)
	f, err := os.OpenFile(args[0], fflags, fmode)
	if err != nil {
		return err
	}
	defer f.Close()
	if *tFlag {
		w = io.MultiWriter(w, f)
	} else {
		w = f
	}
	return do(Select, ctx, r, w, path, m, args[1:]...)
}

var completionShellScript = map[string]string{
	"bash": `{{/*
*/}}_{{.}} ()
{
	if [ -z ${COMP_WORDS[COMP_CWORD]} ] ; then
		COMPREPLY=($({{.}} complete ${COMP_WORDS[@]:1} ''))
	else
		COMPREPLY=($({{.}} complete ${COMP_WORDS[@]:1}))
	fi
	return 0
}

type -p {{.}} >/dev/null && complete -F _{{.}} -o filenames {{.}}
`,
	"zsh": `{{/*
*/}}#compdef {{.}}

if [ -z ${COMP_WORDS[COMP_CWORD]} ] ; then
	COMPREPLY=( $({{.}} complete ${COMP_WORDS[@]:1} '') )
else
	COMPREPLY=( $({{.}} complete ${COMP_WORDS[@]:1}) )
fi

return 0
`,
}

func IntegralShowCompletion(
	ctx context.Context,
	w io.Writer,
	path []string,
	args ...string,
) error {
	const usage = `{{/*
*/}}usage: {{join . " "}} <shell>
Print shell completion script.
`
	if complete.Parameter.Value(ctx) {
		style.Completions(args, completionShellScript)
		return nil
	}
	if help.Parameter.Value(ctx) {
		return style.Usage(usage, path)
	}
	if len(args) == 0 {
		return ErrIncomplete
	}
	for _, shell := range args {
		script, ok := completionShellScript[shell]
		if !ok {
			return fmt.Errorf("%s: %w", shell, ErrNotFound)
		}
		t, err := template.New(shell).Parse(script)
		if err != nil {
			return err
		}
		if err = t.Execute(w, path[0]); err != nil {
			return err
		}
	}
	return nil

}

func IntegralShowFS(
	ctx context.Context,
	w io.Writer,
	path []string,
	efs embed.FS,
	args ...string,
) error {
	const usage = `{{/*
*/}}usage: {{join . " "}}
Print embedded file.
`
	if complete.Parameter.Value(ctx) {
		if len(args) == 0 {
			// exact path match
			style.Plain.Notice.Println(path[len(path)-1])
		}
		return nil
	}
	if help.Parameter.Value(ctx) {
		return style.Usage(usage, path)
	}
	b, err := efs.ReadFile(path[len(path)-1])
	if err == nil {
		_, err = w.Write(b)
	}
	return err
}

// Use this to hold container until interrupt or termination signal.
func IntegralStandby(
	ctx context.Context,
	w io.Writer,
	path []string,
	args ...string,
) error {
	const usage = `usage: {{join . " "}} [<pids>]
Wait until interrupt or termination signal.
`
	if complete.Parameter.Value(ctx) {
		return nil
	}
	if help.Parameter.Value(ctx) {
		return style.Usage(usage, path)
	}
	<-ctx.Done()
	return ctx.Err()
}

func IntegralStart(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	m map[string]any,
	args ...string,
) error {
	const usage = `{{$path := join .Path " "}}{{/*
*/}}usage: {{$path}} <daemon> [<options>]
Fork self to run <daemon> like this,

	{{$path}} daemon <daemon> [<options>]

Daemons
{{keys .Map}}`
	daemons := m["daemon"].(map[string]any)
	if complete.Parameter.Value(ctx) {
		if len(args) == 0 {
			style.Completions(args, daemons)
			return nil
		}
		return do(Select, ctx, r, w, path, daemons, args...)
	}
	if help.Parameter.Value(ctx) {
		if len(args) == 0 {
			return style.Usage(usage, struct {
				Path []string
				Map  map[string]any
			}{path, daemons})
		}
		return do(Select, ctx, r, w, path, daemons, args...)
	}
	if os.Getpid() == 1 {
		dpath := []string{path[0], "daemon"}
		return do(Select, ctx, r, w, dpath, daemons, args...)
	}
	u, err := user.Current()
	if err != nil {
		return err
	}
	cred := &syscall.Credential{NoSetGroups: true}
	if _, err = fmt.Sscan(u.Uid, &cred.Uid); err != nil {
		return fmt.Errorf("user:uid: %w", err)
	}
	if _, err = fmt.Sscan(u.Gid, &cred.Gid); err != nil {
		return fmt.Errorf("user:gid: %w", err)
	}
	args = append([]string{"daemon"}, args...)
	cmd := exec.Command(program.Executable(), args...)
	cmd.Env = []string{
		Path(),
	}
	for _, name := range []string{
		"AppData",
		"LocalAppData",
		"home",
		"HOME",
		"TMPDIR",
		"USERPROFILE",
		"XDG_CACHE_HOME",
	} {
		if val, ok := os.LookupEnv(name); ok {
			cmd.Env = append(cmd.Env, fmt.Sprint(name, "=", val))
		}
	}
	if os.Geteuid() == 0 {
		cmd.Dir = "/var/run"
	} else if d, err := os.UserCacheDir(); err == nil {
		cmd.Dir = d
	}
	if len(cmd.Dir) == 0 || func(s string) error {
		_, err := os.Stat(s)
		return err
	}(cmd.Dir) != nil {
		cmd.Dir = os.TempDir()
	}
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: cred,
		Setsid:     true,
	}
	if err = cmd.Start(); err == nil {
		fmt.Fprint(w, program.Base(), ":daemon:", args[0],
			":pid: ", cmd.Process.Pid, "\n")
	}
	return err
}

func IntegralTimeout(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	m map[string]any,
	args ...string,
) error {
	const usage = `{{/*
*/}}usage: {{join .Path " "}} <duration> <command|object> [<args>]

Command/Objects
{{keys .Map}}`
	if complete.Parameter.Value(ctx) {
		if len(args) < 2 {
			return nil
		}
		return do(Select, ctx, r, w, path, m, args[1:]...)
	}
	if help.Parameter.Value(ctx) {
		if len(args) < 2 {
			return style.Usage(usage, struct {
				Path []string
				Map  map[string]any
			}{path, m})
		}
		return do(Select, ctx, r, w, path, m, args[1:]...)
	}
	if len(args) < 2 {
		return ErrIncomplete
	}
	dur, err := time.ParseDuration(args[0])
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, dur)
	defer cancel()
	return do(Select, ctx, r, w, path, m, args[1:]...)
}
