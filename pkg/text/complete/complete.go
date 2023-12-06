// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package complete

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var Help = flag.Bool("complete", false, "Finish last arg.")

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

// Files matching last arg + pattern.
func Glob(pat, arg string) (c []string) {
	ps := string(os.PathSeparator)
	c, _ = filepath.Glob(fmt.Sprint(arg, pat))
	for i, fn := range c {
		if fi, err := os.Stat(fn); err == nil {
			if fi.IsDir() {
				c[i] += ps
			}
		}
	}
	if len(c) == 1 && strings.HasSuffix(c[0], ps) {
		if fis, err := ioutil.ReadDir(c[0]); err == nil {
			for _, fi := range fis {
				name := filepath.Join(c[0], fi.Name())
				realname, err := filepath.EvalSymlinks(name)
				if err == nil {
					realfi, err := os.Stat(realname)
					if err == nil {
						fi = realfi
					}
				}
				if fi.IsDir() {
					name += ps
				}
				c = append(c, name)
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
