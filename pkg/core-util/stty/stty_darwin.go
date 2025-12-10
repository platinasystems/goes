// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package stty

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"golang.org/x/sys/unix"
)

const Settings = `
FIXME
`

var (
	namedVchars = []namedVchar{
		{"discard", syscall.VDISCARD},
		{"dsup", syscall.VDSUSP},
		{"eof", syscall.VEOF},
		{"eol", syscall.VEOL},
		{"eol2", syscall.VEOL2},
		{"erase", syscall.VERASE},
		{"intr", syscall.VINTR},
		{"kill", syscall.VKILL},
		{"lnext", syscall.VLNEXT},
		{"min", syscall.VMIN},
		{"quit", syscall.VQUIT},
		{"reprint", syscall.VREPRINT},
		{"start", syscall.VSTART},
		{"status", syscall.VSTATUS},
		{"stop", syscall.VSTOP},
		{"susp", syscall.VSUSP},
		{"time", syscall.VTIME},
		{"werase", syscall.VWERASE},
	}
	namedCFlags = []namedFlag{
		{"clocal", unix.CLOCAL},
		{"cread", unix.CREAD},
		{"cstopb", unix.CSTOPB},
		{"hupcl", unix.HUPCL},
		{"parenb", unix.PARENB},
		{"parodd", unix.PARODD},
	}
	namedIFlags = []namedFlag{
		{"ignbrk", unix.IGNBRK},
		{"brkint", unix.BRKINT},
		{"ignpar", unix.IGNPAR},
		{"parmrk", unix.PARMRK},
		{"inpck", unix.INPCK},
		{"istrip", unix.ISTRIP},
		{"inlcr", unix.INLCR},
		{"igncr", unix.IGNCR},
		{"icrnl", unix.ICRNL},
		{"ixon", unix.IXON},
		{"ixoff", unix.IXOFF},
		{"ixany", unix.IXANY},
		{"imaxbel", unix.IMAXBEL},
		{"iutf8", unix.IUTF8},
	}
	namedLFlags = []namedFlag{
		{"isig", unix.ISIG},
		{"icanon", unix.ICANON},
		{"iexten", unix.IEXTEN},
		{"echo", unix.ECHO},
		{"echoe", unix.ECHOE},
		{"echok", unix.ECHOK},
		{"echonl", unix.ECHONL},
		{"noflsh", unix.NOFLSH},
		{"tostop", unix.TOSTOP},
		{"echoctl", unix.ECHOCTL},
		{"echoprt", unix.ECHOPRT},
		{"echoke", unix.ECHOKE},
		{"flusho", unix.FLUSHO},
		{"pendin", unix.PENDIN},
	}
	namedOFlags = []namedFlag{
		{"opost", unix.OPOST},
		{"ocrnl", unix.OCRNL},
		{"onlcr", unix.ONLCR},
		{"onocr", unix.ONOCR},
		{"onlret", unix.ONLRET},
		{"ofill", unix.OFILL},
		{"ofdel", unix.OFDEL},
	}
)

var termios unix.Termios

func baud() uint64 {
	for _, entry := range []struct {
		baud,
		code uint64
	}{
		{0, unix.B0},
		{50, unix.B50},
		{75, unix.B75},
		{110, unix.B110},
		{134, unix.B134},
		{150, unix.B150},
		{200, unix.B200},
		{300, unix.B300},
		{600, unix.B600},
		{1200, unix.B1200},
		{1800, unix.B1800},
		{2400, unix.B2400},
		{4800, unix.B4800},
		{7200, unix.B7200},
		{9600, unix.B9600},
		{14400, unix.B14400},
		{19200, unix.B19200},
		{28800, unix.B28800},
		{38400, unix.B38400},
		{57600, unix.B57600},
		{76800, unix.B76800},
		{115200, unix.B115200},
		{230400, unix.B230400},
	} {
		if entry.code == termios.Ispeed || entry.code == termios.Ospeed {
			return entry.baud
		}
	}
	return 0
}

func get() (err error) {
	p := uintptr(unsafe.Pointer(&termios))
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, tty.Fd(), unix.TIOCGETA, p)
	if errno != 0 {
		err = errno
	}
	return
}

func getCc() []uint8   { return termios.Cc[:] }
func getCflag() uint64 { return termios.Cflag }
func getCSize() string {
	return []string{
		unix.CS5: "cs5",
		unix.CS6: "cs6",
		unix.CS7: "cs7",
		unix.CS8: "cs8",
	}[termios.Cflag&unix.CSIZE]
}
func getIflag() uint64 { return termios.Iflag }
func getLflag() uint64 { return termios.Lflag }
func getOflag() uint64 { return termios.Oflag }

func printRestoration() {
	fmt.Println(xerrors.FIXME("restoration"))
}

func set(args []string) error {
	return xerrors.FIXME("set")
}
