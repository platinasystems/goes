// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Parse /etc/resolv.conf
package xresolv

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/platinasystems/goes/v2/pkg/kvc"
)

var Path = "/etc/resolv.conf"
var DefaultNameserver = "localhost"

// If available, return domain suffix of [os.Hostname].
var HostDomain = func() string {
	if hn, err := os.Hostname(); err == nil {
		if i := strings.Index(hn, "."); i > 0 {
			return hn[i:]
		}
	}
	return ""
}

type Resolv struct {
	Nameserver,
	Search,
	SortList []string
	// Options may be a bool, uint, or string
	Options map[string]any
}

// [Text] with contents of [Path] or nothing if that's unavailable.
func New() *Resolv {
	data, _ := os.ReadFile(Path)
	return Text(data)
}

// Unmarshal string into new [Resolv]
func String(s string) *Resolv {
	r := new(Resolv)
	r.UnmarshalText([]byte(s))
	return r
}

// Unmarshal data into new [Resolv]
func Text(data []byte) *Resolv {
	r := new(Resolv)
	r.UnmarshalText(data)
	return r
}

func (r *Resolv) Format(w fmt.State, verb rune) {
	for _, s := range r.Nameserver {
		fmt.Fprintln(w, "nameserver", s)
	}
	if len(r.Search) > 0 {
		fmt.Fprint(w, "search")
		for _, s := range r.Search {
			fmt.Fprint(w, " ", s)
		}
		fmt.Fprintln(w)
	}
	if len(r.SortList) > 0 {
		fmt.Fprint(w, "sortlist")
		for _, s := range r.SortList {
			fmt.Fprint(w, " ", s)
		}
		fmt.Fprintln(w)
	}
	if len(r.Options) > 0 {
		unsorted := maps.Keys(r.Options)
		sorted := slices.Collect(unsorted)
		slices.Sort(sorted)
		fmt.Fprint(w, "options")
		for _, k := range sorted {
			fmt.Fprint(w, " ", k)
			if o, ok := r.Options[k]; ok {
				if _, ok = o.(bool); !ok {
					fmt.Fprint(w, ":", o)
				}
			}
		}
		fmt.Fprintln(w)
	}
}

// If after unmarshal [Resolv.Nameserver] or [Resolv.Search] are empty, add
// [DefaultNameserver] and non-empty result of [HostDomain] respectively.
func (r *Resolv) UnmarshalText(data []byte) error {
	if r.Options == nil {
		r.Options = make(map[string]any)
	} else {
		clear(r.Options)
	}
	kvc.RangeText(data, r.SplitLine, r.ProcKeyVals)
	if len(r.Nameserver) == 0 {
		r.Nameserver = []string{DefaultNameserver}
	}
	if len(r.Search) == 0 {
		if hd := HostDomain(); len(hd) > 0 {
			r.Search = []string{hd}
		}
	}
	return nil
}

func (r *Resolv) SplitLine(s string) []string {
	s = strings.TrimSpace(s)
	if len(s) == 0 || !unicode.IsLetter([]rune(s)[0]) {
		return nil
	}
	args := strings.Fields(s)
	for i, arg := range args {
		if len(arg) == 0 {
			args = args[:i]
			break
		}
		if r := []rune(arg)[0]; r == '#' || r == ';' {
			args = args[:i]
			break
		}
		if r, sz := utf8.DecodeLastRuneInString(arg); r == ';' {
			args[i] = arg[:len(arg)-sz]
			args = args[:i+1]
			break
		}
	}
	return args
}

func (r *Resolv) ProcKeyVals(lno int, key string, vals []string) error {
	if len(vals) == 0 {
		return nil
	}
	switch key {
	case "nameserver":
		r.Nameserver = append(r.Nameserver, vals[0])
	case "search":
		if r.Search != nil {
			// per spec, only accept last search directive
			r.Search = r.Search[:0]
		}
		r.Search = append(r.Search, vals...)
	case "sortlist":
		r.SortList = append(r.SortList, vals...)
	case "options":
		for _, opt := range vals {
			if i := strings.Index(opt, ":"); i > 1 {
				k, s := opt[:i], opt[i+1:]
				u, err := strconv.ParseUint(s, 0, 32)
				if err == nil {
					r.Options[k] = uint(u)
				} else {
					r.Options[k] = s
				}
			} else {
				r.Options[opt] = true
			}
		}
	}
	return nil
}
