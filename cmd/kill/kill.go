// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package kill

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/platinasystems/goes/external/flags"
	"github.com/platinasystems/goes/lang"
)

var sigByOptName = map[string]syscall.Signal{
	"-abrt":   syscall.SIGABRT,
	"-alrm":   syscall.SIGALRM,
	"-bus":    syscall.SIGBUS,
	"-chld":   syscall.SIGCHLD,
	"-cld":    syscall.SIGCLD,
	"-cont":   syscall.SIGCONT,
	"-fpe":    syscall.SIGFPE,
	"-hup":    syscall.SIGHUP,
	"-ill":    syscall.SIGILL,
	"-int":    syscall.SIGINT,
	"-io":     syscall.SIGIO,
	"-iot":    syscall.SIGIOT,
	"-kill":   syscall.SIGKILL,
	"-pipe":   syscall.SIGPIPE,
	"-poll":   syscall.SIGPOLL,
	"-prof":   syscall.SIGPROF,
	"-pwr":    syscall.SIGPWR,
	"-quit":   syscall.SIGQUIT,
	"-segv":   syscall.SIGSEGV,
	"-stkflt": syscall.SIGSTKFLT,
	"-stop":   syscall.SIGSTOP,
	"-sys":    syscall.SIGSYS,
	"-term":   syscall.SIGTERM,
	"-trap":   syscall.SIGTRAP,
	"-tstp":   syscall.SIGTSTP,
	"-ttin":   syscall.SIGTTIN,
	"-ttou":   syscall.SIGTTOU,
	"-unused": syscall.SIGUNUSED,
	"-urg":    syscall.SIGURG,
	"-usr1":   syscall.SIGUSR1,
	"-usr2":   syscall.SIGUSR2,
	"-vtalrm": syscall.SIGVTALRM,
	"-winch":  syscall.SIGWINCH,
	"-xcpu":   syscall.SIGXCPU,
	"-xfsz":   syscall.SIGXFSZ,
}

type Command struct{}

func (Command) String() string { return "kill" }

func (Command) Usage() string { return "kill [OPTION] [PID]..." }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "signal a process",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	The default signal for kill is 'SIGTERM'.  Use -l to list available
	signals.  Particularly useful signals include '-hup', '-int', '-kill',
	'-stop', '-cont', and '-0'.  Signals may be specified by name or
	number, e.g: '-9' or or '-kill'.  Negative PID values may be used to
	choose whole process groups; see the PGID column in ps  command
	output.  A PID of -1 is special; it indicates all processes except the
	kill process itself and init.

OPTIONS
	<PID> [...]
		Send signal to every <PID> listed.

       -<NAME>
       -<NUMBER>
		Specify the signal to be sent.

	-l	List signal names.

EXAMPLES
	kill -9 -1
		Kill all processes you can kill.

	kill 123 543 2341 3453
		Send the default signal, SIGTERM, to all those processes.`,
	}
}

func (Command) Main(args ...string) error {
	flag, args := flags.New(args, "-l")

	sigByOptNumb := make(map[string]syscall.Signal)
	for _, sig := range []syscall.Signal{
		syscall.SIGABRT,
		syscall.SIGALRM,
		syscall.SIGBUS,
		syscall.SIGCHLD,
		syscall.SIGCLD,
		syscall.SIGCONT,
		syscall.SIGFPE,
		syscall.SIGHUP,
		syscall.SIGILL,
		syscall.SIGINT,
		syscall.SIGIO,
		syscall.SIGIOT,
		syscall.SIGKILL,
		syscall.SIGPIPE,
		syscall.SIGPOLL,
		syscall.SIGPROF,
		syscall.SIGPWR,
		syscall.SIGQUIT,
		syscall.SIGSEGV,
		syscall.SIGSTKFLT,
		syscall.SIGSTOP,
		syscall.SIGSYS,
		syscall.SIGTERM,
		syscall.SIGTRAP,
		syscall.SIGTSTP,
		syscall.SIGTTIN,
		syscall.SIGTTOU,
		syscall.SIGUNUSED,
		syscall.SIGURG,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
		syscall.SIGVTALRM,
		syscall.SIGWINCH,
		syscall.SIGXCPU,
		syscall.SIGXFSZ,
	} {
		sigByOptNumb[fmt.Sprintf("-%d", sig)] = sig
	}

	if flag.ByName["-l"] {
		if len(args) > 0 {
			return fmt.Errorf("%v: unexpected", args)
		}
		fmt.Print(`
 1) hup		 2) int		 3) quit	 4) ill		 5) trap
 6) abrt	 7) bus		 8) fpe		 9) kill	10) usr1
11) segv	12) usr2	13) pipe	14) alrm	15) term
16) stkflt	17) chld	18) cont	19) stop	20) tstp
21) ttin	22) ttou	23) urg		24) xcpu	25) xfsz
`[1:])
		return nil
	}
	if len(args) == 0 {
		return fmt.Errorf("PID: missing")
	}
	sig := syscall.SIGTERM
	if strings.HasPrefix(args[0], "-") {
		opt := args[0]
		args = args[1:]
		if len(args) == 0 {
			return fmt.Errorf("PID: missing")
		}
		if t, found := sigByOptName[opt]; found {
			sig = t
		} else if t, found := sigByOptNumb[opt]; found {
			sig = t
		} else {
			return fmt.Errorf("%s: unknown", opt)
		}
	}
	for _, arg := range args {
		pid, err := strconv.ParseInt(arg, 0, 0)
		if err != nil {
			return err
		}
		proc, err := os.FindProcess(int(pid))
		if err != nil {
			return err
		}
		err = proc.Signal(sig)
		if err != nil {
			return err
		}
	}
	return nil
}
