/* Copyright(c) 2018 Platina Systems, Inc.
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
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/platinasystems/xeth/platina/mk1"
)

var machine = flag.String("test.machine", "platina-mk1",
	"reference platform's ethtool flag and stat names")

func TestMain(m *testing.M) {
	flag.Parse()
	switch *machine {
	case "platina-mk1":
		EthtoolPrivFlagNames = mk1.EthtoolFlags
		EthtoolStatNames = mk1.EthtoolStats
	default:
		fmt.Fprintf(os.Stderr, "machine %q unknown\n", *machine)
		os.Exit(1)
	}
	if err := Start("xeth-test"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer Stop()
	os.Exit(m.Run())
}

func TestShowInterfaces(t *testing.T) {
	Interface.Iterate(func(entry *InterfaceEntry) error {
		fmt.Println(entry)
		return nil
	})
}
