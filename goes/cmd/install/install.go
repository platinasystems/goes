// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package install provides the named command that installs this executable to
// /usr/bin/NAME; creates /etc/init.d/NAME and /etc/default/goes; then a bunch
// of other stuff. If the executable is a self extracting archive, this unzips
// the appended contents.
package install

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/assert"
	"github.com/platinasystems/go/internal/prog"
)

type Command struct {
	// Machines may use this Hook to complete its installation.
	Hook func() error
}

func (*Command) String() string { return "install" }

func (*Command) Usage() string {
	return "install [START, STOP and REDISD options]..."
}

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "install this goes machine",
	}
}

func (c *Command) Main(args ...string) error {
	err := assert.Root()
	if err != nil {
		return err
	}
	if err = stop(args...); err != nil {
		return err
	}
	if err = install_self(); err != nil {
		return err
	}
	if err = install_default(); err != nil {
		return err
	}
	if err = install_init(); err != nil {
		return err
	}
	if err = update_rc(); err != nil {
		return err
	}
	if err = bash_completion(); err != nil {
		return err
	}
	if err = unzip(); err != nil {
		return err
	}
	if c.Hook != nil {
		if err = c.Hook(); err != nil {
			return err
		}
	}
	return start(args...)

}

func run(args ...string) error {
	cmd := exec.Command(prog.Install, args...)
	cmd.Stdin = nil
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = "/"
	return cmd.Run()
}

// run "goes stop"
func stop(args ...string) error {
	_, err := os.Stat(prog.Install)
	if err != nil {
		return nil
	}
	stop := []string{"stop"}
	if len(args) > 0 {
		args = append(stop, args...)
	} else {
		args = stop
	}
	err = run(args...)
	if err != nil {
		err = fmt.Errorf("stop: %v", err)
	}
	return err
}

func start(args ...string) error {
	start := []string{"start"}
	if len(args) > 0 {
		args = append(start, args...)
	} else {
		args = start
	}
	err := run(args...)
	if err != nil {
		err = fmt.Errorf("start: %v", err)
	}
	return err
}

func install_self() error {
	self, err := os.Readlink("/proc/self/exe")
	if err != nil {
		return err
	}
	r, err := os.Open(self)
	if err != nil {
		return err
	}
	defer r.Close()
	return new_bin(r)
}

func install_init() error {
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	f, err := os.OpenFile("/etc/init.d/goes", flags, os.FileMode(0755))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(`
#!/bin/sh
### BEGIN INIT INFO
# Provides:          goes
# Required-Start:    $local_fs $network $remote_fs $syslog
# Required-Stop:     $local_fs $network $remote_fs $syslog
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: GO-Embedded-System
# Description:       a busybox like app collection
### END INIT INFO

# Author: Tom Grennan <tgrennan@platinasystems.com>

# Do NOT "set -e"

# PATH should only include /usr/* if it runs after the mountnfs.sh script
PATH=/sbin:/usr/sbin:/bin:/usr/bin
DESC="GO-Embedded-System"
NAME=goes
DAEMON=/usr/bin/${NAME}
SCRIPTNAME=/etc/init.d/$NAME

# Exit if the package is not installed
[ -x "$DAEMON" ] || exit 0

[ -r /etc/default/goes ] && . /etc/default/goes

# Load the VERBOSE setting and other rcS variables
. /lib/init/vars.sh

# Define LSB log_* functions.
# Depend on lsb-base (>= 3.2-14) to ensure that this file is present
# and status_of_proc is working.
. /lib/lsb/init-functions

args=""
case "$1" in
  start)
	cmd="start"
	args="$ARGS"
	msg="Starting"
	;;
  stop)
	cmd="stop"
	args="$ARGS"
	msg="Stopping"
	;;
  restart)
	cmd="restart"
	args="$ARGS"
	msg="Restarting"
	;;
  force-reload)
	cmd="reload"
	msg="Reloading"
	;;
  status)
	if [ -S /run/goes/socks/redisd ] ; then
		log_success_msg "$NAME is running"
		exit 0
	fi
	log_failure_msg "$NAME is not running"
	exit 1
	;;
  *)
	echo "Usage: $SCRIPTNAME {start|stop|status|restart|force-reload}" >&2
	exit 3
	;;
esac

[ "$VERBOSE" != no ] && log_daemon_msg "${msg} $DESC" "$NAME"
$DAEMON $cmd $args
ecode="$?"
[ "$VERBOSE" != no ] && log_end_msg $ecode

:
`[1:])
	return err
}

// Install /etc/default/goes if and only if not already present.
func install_default() error {
	const fn = "/etc/default/goes"
	_, err := os.Stat(fn)
	if err == nil {
		return nil
	}
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	f, err := os.OpenFile(fn, flags, os.FileMode(0644))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(`
# goes start arguments

# ARGS: [-start URL] [-stop URL] [REDISD OPTIONS]...
# URL:	goes command scripts that are sourced after starting or before stopping
#	the embedded daemons; defaults: /etc/goes/{start,stop}
#
# See also, $(goes man redisd)

#ARGS=""
`[1:])
	return err
}

func update_rc() error {
	_, err := os.Stat("/usr/sbin/update-rc.d")
	if err != nil {
		// no update-rc.d, may not be debian
		return nil
	}
	err = exec.Command("/usr/sbin/update-rc.d", "goes", "defaults").Run()
	if err != nil {
		err = fmt.Errorf("update-rc.d: %v", err)
	}
	return err
}

func bash_completion() error {
	const fn = "/usr/share/bash-completion/completions/goes"
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	f, err := os.OpenFile(fn, flags, os.FileMode(0644))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(`
_goes ()
{
	if [ -z ${COMP_WORDS[COMP_CWORD]} ] ; then
		COMPREPLY=($(goes complete ${COMP_WORDS[@]:1} ''))
	else
		COMPREPLY=($(goes complete ${COMP_WORDS[@]:1}))
	fi
	return 0
}

type -p goes >/dev/null && complete -F _goes -o filenames goes
`[1:])
	return err
}

func unzip() error {
	progname := prog.Name()
	if z, err := zip.OpenReader(progname); err == nil {
		defer z.Close()
		for _, f := range z.File {
			r, err := f.Open()
			if err != nil {
				return fmt.Errorf("%s: %s: %v",
					progname, f.Name, err)
			}
			if strings.HasPrefix(f.Name, "goes-") {
				err = new_bin(r)
			} else if strings.HasSuffix(f.Name, ".ko") {
				err = new_mod(f.Name, r)
			} else {
				err = new_lib(f.Name, r)
			}
			r.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func new_bin(r io.Reader) error {
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	w, err := os.OpenFile(prog.Install, flags, os.FileMode(0755))
	if err == nil {
		defer w.Close()
		_, err = io.Copy(w, r)
	}
	return err
}

func new_lib(fn string, r io.Reader) error {
	const libdir = "/usr/lib/goes"
	err := os.MkdirAll(libdir, os.FileMode(0755))
	if err != nil {
		return err
	}
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	mode := os.FileMode(0755)
	if strings.HasSuffix(fn, ".so") {
		mode = os.FileMode(0644)
	}
	w, err := os.OpenFile(filepath.Join(libdir, fn), flags, mode)
	if err == nil {
		defer w.Close()
		_, err = io.Copy(w, r)
	}
	return err
}

func new_mod(fn string, r io.Reader) error {
	return fmt.Errorf("FIXME install %s; then depmod", fn)
}
