// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package install provides the named command that installs this executable to
// /usr/bin/NAME; creates /etc/init.d/NAME and /etc/default/goes; then a bunch
// of other stuff.
package install

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/platinasystems/go/commands/machine/internal"
)

const Name = "install"
const UsrBinGoes = "/usr/bin/goes"
const EtcInitdGoes = "/etc/init.d/goes"
const EtcDefaultGoes = "/etc/default/goes"

// Machines may use this Hook to complete its installation.
var Hook = func() error { return nil }

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name }

func (cmd) Main(...string) error {
	err := internal.AssertRoot()
	if err != nil {
		return err
	}
	if err = stop(); err != nil {
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
	if err = Hook(); err != nil {
		return err
	}
	return start()

}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "install this goes machine",
	}
}

func stop() error {
	_, err := os.Stat(UsrBinGoes)
	if err != nil {
		return nil
	}
	err = exec.Command(UsrBinGoes, "stop").Run()
	if err != nil {
		err = fmt.Errorf("goes stop: %v", err)
	}
	return err
}

func start() error {
	err := exec.Command(UsrBinGoes, "start").Run()
	if err != nil {
		err = fmt.Errorf("goes start: %v", err)
	}
	return err
}

func install_self() error {
	self, err := os.Readlink("/proc/self/exe")
	if err != nil {
		return err
	}
	src, err := os.Open(self)
	if err != nil {
		return err
	}
	defer src.Close()
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	dst, err := os.OpenFile(UsrBinGoes, flags, os.FileMode(0755))
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}

func install_init() error {
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	f, err := os.OpenFile(EtcInitdGoes, flags, os.FileMode(0755))
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
DAEMON_ARGS=""
SCRIPTNAME=/etc/init.d/$NAME

# Exit if the package is not installed
[ -x "$DAEMON" ] || exit 0

REDISD=
MACHINED=
[ -r /etc/default/goes ] && . /etc/default/goes
export REDISD
export MACHINED

# Load the VERBOSE setting and other rcS variables
. /lib/init/vars.sh

# Define LSB log_* functions.
# Depend on lsb-base (>= 3.2-14) to ensure that this file is present
# and status_of_proc is working.
. /lib/lsb/init-functions

case "$1" in
  start)
	cmd="start"
	msg="Starting"
	;;
  stop)
	cmd="stop"
	msg="Stopping"
	;;
  restart)
	cmd="restart"
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
$DAEMON $cmd
ecode="$?"
[ "$VERBOSE" != no ] && log_end_msg $ecode

:
`[1:])
	return err
}

// Install /etc/default/goes if and only if not already present.
func install_default() error {
	_, err := os.Stat(EtcDefaultGoes)
	if err == nil {
		return nil
	}
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	f, err := os.OpenFile(EtcDefaultGoes, flags, os.FileMode(0644))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(`
# goes configuration options

# REDISD lists the netdevices that the daemon should listen to.
#REDISD="lo eth0"

# MACHINED lists command script URLs that are sourced after daemon start.
#MACHINED=""
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
	_, err := os.Stat(filepath.Dir(fn))
	if err != nil {
		return nil
	}
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	f, err := os.OpenFile(fn, flags, os.FileMode(0644))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(`
_goes ()
{
	COMPREPLY=($(goes -complete ${COMP_WORDS[@]}))
	return 0
}

type -p goes >/dev/null && complete -F _goes goes
`[1:])
	return err
}
