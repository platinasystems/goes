// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/platinasystems/go/elib/elog"

	"bytes"
	"fmt"
	"os"
	"time"
)

// Event logging.
type event struct {
	i uint32
}

func (e *event) String() string          { return fmt.Sprintf("event #%d", e.i) }
func (e *event) Encode(b []byte) int     { return elog.EncodeUint32(b, e.i) }
func (e *event) Decode(b []byte) (i int) { e.i, i = elog.DecodeUint32(b, i); return }

//go:generate gentemplate -d Package=main -id event -d Type=event github.com/platinasystems/go/elib/elog/event.tmpl

func main() {
	elog.Enable(true)
	for i := uint32(0); i < 10; i++ {
		e := event{i: i}
		e.Log()
	}
	var b bytes.Buffer

	elog.Print(os.Stdout)

	v := elog.NewView()

	err := v.Save(&b)
	if err != nil {
		panic(err)
	}

	if nb, ne := b.Len(), elog.Len(); ne > 0 {
		fmt.Printf("%d events, %d bytes, %.4f bytes/event\n", ne, nb, float64(nb)/float64(ne))
	}

	err = v.Restore(&b)
	if err != nil {
		panic(err)
	}

	v.Print(os.Stdout)

	go elog.PrintOnHangupSignal(os.Stdout)

	for {
		time.Sleep(1 * time.Second)
	}
}
