// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/log/oslog"
	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/override"
	"github.com/platinasystems/goes/v2/pkg/path/restricted"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
	"golang.org/x/term"
)

const (
	IntegralBashCompletion = `
_{{$arg0 := branch . 0 1}}{{$arg0}}()
{
	if [ -z ${COMP_WORDS[COMP_CWORD]} ] ; then
		COMPREPLY=( $({{$arg0}} complete ${COMP_WORDS[@]:1} '') )
	else
		COMPREPLY=( $({{$arg0}} complete ${COMP_WORDS[@]:1}) )
	fi
	return 0
}

type -p {{$arg0}} >/dev/null &&
	complete -F _{{$arg0}} -o filenames {{$arg0}}`
	IntegralZshCompletion = `
#compdef {{$arg0 := branch . 0 1}}{{$arg0}}

if [ -z ${COMP_WORDS[COMP_CWORD]} ] ; then
	COMPREPLY=( $({{$arg0}} complete ${COMP_WORDS[@]:1} '') )
else
	COMPREPLY=( $({{$arg0}} complete ${COMP_WORDS[@]:1}) )
fi

return 0`
)

var IntegralCommands = map[string]any{
	"command":  IntegralCommand,
	"complete": IntegralComplete,
	"cutoff":   IntegralCutoff,
	"help":     IntegralHelp,
	"input":    IntegralInput,
	"output":   IntegralOutput,
	"pty":      IntegralPTY,
	"start":    IntegralStart,
}

var IntegralDaemons = map[string]any{
	"standby": IntegralStandby,
}

var IntegralLoggers = map[string]any{
	"errata": IntegralErrata,
	"notice": IntegralNotice,
}

var IntegralShow = map[string]any{
	"build": program.Build,
	"completion": map[string]any{
		"bash": IntegralBashCompletion,
		"zsh":  IntegralZshCompletion,
	},
	"main": program.Main,
}

func IntegralCommand(ctx context.Context, args []string) error {
	const usage = `
usage: {{branch .}} [<options>] <command> [<args>]
Run an external command.
{{flags .}}`
	var flags flag.FlagSet
	ctx = FlagsContext(ctx, &flags)
	pFlag := flags.Bool("p", false, "Restricted path search.")
	vFlag := flags.Bool("v", false, "Report path found.")
	vvFlag := flags.Bool("V", false, "More verbose report.")
	if ContextComplete(ctx) {
		return complete.Last(args, flags)
	}
	ctx, err := ParseFlagsContext(ctx, args)
	if err != nil {
		return err
	}
	if ContextHelp(ctx) {
		return Usage(ctx, usage)
	}
	args = flags.Args()
	if len(args) == 0 {
		return ErrIncomplete
	}

	r := ContextStdin(ctx)
	w := ContextStdout(ctx)

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
	if fd := int(r.Fd()); term.IsTerminal(fd) {
		cmd.Stderr = w
		cmd.SysProcAttr, err = InteractiveSysProcAttr()
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

func IntegralComplete(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return complete.Last(args, ContextRoot(ctx))
	}
	branch := ContextBranch(ctx)
	ctx = BranchContext(ctx, branch[:len(branch)-1])
	ctx = CompleteContext(ctx, true)
	return Do(ctx, Select, args)
}

func IntegralCutoff(ctx context.Context, args []string) error {
	const usage = `
usage: {{branch .}} <duarion> <command> [<options>]
Run <command> until max <duration>.`
	if ContextComplete(ctx) {
		return Do(ctx, Select, args)
	}
	if ContextHelp(ctx) {
		if len(args) == 0 {
			return Usage(ctx, usage)
		}
		return Do(ctx, Select, args)
	}
	if len(args) < 2 {
		return ErrIncomplete
	}
	timeout, err := time.ParseDuration(args[0])
	if err != nil {
		return err
	}
	toctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return Do(toctx, Select, args[1:])
}

func IntegralErrata(ctx context.Context, args []string) error {
	const usage = `
usage: {{branch .}} [<message>]...
Log space separated message or stdin as errata.`
	if ContextComplete(ctx) {
		return nil
	}
	if ContextHelp(ctx) {
		return Usage(ctx, usage)
	}
	w, err := oslog.OpenError()
	if err != nil {
		return err
	}
	defer w.Close()
	if len(args) > 0 {
		fmt.Fprintln(w, strings.Join(args, " "))
		return nil
	}
	data, err := io.ReadAll(ContextStdin(ctx))
	if err == nil {
		_, err = w.Write(data)
	}
	return err
}

func IntegralHelp(ctx context.Context, args []string) error {
	const usage = `
{{$cmd := branch . 0 1 -}}
{{$branch := branch . 1 -}}
{{if eq $branch "help"}}
{{- $branch = ""}}
{{- end -}}
{{if $branch}}
{{- $branch = print $branch " "}}
{{- end -}}
{{$trunk := branch . 1 2 -}}
{{$args := "<command> [<args>]" -}}
{{$syn := "Run command" -}}
{{$heading := "Commands" -}}
{{if eq $trunk "show"}}
{{- $args = "<object>"}}
{{- $syn = "Show object"}}
{{- $heading = "Objects"}}
{{- end -}}
usage: {{$cmd}} [option] {{$branch}}{{$args}}
{{$syn}}.
{{if eq $branch ""}}
Options{{flags .}}
{{- end}}
{{$heading}}
{{root . "daemon"}}`
	if len(args) == 0 {
		return Usage(ctx, usage)
	}
	branch := ContextBranch(ctx)
	ctx = BranchContext(ctx, branch[:len(branch)-1])
	ctx = HelpContext(ctx, true)
	return Do(ctx, Select, args)
}

func IntegralInput(ctx context.Context, args []string) error {
	const usage = `
usage: {{branch .}} <file> <command> [<options>]
Run <command> with <file> input.`
	if ContextComplete(ctx) {
		return Do(ctx, Select, args)
	}
	if ContextHelp(ctx) {
		if len(args) == 0 {
			return Usage(ctx, usage)
		}
		return Do(ctx, Select, args)
	}
	if len(args) < 2 {
		return ErrIncomplete
	}

	r, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer r.Close()
	defer override.Value(&os.Stdin, r)()
	return Do(StdinContext(ctx, r), Select, args[1:])
}

func IntegralNotice(ctx context.Context, args []string) error {
	const usage = `
usage: {{branch .}} [<message>]...
Log space separated message or stdin as notice.`
	if ContextComplete(ctx) {
		return nil
	}
	if ContextHelp(ctx) {
		return Usage(ctx, usage)
	}

	w, err := oslog.OpenNotice()
	if err != nil {
		return err
	}
	defer w.Close()
	if len(args) > 0 {
		fmt.Fprintln(w, strings.Join(args, " "))
		return nil
	}
	data, err := io.ReadAll(ContextStdin(ctx))
	if err == nil {
		_, err = w.Write(data)
	}
	return err
}

func IntegralLogDaemon(ctx context.Context, args []string) error {
	const usage = `
usage: {{branch .}} <daemon> [<options>]
Fork self again to pipe <daemon> output to syslog.

   {{branch . 0 1}} daemon <daemon> [<options>]

Daemons
{{root .}}`
	if ContextComplete(ctx) {
		return Do(ctx, Select, args)
	}
	if ContextHelp(ctx) {
		if len(args) == 0 {
			return Usage(ctx, usage)
		}
		return Do(ctx, Select, args)
	}
	if len(args) == 0 {
		return ErrIncomplete
	}

	errLog, err := oslog.OpenError()
	if err != nil {
		return err
	}
	defer errLog.Close()

	defer func() {
		if err != nil {
			fmt.Fprintln(errLog, err)
		}
	}()

	outLog, err := oslog.OpenNotice()
	if err != nil {
		return err
	}
	defer outLog.Close()

	cmd := exec.CommandContext(ctx, program.Executable(),
		append([]string{"daemon"}, args...)...)
	cmd.Stdin = nil

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	defer errPipe.Close()

	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	defer outPipe.Close()

	if err = cmd.Start(); err != nil {
		return err
	}

	go loglines(outLog, outPipe)

	errData, err := io.ReadAll(errPipe)
	errLog.Write(errData)

	err = cmd.Wait()
	return err
}

func loglines(w io.Writer, r io.Reader) {
	for sc := bufio.NewScanner(r); sc.Scan(); {
		w.Write(sc.Bytes())
	}
}

func IntegralOutput(ctx context.Context, args []string) error {
	const usage = `
usage: {{branch .}} [<modifiers>] <file> <command> [<options>]
Run <command> with output to <file>.

Modifiers{{flags .}}`
	flags := flag.NewFlagSet("", 0)
	aFlag := flags.Bool("a", false, "Append <file> instead of truncate.")
	tFlag := flags.Bool("t", false, "Tee to <file> and stdout.")
	mFlag := flags.Uint("m", 0, "Output file mode (default 0666).")
	fctx := FlagsContext(ctx, flags)

	ctx, err := ParseFlagsContext(fctx, args)
	if err != nil {
		return err
	}
	if ContextComplete(fctx) {
		return Do(ctx, Select, args)
	}
	if ContextHelp(fctx) {
		if len(args) == 0 {
			return Usage(fctx, usage)
		}
		return Do(ctx, Select, args)
	}
	if args = flags.Args(); len(args) < 2 {
		return ErrIncomplete
	}
	oflags := os.O_RDWR | os.O_CREATE
	if *aFlag {
		oflags |= os.O_APPEND
	} else {
		oflags |= os.O_TRUNC
	}
	mode := os.FileMode(0666)
	if *mFlag != 0 {
		mode = os.FileMode(*mFlag)
	}
	f, err := os.OpenFile(args[0], oflags, mode)
	if err != nil {
		return err
	}
	defer f.Close()
	if *tFlag {
		pr, pw, err := os.Pipe()
		if err != nil {
			return err
		}
		defer pw.Close()
		defer override.Value(&os.Stdout, pw)()
		go io.Copy(io.MultiWriter(os.Stdout, f), pr)
		ctx = StdoutContext(ctx, pw)
	} else {
		defer override.Value(&os.Stdout, f)()
		ctx = StdoutContext(ctx, f)
	}
	return Do(ctx, Select, args[1:])
}

// Use this to hold container until interrupt or termination signal.
func IntegralStandby(ctx context.Context, args []string) error {
	const usage = `
usage: {{branch .}}
Wait until interrupt or termination signal.`
	if ContextComplete(ctx) {
		return nil
	}
	if ContextHelp(ctx) {
		return Usage(ctx, usage)
	}
	<-ctx.Done()
	return ctx.Err()
}

func IntegralStart(ctx context.Context, args []string) error {
	const usage = `
usage: {{branch .}} <daemon> [<options>]
Detach self to run <daemon> like this,

   {{branch . 0 1}} log-daemon <daemon> [<options>]

Daemons
{{root .}}`
	vdaemon, ok := ContextRoot(ctx)["daemon"]
	if !ok {
		return egress.Mark(FIXME)
	}
	daemons, ok := vdaemon.(map[string]any)
	if !ok {
		return egress.Mark(FIXME)
	}
	ctx = RootContext(ctx, daemons)
	if ContextComplete(ctx) {
		if len(args) == 0 {
			return complete.Last(args, daemons)
		}
		return Do(ctx, Select, args)
	}
	if ContextHelp(ctx) {
		if len(args) == 0 {
			return Usage(ctx, usage)
		}
		return Do(ctx, Select, args)
	}
	if len(args) == 0 {
		return ErrIncomplete
	}
	cmd := exec.Command(program.Executable(),
		append([]string{"log-daemon"}, args...)...)
	cmd.Env = DaemonEnv()
	cmd.Dir = program.RunTimeDir()
	_, err := os.Stat(cmd.Dir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		cmd.Dir = os.TempDir()
	}
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	if cmd.SysProcAttr, err = DaemonSysProcAttr(); err == nil {
		err = cmd.Start()
	}
	return err
}

func DaemonEnv() []string {
	env := []string{
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
		"XDG_CONFIG_DIRS",
		"XDG_CONFIG_HOME",
		"XDG_DATA_DIRS",
		"XDG_DATA_HOME",
		"XDG_RUNTIME_DIR",
		"XDG_STATE_HOME",
	} {
		if val, ok := os.LookupEnv(name); ok {
			env = append(env, fmt.Sprint(name, "=", val))
		}
	}
	return env
}
