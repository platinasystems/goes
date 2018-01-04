// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package stty

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"github.com/platinasystems/go/goes/lang"
)

type Command struct{}

func (Command) String() string { return "stty" }

func (Command) Usage() string { return "stty [DEVICE]" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "print info for given or current TTY",
	}
}

func (Command) Main(args ...string) error {
	dev := os.Stdin
	if len(args) > 0 {
		var err error
		name := args[0]
		if len(args) > 1 {
			return fmt.Errorf("%v: unexpected", args[1:])
		}
		dev, err = os.Open(name)
		if err != nil {
			return err
		}
		defer dev.Close()
	}

	t := &syscall.Termios{}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(dev.Fd()),
		uintptr(syscall.TCGETS),
		uintptr(unsafe.Pointer(t)))
	if errno != 0 {
		return fmt.Errorf("TCGETS: %v", errno)
	}

	name := dev.Name()
	for {
		link, err := os.Readlink(name)
		if err != nil {
			break
		}
		name = link
	}
	fmt.Println("name ", name)

	fmt.Print("speed ", []string{
		syscall.B0:       "0",
		syscall.B50:      "50",
		syscall.B75:      "75",
		syscall.B110:     "110",
		syscall.B134:     "134",
		syscall.B150:     "150",
		syscall.B200:     "200",
		syscall.B300:     "300",
		syscall.B600:     "600",
		syscall.B1200:    "1200",
		syscall.B1800:    "1800",
		syscall.B2400:    "2400",
		syscall.B4800:    "4800",
		syscall.B9600:    "9600",
		syscall.B19200:   "19200",
		syscall.B38400:   "38400",
		syscall.B57600:   "57600",
		syscall.B115200:  "115200",
		syscall.B230400:  "230400",
		syscall.B460800:  "460800",
		syscall.B500000:  "500000",
		syscall.B576000:  "576000",
		syscall.B921600:  "921600",
		syscall.B1000000: "1000000",
		syscall.B1152000: "1152000",
		syscall.B1500000: "1500000",
		syscall.B2000000: "2000000",
		syscall.B2500000: "2500000",
		syscall.B3000000: "3000000",
		syscall.B3500000: "3500000",
		syscall.B4000000: "4000000",
	}[t.Cflag&^0x0ff0], " baud")
	fmt.Print("; rows ", os.Getenv("ROWS"))
	fmt.Print("; columns ", os.Getenv("COLUMNS"))
	fmt.Print("; line ", t.Line)
	fmt.Print(";\n")

	for i, x := range []struct {
		name  string
		index uint8
	}{
		{"intr", syscall.VINTR},
		{"quit", syscall.VQUIT},
		{"erase", syscall.VERASE},
		{"kill", syscall.VKILL},
		{"eof", syscall.VEOF},
		{"time", syscall.VTIME},
		{"min", syscall.VMIN},
		{"swtc", syscall.VSWTC},
		{"start", syscall.VSTART},
		{"stop", syscall.VSTOP},
		{"susp", syscall.VSUSP},
		{"eol", syscall.VEOL},
		{"reprint", syscall.VREPRINT},
		{"discard", syscall.VDISCARD},
		{"werase", syscall.VWERASE},
		{"lnext", syscall.VLNEXT},
		{"eol2", syscall.VEOL2},
	} {
		if r := rune(t.Cc[x.index]); r != 0 {
			if i > 0 {
				fmt.Print("; ")
			}
			fmt.Print(x.name, " = ")
			if r < ' ' {
				fmt.Printf("^%c", r+'A'-1)
			} else if r == 127 {
				fmt.Print("^?")
			} else {
				fmt.Print(r)
			}
		}
	}
	fmt.Print(";\n")

	for i, x := range []struct {
		name string
		bit  uint32
	}{
		{"parenb", syscall.PARENB},
		{"parodd", syscall.PARODD},
	} {
		if i > 0 {
			fmt.Print(" ")
		}
		if t.Cflag&x.bit != x.bit {
			fmt.Print("-")
		}
		fmt.Print(x.name)
	}
	fmt.Print(" -cmspar")
	fmt.Print(" ", []string{
		syscall.CS5: "cs5",
		syscall.CS6: "cs6",
		syscall.CS7: "cs7",
		syscall.CS8: "cs8",
	}[t.Cflag&syscall.CSIZE])
	for _, x := range []struct {
		name string
		bit  uint32
	}{
		{"hupcl", syscall.HUPCL},
		{"cstopb", syscall.CSTOPB},
		{"cread", syscall.CREAD},
		{"clocal", syscall.CLOCAL},
	} {
		fmt.Print(" ")
		if t.Cflag&x.bit != x.bit {
			fmt.Print("-")
		}
		fmt.Print(x.name)
	}
	fmt.Println(" -crtscts")

	for i, x := range []struct {
		name string
		bit  uint32
	}{
		{"ignbrk", syscall.IGNBRK},
		{"brkint", syscall.BRKINT},
		{"ignpar", syscall.IGNPAR},
		{"parmrk", syscall.PARMRK},
		{"inpck", syscall.INPCK},
		{"istrip", syscall.ISTRIP},
		{"inlcr", syscall.INLCR},
		{"igncr", syscall.IGNCR},
		{"icrnl", syscall.ICRNL},
		{"ixon", syscall.IXON},
		{"ixoff", syscall.IXOFF},
		{"iuclc", syscall.IUCLC},
		{"ixany", syscall.IXANY},
		{"imaxbel", syscall.IMAXBEL},
		{"iutf8", syscall.IUTF8},
	} {
		if i > 0 {
			fmt.Print(" ")
		}
		if t.Iflag&x.bit != x.bit {
			fmt.Print("-")
		}
		fmt.Print(x.name)
	}
	fmt.Println()

	for i, x := range []struct {
		name string
		bit  uint32
	}{
		{"opost", syscall.OPOST},
		{"olcuc", syscall.OLCUC},
		{"ocrnl", syscall.OCRNL},
		{"onlcr", syscall.ONLCR},
		{"onocr", syscall.ONOCR},
		{"onlret", syscall.ONLRET},
		{"ofill", syscall.OFILL},
		{"ofdel", syscall.OFDEL},
	} {
		if i > 0 {
			fmt.Print(" ")
		}
		if t.Oflag&x.bit != x.bit {
			fmt.Print("-")
		}
		fmt.Print(x.name)
	}
	fmt.Println()

	for i, x := range []struct {
		name string
		bit  uint32
	}{
		{"isig", syscall.ISIG},
		{"icanon", syscall.ICANON},
		{"iexten", syscall.IEXTEN},
		{"echo", syscall.ECHO},
		{"echoe", syscall.ECHOE},
		{"echok", syscall.ECHOK},
		{"echonl", syscall.ECHONL},
		{"noflsh", syscall.NOFLSH},
		{"xcase", syscall.XCASE},
		{"tostop", syscall.TOSTOP},
		{"echoctl", syscall.ECHOCTL},
		{"echoprt", syscall.ECHOPRT},
		{"echoke", syscall.ECHOKE},
		{"flusho", syscall.FLUSHO},
		{"pendin", syscall.PENDIN},
	} {
		if i > 0 {
			fmt.Print(" ")
		}
		if t.Lflag&x.bit != x.bit {
			fmt.Print("-")
		}
		fmt.Print(x.name)
	}
	fmt.Println()
	return nil
}
