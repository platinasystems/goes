// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package femtocom

import (
	"bytes"
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"github.com/platinasystems/goes/external/flags"
	"github.com/platinasystems/goes/external/parms"
	"github.com/platinasystems/goes/lang"
)

type Command struct{}

func (Command) String() string { return "femtocom" }

func (Command) Usage() string {
	return "femtocom [OPTION]... DEVICE"
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "tiny serial-terminal emulation",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	femtocom copies console input to DEVICE and DEVICE output to the
	console until input of "^A^X".

OPTIONS
	-baud BAUD
		The valid baud rates are 300, 600, 1200, 2400,
		4800, 9600, 19200, 38400, 57600, and 115200.

	-parity PARITY
		The valid parity options are "odd", "even" and default, "none".

	-databits BITS
		The valid bits per character are 5, 6, 7, and default, 8.

	-stopbits BITS
		The valid stop bits per character are 2 and default, 1.

	-noinit
		Don't initialize the device at start-up or reset on exit.

	-noreset
		Don't reset the device on exit.

	-nolock
		Don't attempt exclusive device use.`,
	}
}

func (Command) Main(args ...string) error {
	const (
		ctrlA rune = 1
		ctrlX rune = 'x' - 'a' + 1
	)
	var err error
	flag, args := flags.New(args, "-noinit", "-noreset", "-nolock")
	parm, args := parms.New(args, "-baud", "-parity", "-databits",
		"-stopbits")

	if len(parm.ByName["-baud"]) == 0 {
		parm.ByName["-baud"] = "115200"
	}
	baud, found := map[string]uint32{
		"50":      syscall.B50,
		"75":      syscall.B75,
		"110":     syscall.B110,
		"134":     syscall.B134,
		"150":     syscall.B150,
		"200":     syscall.B200,
		"300":     syscall.B300,
		"600":     syscall.B600,
		"1200":    syscall.B1200,
		"1800":    syscall.B1800,
		"2400":    syscall.B2400,
		"4800":    syscall.B4800,
		"9600":    syscall.B9600,
		"19200":   syscall.B19200,
		"38400":   syscall.B38400,
		"57600":   syscall.B57600,
		"115200":  syscall.B115200,
		"230400":  syscall.B230400,
		"460800":  syscall.B460800,
		"500000":  syscall.B500000,
		"576000":  syscall.B576000,
		"921600":  syscall.B921600,
		"1000000": syscall.B1000000,
		"1152000": syscall.B1152000,
		"1500000": syscall.B1500000,
		"2000000": syscall.B2000000,
		"2500000": syscall.B2500000,
		"3000000": syscall.B3000000,
		"3500000": syscall.B3500000,
		"4000000": syscall.B4000000,
	}[parm.ByName["-baud"]]
	if !found {
		return fmt.Errorf("%s: invalid baud", parm.ByName["-baud"])
	}

	if len(parm.ByName["-parity"]) == 0 {
		parm.ByName["-parity"] = "none"
	}
	parity, found := map[string]uint32{
		"odd":  syscall.PARENB | syscall.PARODD,
		"even": syscall.PARENB,
		"none": 0,
	}[parm.ByName["-parity"]]
	if !found {
		return fmt.Errorf("%s: invalid parity", parm.ByName["-parity"])
	}

	if len(parm.ByName["-databits"]) == 0 {
		parm.ByName["-databits"] = "8"
	}
	databits, found := map[string]uint32{
		"5": syscall.CS5,
		"6": syscall.CS6,
		"7": syscall.CS7,
		"8": syscall.CS8,
	}[parm.ByName["-databits"]]
	if !found {
		return fmt.Errorf("%s: invalid databits", parm.ByName["-databits"])
	}

	if len(parm.ByName["-stopbits"]) == 0 {
		parm.ByName["-stopbits"] = "1"
	}
	stopbits, found := map[string]uint32{
		"1": 0,
		"2": syscall.CSTOPB,
	}[parm.ByName["-stopbits"]]
	if !found {
		return fmt.Errorf("%s: invalid stopbits", parm.ByName["-stopbits"])
	}

	if len(args) == 0 {
		return fmt.Errorf("DEVICE: missing")
	}
	if len(args) > 1 {
		return fmt.Errorf("%v: unexpected", args[1:])
	}

	dev, err := os.OpenFile(args[0], os.O_RDWR, 0664)
	if err != nil {
		return err
	}
	defer dev.Close()

	if !flag.ByName["-nolock"] {
		_, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
			uintptr(dev.Fd()),
			uintptr(syscall.TIOCEXCL),
			uintptr(0))
		if errno != 0 {
			return fmt.Errorf("TIOCEXCL: %v", errno)
		}
		defer syscall.Syscall(syscall.SYS_IOCTL,
			uintptr(dev.Fd()),
			uintptr(syscall.TIOCNXCL),
			uintptr(0))
	}

	var savedStdin, savedDev syscall.Termios

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TCGETS),
		uintptr(unsafe.Pointer(&savedStdin)))
	if errno != 0 {
		return fmt.Errorf("TCGETS: stdin: %v", errno)
	}
	_, _, errno = syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(dev.Fd()),
		uintptr(syscall.TCGETS),
		uintptr(unsafe.Pointer(&savedDev)))
	if errno != 0 {
		return fmt.Errorf("TCGETS: %s: %v", args[0], errno)
	}
	defer func() {
		syscall.Syscall(syscall.SYS_IOCTL,
			uintptr(syscall.Stdin),
			uintptr(syscall.TCSETS),
			uintptr(unsafe.Pointer(&savedStdin)))
		if !flag.ByName["-noinit"] && !flag.ByName["-noreset"] {
			syscall.Syscall(syscall.SYS_IOCTL,
				uintptr(dev.Fd()),
				uintptr(syscall.TCSETS),
				uintptr(unsafe.Pointer(&savedDev)))
		}
	}()

	if !flag.ByName["-noinit"] {
		t := savedDev
		t.Iflag &^= syscall.IGNBRK |
			syscall.BRKINT |
			syscall.PARMRK |
			syscall.ISTRIP |
			syscall.INLCR |
			syscall.IGNCR |
			syscall.ICRNL |
			syscall.IXON
		t.Oflag &^= syscall.OPOST
		t.Lflag &^= syscall.ECHO |
			syscall.ECHONL |
			syscall.ICANON |
			syscall.ISIG |
			syscall.IEXTEN
		t.Cflag = baud | parity | databits | stopbits |
			syscall.HUPCL |
			syscall.CREAD |
			syscall.CLOCAL
		_, _, errno = syscall.Syscall(syscall.SYS_IOCTL,
			uintptr(dev.Fd()),
			uintptr(syscall.TCSETS),
			uintptr(unsafe.Pointer(&t)))
		if errno != 0 {
			return fmt.Errorf("TCSETS: %v", errno)
		}
	}
	{
		t := savedStdin
		t.Iflag &^= syscall.IGNBRK |
			syscall.BRKINT |
			syscall.PARMRK |
			syscall.ISTRIP |
			syscall.INLCR |
			syscall.IGNCR |
			syscall.ICRNL |
			syscall.IXON
		t.Oflag &^= syscall.OPOST
		t.Lflag &^= syscall.ECHO |
			syscall.ECHONL |
			syscall.ICANON |
			syscall.ISIG |
			syscall.IEXTEN
		_, _, errno = syscall.Syscall(syscall.SYS_IOCTL,
			uintptr(syscall.Stdin),
			uintptr(syscall.TCSETS),
			uintptr(unsafe.Pointer(&t)))
		if errno != 0 {
			return fmt.Errorf("TCSETS: %v", errno)
		}
	}

	fmt.Print("Type ^A^X to exit.\r\n")

	go func() {
		// Copy from dev to Stdout
		rx := make([]byte, 4096)
		for {
			n, err := dev.Read(rx)
			if err != nil {
				break
			}
			os.Stdout.Write(rx[:n])
		}
		rx = rx[:0]
	}()
	tx := make([]byte, 4096)
	escaped := false
	for {
		// Copy from Stdin to DEVICE up until C-a C-x
		n, err := os.Stdin.Read(tx)
		if err != nil {
			break
		}
		txs := tx
		if !escaped {
			if i := bytes.IndexRune(txs, ctrlA); i >= 0 {
				escaped = true
				if i > 0 {
					dev.Write(txs[:i])
					txs = txs[i+1:]
					n -= i + 1
				} else {
					txs = txs[1:]
					n -= 1
				}
			}
		}
		if n <= 0 {
			continue
		}
		if escaped {
			if r := rune(txs[0]); r == ctrlX {
				break
			}
			escaped = false
		}
		dev.Write(txs[:n])
	}
	tx = tx[:0]
	return nil
}
