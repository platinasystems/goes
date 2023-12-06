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
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/context/ctxparm"
	"github.com/platinasystems/goes/v2/pkg/errors/usage"
	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/path/restricted"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
	"golang.org/x/term"
)

// These are merged into the Root of the command tree.
var Integral = map[string]any{
	"command":  IntegralCommand,
	"complete": IntegralComplete,
	"help":     IntegralHelp,
	"standby":  IntegralStandby,
	"start":    IntegralStart,
}

const IntegralCommandUsageTemplate = `
usage: {{.Path}} [<options>] <command> [<args>]
Run an external command.
{{.Flag}}`

func IntegralCommandUsageData(ctx context.Context) any {
	return struct{ Path, Flag string }{
		Path: strings.Join(ctxparm.Strings.In(ctx), " "),
		Flag: ctxparm.SprintFlagsIn(ctx),
	}
}

func IntegralCommand(ctx context.Context, args ...string) error {
	flags := usage.NewFlags("command")
	ctx = ctxparm.Flags.With(ctx, flags)
	pFlag := flags.Bool("p", false, "Restricted path search.")
	vFlag := flags.Bool("v", false, "Report path found.")
	vvFlag := flags.Bool("V", false, "More verbose report.")
	if *complete.Help {
		return complete.Last(args, flags)
	}
	err := flags.Parse(args)
	if err != nil {
		return err
	}
	if *usage.Help {
		return usage.Error(IntegralCommandUsageTemplate[1:],
			IntegralCommandUsageData(ctx))
	}
	args = flags.Args()
	if len(args) == 0 {
		return ErrIncomplete
	}
	ctx = ctxparm.Flags.With(ctx, flags)
	r := ctxparm.Reader.In(ctx)
	w := ctxparm.Writer.In(ctx)

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

func IntegralComplete(ctx context.Context, args ...string) error {
	if len(args) == 0 {
		return complete.Last(args, ctxparm.MapKeysIn(ctx))
	}
	path := ctxparm.Strings.In(ctx)
	ctx = ctxparm.Strings.With(ctx, path[:len(path)-1])
	*complete.Help = true
	return Select(ctx, args...)
}

const IntegralHelpUsageTemplate = `
usage: {{.Path}} [option] <command|object> [<args>],
Show command or object's help text.

Options{{.Flag}}
Command/Objects
{{.Commands}}`

func IntegralHelpUsageData(ctx context.Context) any {
	return struct{ Path, Flag, Commands string }{
		Path:     strings.Join(ctxparm.Strings.In(ctx), " "),
		Flag:     ctxparm.SprintFlagsIn(ctx),
		Commands: ctxparm.MapKeysIn(ctx),
	}
}

func IntegralHelp(ctx context.Context, args ...string) error {
	if len(args) == 0 {
		return usage.Error(IntegralHelpUsageTemplate[1:],
			IntegralHelpUsageData(ctx))
	}
	path := ctxparm.Strings.In(ctx)
	ctx = ctxparm.Strings.With(ctx, path[:len(path)-1])
	*usage.Help = true
	return do(ctx, Select, args...)
}

var completionShellScript = map[string]string{
	"bash": `
_{{.}} ()
{
	if [ -z ${COMP_WORDS[COMP_CWORD]} ] ; then
		COMPREPLY=($({{.}} -complete ${COMP_WORDS[@]:1} ''))
	else
		COMPREPLY=($({{.}} -complete ${COMP_WORDS[@]:1}))
	fi
	return 0
}

type -p {{.}} >/dev/null && complete -F _{{.}} -o filenames {{.}}
`[1:],
	"zsh": `
#compdef {{.}}

if [ -z ${COMP_WORDS[COMP_CWORD]} ] ; then
	COMPREPLY=( $({{.}} -complete ${COMP_WORDS[@]:1} '') )
else
	COMPREPLY=( $({{.}} -complete ${COMP_WORDS[@]:1}) )
fi

return 0
`[1:],
}

const IntegralShowCompletionUsageTemplate = `
usage: {{.}} <shell>
Print shell completion script.

Shells
  bash
  zsh`

func IntegralShowCompletionUsageData(ctx context.Context) any {
	return ctxparm.Strings.In(ctx)
}

func IntegralShowCompletion(ctx context.Context, args ...string) error {
	path := ctxparm.Strings.In(ctx)
	if *complete.Help {
		return complete.Last(args, completionShellScript)
	}
	if *usage.Help {
		return usage.Error(IntegralShowCompletionUsageTemplate[1:],
			IntegralShowCompletionUsageData(ctx))
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
		err = t.Execute(ctxparm.Writer.In(ctx), path[0])
		if err != nil {
			return err
		}
	}
	return nil

}

const IntegralShowFSUsageTemplate = `
usage: {{.}}
Print embedded file.`

var IntegralShowFSUsageData = IntegralShowCompletionUsageData

func IntegralShowFS(ctx context.Context, efs embed.FS, args ...string) error {
	path := ctxparm.Strings.In(ctx)
	if *complete.Help {
		if len(args) == 0 {
			// exact path match
			fmt.Println(path[len(path)-1])
		}
		return nil
	}
	if *usage.Help {
		return usage.Error(IntegralShowFSUsageTemplate[1:],
			IntegralShowFSUsageData(ctx))
	}
	b, err := efs.ReadFile(path[len(path)-1])
	if err == nil {
		_, err = ctxparm.Writer.In(ctx).Write(b)
	}
	return err
}

const IntegralStandbyUsageTemplate = `
usage: {{.}} [<pids>]
Wait until interrupt or termination signal.`

var IntegralStandbyUsageData = IntegralShowCompletionUsageData

// Use this to hold container until interrupt or termination signal.
func IntegralStandby(ctx context.Context, args ...string) error {
	if *complete.Help {
		return nil
	}
	if *usage.Help {
		return usage.Error(IntegralStandbyUsageTemplate[1:],
			IntegralStandbyUsageData(ctx))
	}
	<-ctx.Done()
	return ctx.Err()
}

const IntegralStartUsageTemplate = `
usage: {{.Path}} <daemon> [<options>]
Fork self to run <daemon> like this,

   {{.Prog}} daemon <daemon> [<options>]

Daemons
{{.Daemons}}`

func IntegralStartUsageData(ctx context.Context) any {
	return struct{ Path, Prog, Daemons string }{
		Path:    strings.Join(ctxparm.Strings.In(ctx), " "),
		Prog:    program.Base(),
		Daemons: ctxparm.MapKeysIn(ctx),
	}
}

func IntegralStart(ctx context.Context, args ...string) error {
	daemons := ctxparm.Map.In(ctx)["daemon"].(map[string]any)
	ctx = ctxparm.Map.With(ctx, daemons)
	if *complete.Help {
		if len(args) == 0 {
			return complete.Last(args, daemons)
		}
		return do(ctx, Select, args...)
	}
	if *usage.Help {
		if len(args) == 0 {
			return usage.Error(IntegralStartUsageTemplate[1:],
				IntegralStartUsageData(ctx))
		}
		return do(ctx, Select, args...)
	}
	if os.Getpid() == 1 {
		path := ctxparm.Strings.In(ctx)
		ctx = ctxparm.Strings.With(ctx, append(path[:1], "daemon"))
		return do(ctx, Select, args...)
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
		path := ctxparm.Strings.In(ctx)
		w := ctxparm.Writer.In(ctx)
		pid := cmd.Process.Pid
		fmt.Fprint(w, path[0], ":daemon:", args[0], ":pid: ", pid, "\n")
	}
	return err
}
