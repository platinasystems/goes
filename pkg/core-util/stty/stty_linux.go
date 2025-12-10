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
		{"stop", syscall.VSTOP},
		{"susp", syscall.VSUSP},
		{"swtc", syscall.VSWTC},
		{"time", syscall.VTIME},
		{"werase", syscall.VWERASE},
	}
	namedCFlags = []namedFlag{
		{"hupcl", unix.HUPCL},
		{"cstopb", unix.CSTOPB},
		{"cread", unix.CREAD},
		{"clocal", unix.CLOCAL},
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
		{"iuclc", unix.IUCLC},
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
		{"xcase", unix.XCASE},
		{"tostop", unix.TOSTOP},
		{"echoctl", unix.ECHOCTL},
		{"echoprt", unix.ECHOPRT},
		{"echoke", unix.ECHOKE},
		{"flusho", unix.FLUSHO},
		{"pendin", unix.PENDIN},
	}
	namedOFlags = []namedFlag{
		{"opost", unix.OPOST},
		{"olcuc", unix.OLCUC},
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
	b := termios.Cflag & unix.CBAUD
	if b&unix.CBAUDEX == unix.CBAUDEX {
		return []uint64{
			0,
			57600,
			115200,
			230400,
			460800,
			500000,
			576000,
			921600,
			1000000,
			1152000,
			1500000,
			2000000,
			2500000,
			3000000,
			3500000,
			4000000,
		}[b&0xf]
	}
	return []uint64{
		0,
		50,
		75,
		110,
		134,
		150,
		200,
		300,
		600,
		1200,
		1800,
		2400,
		4800,
		9600,
		19200,
		38400,
	}[b&0xf]
}

func get() (err error) {
	p := uintptr(unsafe.Pointer(&termios))
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, tty.Fd(), unix.TCGETS, p)
	if errno != 0 {
		err = errno
	}
	return
}

func getCc() []uint8   { return termios.Cc[:] }
func getCflag() uint64 { return uint64(termios.Cflag) }
func getCSize() string {
	return []string{
		unix.CS5: "cs5",
		unix.CS6: "cs6",
		unix.CS7: "cs7",
		unix.CS8: "cs8",
	}[termios.Cflag&unix.CSIZE]
}
func getIflag() uint64 { return uint64(termios.Iflag) }
func getLflag() uint64 { return uint64(termios.Lflag) }
func getOflag() uint64 { return uint64(termios.Oflag) }

func printRestoration() {
	fmt.Println(xerrors.FIXME("restoration"))
}

func set(args []string) error {
	return xerrors.FIXME("set")
}
