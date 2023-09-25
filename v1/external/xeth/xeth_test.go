// Copyright © 2018-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xeth

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/internal/assert"
)

var Continue = flag.Bool("test.continue", false,
	"continue after ifinfo and fib dumps unil SIGINT")

func Test(t *testing.T) {
	var dev string
	for _, dev = range []string{"platina-mk1", "xeth", ""} {
		if len(dev) == 0 {
			t.Skip("no xeth driver")
		}
		if _, err := net.InterfaceByName(dev); err == nil {
			break
		}
	}
	if err := Provision(dev, " "); err != nil && !os.IsNotExist(err) {
		t.Fatal("provision:", err)
	}
	if err := assert.Root(); err != nil {
		t.Skip(err)
	}

	goes.Stop = make(chan struct{})

	defer goes.WG.Wait()
	defer close(goes.Stop)

	flag.Parse()

	w := ioutil.Discard
	if testing.Verbose() {
		w = os.Stdout
	}

	task, err := Start(dev)
	if err != nil {
		t.Fatal(err)
	}

	task.DumpIfInfo()
	for buf := range task.RxCh {
		if Class(buf) == ClassBreak {
			break
		}
		// Load the attribute cache through Parse
		if msg := Parse(buf); msg != nil {
			Pool(msg)
		} else {
			t.Fatal("Parsed buf is nil msg")
		}
	}
	if task.RxErr != nil {
		t.Fatal(task.RxErr)
	}

	LinkRange(func(xid Xid, l *Link) bool {
		fmt.Fprint(w, l.IfInfoName(), ", xid ")
		if xid < VlanNVid {
			fmt.Fprint(w, uint32(xid))
		} else {
			fmt.Fprint(w, "(", uint32(xid/VlanNVid), ", ",
				uint32(xid&VlanVidMask), ")")
		}
		fmt.Fprint(w, ", ifindex ", l.IfInfoIfIndex(),
			", netns ", l.IfInfoNetNs(),
			", kind ", l.IfInfoDevKind())
		if ipnets := l.IPNets(); len(ipnets) > 0 {
			fmt.Fprint(w, ", ipnets ", ipnets)
		}
		if uppers := l.Uppers(); len(uppers) > 0 {
			fmt.Fprint(w, ", uppers ", uppers)
		}
		if lowers := l.Lowers(); len(lowers) > 0 {
			fmt.Fprint(w, ", lowers ", lowers)
		}
		fmt.Fprintln(w)
		return true
	})

	task.DumpFib()
	for buf := range task.RxCh {
		if Class(buf) == ClassBreak {
			break
		}
		msg := Parse(buf)
		fmt.Fprintln(w, msg)
		Pool(msg)
	}
	if task.RxErr != nil {
		t.Fatal(task.RxErr)
	}

	if *Continue {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch,
			syscall.SIGTERM,
			syscall.SIGINT,
			syscall.SIGHUP,
			syscall.SIGQUIT)
		goes.WG.Add(1)
		go func() {
			defer goes.WG.Done()
			fmt.Fprintln(w, "continue...")
			for buf := range task.RxCh {
				msg := Parse(buf)
				fmt.Fprintln(w, msg)
				Pool(msg)
			}
			if task.RxErr != nil {
				t.Error(task.RxErr)
			}
		}()
		<-sigch
		signal.Stop(sigch)
		fmt.Fprintln(w, "stopped")
	}
}
