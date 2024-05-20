// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package complete

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FlagSetter interface{ VisitAll(fn func(*flag.Flag)) }

// Print the sorted matches of args[len(args)-1] completed by:
//
//	FlagSetter
//	string - glob files
//	[]string - glossary
//	map[string]any - matching keys
func Last(args []string, completers ...any) error {
	var last string
	var matches []string
	if len(args) > 0 {
		last = args[len(args)-1]
	}
	for _, v := range completers {
		switch t := v.(type) {
		case FlagSetter:
			matches = append(matches, Flags(t, last)...)
		case string:
			matches = append(matches, Glob(t, last)...)
		case []string:
			matches = append(matches, Glossary(t, last)...)
		case map[string]string:
			matches = append(matches, MapKeys(t, last)...)
		case map[string]any:
			matches = append(matches, MapKeys(t, last)...)
		default:
			return fmt.Errorf("can't complete %T", t)
		}
	}
	sort.Strings(matches)
	for _, match := range matches {
		fmt.Println(match)
	}
	return nil
}

func Flags(fs FlagSetter, arg string) (c []string) {
	if !strings.HasPrefix(arg, "-") {
		return
	}
	arg = strings.TrimLeft(arg, "-")
	fs.VisitAll(func(f *flag.Flag) {
		if len(arg) == 0 || strings.HasPrefix(f.Name, arg) {
			c = append(c, fmt.Sprint("-", f.Name))
		}
	})
	return
}

// Pattern matching files.
func Glob(pat, arg string) (c []string) {
	const ps = string(os.PathSeparator)
	const tildedir = "~" + ps
	var dn, bn string

	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	athome := strings.HasPrefix(arg, tildedir)
	if athome {
		// FIXME this isn't working for zsh
		if arg == tildedir {
			arg = home
		} else {
			arg = strings.Replace(arg, "~", home, 1)
		}
	}

	if fi, err := os.Stat(arg); err == nil && fi.IsDir() {
		dn, bn = arg, "."
	} else {
		dn, bn = filepath.Dir(arg), filepath.Base(arg)
	}

	dir, err := os.ReadDir(dn)
	if err != nil {
		return
	}
	for _, de := range dir {
		if bn == "." || strings.HasPrefix(de.Name(), bn) {
			match, err := filepath.Match(pat, de.Name())
			if de.IsDir() || (err == nil && match) {
				c = append(c, filepath.Join(dn, de.Name()))
			}
		}
	}
	if athome {
		for i, s := range c {
			if strings.HasPrefix(s, home) {
				c[i] = strings.Replace(s, home, "~", 1)
			} else if strings.HasPrefix(s, "."+ps) {
				c[i] = s[2:]
			}
		}
	}
	return
}

func Glossary(l []string, arg string) (c []string) {
	if len(arg) == 0 {
		for _, s := range l {
			if len(s) > 0 {
				c = append(c, s)
			}
		}
	} else {
		for _, s := range l {
			if len(s) > 0 && strings.HasPrefix(s, arg) {
				c = append(c, s)
			}
		}
	}
	return
}

// Arg prefixed map keys or, if empty arg, all keys w/o '_' prefix.
func MapKeys[V any](m map[string]V, arg string) (c []string) {
	if len(arg) == 0 {
		for k := range m {
			if len(k) > 0 && !strings.HasPrefix(k, "_") {
				c = append(c, k)
			}
		}
	} else {
		for k := range m {
			if strings.HasPrefix(k, arg) {
				c = append(c, k)
			}
		}
	}
	return
}
