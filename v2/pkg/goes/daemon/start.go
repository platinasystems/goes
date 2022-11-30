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
	"strings"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/os/xdg"
)

// fork self to run key'd daemon.
func Start(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	daemon := path[len(path)-1]
	if strings.HasPrefix(daemon, "_") {
		if path[1] == "complete" {
			return nil
		}
		s := strings.TrimPrefix(daemon, "_")
		return fmt.Errorf("%s %w", s, ErrUnavailable)
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
	preempted := path[1] == "complete" || path[1] == "help"
	if preempted {
		cmd.Args = append(cmd.Args, path[1])
	}
	cmd.Args = append(cmd.Args, "daemon", daemon)
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
	if preempted {
		cmd.Stdout = w
		cmd.Stderr = w
		err = cmd.Run()
	} else {
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Credential: cred,
			Setsid:     true,
		}
		if err = cmd.Start(); err == nil {
			fmt.Fprint(w, program.Base(), ":daemon:", daemon,
				":pid: ", cmd.Process.Pid, "\n")
		}
	}
	return err
}
