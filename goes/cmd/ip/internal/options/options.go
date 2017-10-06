// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package options

import (
	"net"
	"path/filepath"
	"sort"
	"strings"

	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/parms"
)

var (
	Family = []interface{}{"-4", "-6", "-B", "-M", "-0"}
	Flags  = []interface{}{
		[]string{"-h", "-human", "-human-readable"},
		[]string{"-s", "-stats", "-statistics"},
		[]string{"-d", "-details"},
		[]string{"-o", "-oneline"},
		[]string{"-r", "-resolve"},
		[]string{"-a", "-all"},
		[]string{"-c", "-color"},
		[]string{"-t", "-timestamp"},
		[]string{"-ts", "-tshort"},
		"-iec",
	}
	Parms = []interface{}{
		[]string{"-l", "-loops"},
		[]string{"-f", "-family"},
		[]string{"-rc", "-rcvbuf"},
	}
	CompleteParmValue = map[string]func(string) []string{
		"-l":      NoComplete,
		"-loops":  NoComplete,
		"-rc":     NoComplete,
		"-rcvbuf": NoComplete,
		"-f":      CompleteFamily,
		"-family": CompleteFamily,
	}
	CompleteOptNames = []string{
		"-human-readable",
		"-statistics",
		"-details",
		"-oneline",
		"-resolve",
		"-all",
		"-color",
		"-timestamp",
		"-tshort",
		"-iec",
		"-family",
		"-loops",
		"-rcvbuf",
	}
)

type Options struct {
	Flags *flags.Flags
	Parms *parms.Parms
}

// Parse common IP options from command arguments.
func New(args []string) (*Options, []string) {
	opt := new(Options)
	opt.Flags, args = flags.New(args, Flags...)
	opt.Parms, args = parms.New(args, Parms...)
	family, args := flags.New(args, Family...)
	switch {
	case family.ByName["-4"]:
		opt.Parms.ByName["-f"] = "inet"
	case family.ByName["-6"]:
		opt.Parms.ByName["-f"] = "inet6"
	case family.ByName["-B"]:
		opt.Parms.ByName["-f"] = "bridge"
	case family.ByName["-M"]:
		opt.Parms.ByName["-f"] = "mpls"
	case family.ByName["-0"]:
		opt.Parms.ByName["-f"] = "link"
	}
	return opt, args
}

func CompleteFamily(s string) (list []string) {
	families := []string{
		"inet",
		"inet6",
		"bridge",
		"mpls",
		"link",
	}
	if len(s) == 0 {
		list = families
	} else {
		for _, family := range families {
			if strings.HasPrefix(family, s) {
				list = append(list, family)
			}
		}
	}
	if len(list) > 0 {
		sort.Strings(list)
	}
	return
}

func CompleteFile(s string) (list []string) {
	fns, err := filepath.Glob(s + ".*")
	if err != nil {
		return []string{}
	}
	for _, fn := range fns {
		if len(s) == 0 || strings.HasPrefix(fn, s) {
			list = append(list, fn)
		}
	}
	return
}

func CompleteIfName(s string) (list []string) {
	itfs, err := net.Interfaces()
	if err != nil {
		return
	}
	for _, itf := range itfs {
		if len(s) == 0 || strings.HasPrefix(itf.Name, s) {
			list = append(list, itf.Name)
		}
	}
	if len(list) > 0 {
		sort.Strings(list)
	}
	return
}

func NoComplete(string) []string { return []string{} }
