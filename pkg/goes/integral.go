// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/path/restricted"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
	"golang.org/x/term"
)

const (
	BashCompletion = `
_{{$arg0 := branch . 0 1}}{{$arg0}}()
{
	if [ -z ${COMP_WORDS[COMP_CWORD]} ] ; then
		COMPREPLY=($({{$arg0}} -complete ${COMP_WORDS[@]:1} ''))
	else
		COMPREPLY=($({{$arg0}} -complete ${COMP_WORDS[@]:1}))
	fi
	return 0
}

type -p {{$arg0}} >/dev/null &&
	complete -F _{{$arg0}} -o filenames {{$arg0}}`
	ZshCompletion = `
#compdef {{$arg0 := branch . 0 1}}{{$arg0}}

if [ -z ${COMP_WORDS[COMP_CWORD]} ] ; then
	COMPREPLY=( $({{$arg0}} -complete ${COMP_WORDS[@]:1} '') )
else
	COMPREPLY=( $({{$arg0}} -complete ${COMP_WORDS[@]:1}) )
fi

return 0`
)

var Completion = map[string]any{
	"bash": BashCompletion,
	"zsh":  ZshCompletion,
}

var Daemons = map[string]any{
	"standby": Standby,
}

var Show = map[string]any{
	"build":      program.Build,
	"main":       program.Main,
	"completion": Completion,
}

func Complete(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return complete.Last(args, ContextRoot(ctx))
	}
	branch := ContextBranch(ctx)
	ctx = BranchContext(ctx, branch[:len(branch)-1])
	ctx = CompleteContext(ctx, true)
	return do(ctx, Select, args)
}

func Help(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return Usage(ctx, `
usage: {{branch . 0 1}} [option] {{branch . 1}} <command|object> [<args>]
{{synopsis .}}

Options{{flags .}}
Command/Objects
{{root . "daemon"}}`)
	}
	branch := ContextBranch(ctx)
	ctx = BranchContext(ctx, branch[:len(branch)-1])
	ctx = HelpContext(ctx, true)
	return do(ctx, Select, args)
}

// Use this to hold container until interrupt or termination signal.
func Standby(ctx context.Context, args []string) error {
	if ContextComplete(ctx) {
		return nil
	}
	if ContextHelp(ctx) {
		return Usage(ctx, `
usage: {{branch .}} [<pids>]
Wait until interrupt or termination signal.`)
	}
	<-ctx.Done()
	return ctx.Err()
}

func ExternalCommand(ctx context.Context, args []string) error {
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
		return Usage(ctx, `
usage: {{branch .}} [<options>] <command> [<args>]
Run an external command.
{{flags .}}`)
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

func Start(ctx context.Context, args []string) error {
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
		return do(ctx, Select, args)
	}
	if ContextHelp(ctx) {
		if len(args) == 0 {
			return Usage(ctx, `
usage: {{branch .}} <daemon> [<options>]
Fork self to run <daemon> like this,

   {{branch . 0 1}} daemon <daemon> [<options>]

Daemons
{{root .}}`)
		}
		return do(ctx, Select, args)
	}
	if os.Getpid() == 1 {
		ctx = AppendBranchContext(ctx, "daemon")
		return do(ctx, Select, args)
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
	} else if d, undetermined := os.UserCacheDir(); undetermined == nil {
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
	var err error
	cmd.SysProcAttr, err = DaemonSysProcAttr()
	if err == nil {
		err = cmd.Start()
		if err == nil {
			w := ContextStdout(ctx)
			branch := ContextBranch(ctx)
			fmt.Fprint(w, branch[0], ":daemon:", args[0],
				":pid: ", cmd.Process.Pid,
				"\n")
		}
	}
	return err
}
