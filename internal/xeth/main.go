/* A sample XETH controller.
 *
 * Copyright(c) 2018 Platina Systems, Inc.
 *
 * This program is free software; you can redistribute it and/or modify it
 * under the terms and conditions of the GNU General Public License,
 * version 2, as published by the Free Software Foundation.
 *
 * This program is distributed in the hope it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
 * more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * this program; if not, write to the Free Software Foundation, Inc.,
 * 51 Franklin St - Fifth Floor, Boston, MA 02110-1301 USA.
 *
 * The full GNU General Public License is included in this distribution in
 * the file called "COPYING".
 *
 * Contact Information:
 * sw@platina.com
 * Platina Systems, 3180 Del La Cruz Blvd, Santa Clara, CA 95054
 */
package xeth

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"unsafe"
)

func Main() {
	name := filepath.Base(os.Args[0])
	args := os.Args[1:]
	usage := fmt.Sprint("usage:\t", name, " ", `
	{ dump DB | set DEVICE STAT COUNT | FILE | - }...

DB	{ ethtool | fdb }
DEVICE	an interface name or its ifindex
STAT	an 'ip link' or 'ethtool' statistic
FILE,-	receive an exception frame from FILE or STDIN`[2:])
	xeth, err := New(name, DialOpt(false))
	defer func() {
		r := recover()
		if err := xeth.Close(); r == nil {
			r = err
		}
		if r != nil {
			fmt.Fprint(os.Stderr, name, ": ", r, "\n")
			os.Exit(1)
		}
	}()
	if err != nil {
		panic(err)
	}
	if len(args) == 0 {
		fmt.Println(usage)
		return
	}
	for len(args) > 0 {
		switch args[0] {
		case "help", "-help", "--help", "-h":
			fmt.Println(usage)
			return
		case "dump", "-dump", "--dump":
			if len(args) < 2 {
				panic(fmt.Errorf("missing DB\n%s", usage))
			}
			xeth.Assert()
			switch args[1] {
			case "ethtool":
				xeth.EthtoolDump()
			case "fdb":
				panic("FIXME")
			default:
				panic(fmt.Errorf("%s: uknown DB\n%s", args[1],
					usage))
			}
			if err := xeth.UntilBreak(dump); err != nil {
				panic(err)
			}
			args = args[2:]
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
			xeth.Assert()
			err = xeth.SetStat(args[1], args[2], count)
			if err != nil {
				panic(err)
			}
			args = args[4:]
		case "-":
			buf, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				panic(err)
			}
			xeth.Assert()
			if err = xeth.ExceptionFrame(buf); err != nil {
				panic(err)
			}
			args = args[1:]
		default:
			buf, err := ioutil.ReadFile(args[0])
			if err != nil {
				panic(err)
			}
			xeth.Assert()
			if err = xeth.ExceptionFrame(buf); err != nil {
				panic(err)
			}
			args = args[1:]
		}
	}
}

func dump(buf []byte) error {
	var stringer fmt.Stringer
	ptr := unsafe.Pointer(&buf[0])
	hdr := (*Hdr)(ptr)
	if !hdr.IsHdr() {
		return fmt.Errorf("invalid xeth msg: %#x", buf)
	}
	switch Op(hdr.Op) {
	case XETH_LINK_STAT_OP, XETH_ETHTOOL_STAT_OP:
		stringer = (*StatMsg)(ptr)
	case XETH_ETHTOOL_FLAGS_OP:
		stringer = (*EthtoolFlagsMsg)(ptr)
	case XETH_ETHTOOL_SETTINGS_OP:
		stringer = (*EthtoolSettingsMsg)(ptr)
	default:
		return fmt.Errorf("invalid op: %d", hdr.Op)
	}
	fmt.Println(stringer)
	return nil
}
