// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package daemon

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/goes/selection"
	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/os/xdg"
)

func Start(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path selection.Path,
	args ...string,
) error {
	if path.HasComplete() {
		return nil
	}
	if path.HasHelp() || selection.HasHelp(args) {
		path.Usage(w, "<daemon> [<args>]\n",
			"Start named daemon.",
		)
		return nil
	}
	if len(args) == 0 {
		return selection.ErrIncomplete
	}
	u, err := user.Current()
	if err != nil {
		return err
	}
	cred := &syscall.Credential{NoSetGroups: true}
	if _, err := fmt.Sscan(u.Uid, &cred.Uid); err != nil {
		return fmt.Errorf("user:uid: %w", err)
	}
	if _, err := fmt.Sscan(u.Gid, &cred.Gid); err != nil {
		return fmt.Errorf("user:gid: %w", err)
	}
	cmd := exec.Command(program.Executable())
	flag.VisitAll(func(f *flag.Flag) {
		if s := f.Value.String(); s != f.DefValue {
			cmd.Args = append(cmd.Args, fmt.Sprint(
				"-", f.Name, "=", s))
		}
	})
	cmd.Args = append(cmd.Args, "daemon")
	cmd.Args = append(cmd.Args, args...)
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
	cmd.Dir = xdg.RunTimeDir()
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: cred,
		Setsid:     true,
	}
	if err = cmd.Start(); err == nil {
		fmt.Fprint(w, program.Base(), "_", args[0], "_pid=",
			cmd.Process.Pid, "\n")
	}
	return err
}
