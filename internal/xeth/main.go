package xeth

import (
	"fmt"
	"io/ioutil"
	"os"
)

// Run sample XETH controller.
func (xeth *Xeth) Main() {
	const assertDial = true
	args := os.Args[1:]
	usage := fmt.Sprint("usage:\t", xeth,
		` { -dump | -set DEVICE STAT COUNT | FILE | - }...

DEVICE	an interface name or its ifindex
STAT	an 'ip link' or 'ethtool' statistic
FILE,-	receive an exception frame from FILE or STDIN`)

	if len(args) == 0 {
		fmt.Println(usage)
		return
	}

	defer func() {
		r := recover()
		err := xeth.Close()
		if r == nil {
			r = err
		}
		if r != nil {
			fmt.Fprint(os.Stderr, xeth, ": ", r, "\n")
			os.Exit(1)
		}
	}()

	for len(args) > 0 {
		switch args[0] {
		case "help", "-help", "--help":
			fmt.Println(usage)
			return
		case "dump", "-dump", "--dump":
			fmt.Println("FIXME dump recvmsg")
		case "set", "-set", "--set":
			var count uint64
			switch len(args) {
			case 1:
				panic(fmt.Errorf("missing DEVICE\n%s", usage))
			case 2:
				panic(fmt.Errorf("missing STAT\n%s", usage))
			case 3:
				panic(fmt.Errorf("missing COUNT\n%s", usage))
			}
			_, err := fmt.Sscan(args[3], &count)
			if err != nil {
				panic(fmt.Errorf("COUNT %q %v", args[3], err))
			}
			err = xeth.Set(args[1], args[2], count)
			if err != nil {
				panic(err)
			}
			args = args[4:]
		case "-":
			buf, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				panic(err)
			}
			if err = xeth.ExceptionFrame(buf); err != nil {
				panic(err)
			}
			args = args[1:]
		default:
			buf, err := ioutil.ReadFile(args[0])
			if err != nil {
				panic(err)
			}
			if err = xeth.ExceptionFrame(buf); err != nil {
				panic(err)
			}
			args = args[1:]
		}
	}
}
