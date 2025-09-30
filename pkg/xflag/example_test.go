// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xflag_test

import (
	"flag"
	"fmt"
	"net/netip"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xflag"
)

var exampleArgs = []string{
	"-a", "a",
	"-b",
	"-f", "3s",
	"-h", "42",
	"-i", "fc00:1234::2/64",
	"-j", "1.23",
}

var (
	aVar string

	bVar, cVar bool

	dVar netip.Addr
)

var pkgFlags = xflag.Labels{
	{"a", "set pkg string", &aVar},
	xflag.Label{"b", "set pkg enable", func() error {
		bVar = true
		return nil
	}},
	{"c", "unset pkg enable", func() error {
		cVar = true
		return nil
	}},
	{"d", "unset pkg netip.Addr", func() any {
		addr, err := netip.ParseAddr("192.168.1.1")
		if err != nil {
			return err
		}
		dVar = addr
		return &dVar
	}},
}

func ExampleLabels() {
	var (
		eVar = "e"
		fVar = 10 * time.Second
		gVar = -1
		hVar int64
		iVar netip.Prefix
		jVar float64
	)
	fs := flag.NewFlagSet("labels", flag.ContinueOnError)
	err := append(pkgFlags, xflag.Labels{
		{"e", "unset func string", &eVar},
		{"f", "set func duration", &fVar},
		{"g", "unset func int", &gVar},
		{"h", "set func int64", &hVar},
		{"i", "set func netip.Prefix", &iVar},
		{"j", "set func float64", &jVar},
	}...).DefineIn(fs)
	if err != nil {
		fmt.Println(err)
	} else if err = fs.Parse(exampleArgs); err != nil {
		fmt.Println(err)
	} else {
		fmt.Print(
			"a: ", aVar, "\n",
			"b: ", bVar, "\n",
			"c: ", cVar, "\n",
			"d: ", dVar, "\n",
			"e: ", eVar, "\n",
			"f: ", fVar, "\n",
			"g: ", gVar, "\n",
			"h: ", hVar, "\n",
			"i: ", iVar, "\n",
			"j: ", jVar, "\n",
		)
	}
	// Output:
	// a: a
	// b: true
	// c: false
	// d: 192.168.1.1
	// e: e
	// f: 3s
	// g: -1
	// h: 42
	// i: fc00:1234::2/64
	// j: 1.23
}
