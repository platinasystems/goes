// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
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

func Files(arg string) (c []string) {
	dn := filepath.Dir(arg)
	bn := filepath.Base(arg)
	if bn == "." || bn == "/" {
		bn = ""
	}
	rdn := dn
	if len(rdn) == 0 {
		rdn = "."
	} else if strings.HasPrefix(dn, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return
		}
		rdn = strings.Replace(dn, "~", home, 1)
	}
	dir, err := os.ReadDir(rdn)
	if err != nil {
		return
	}
	for _, de := range dir {
		if s := de.Name(); len(bn) == 0 || strings.HasPrefix(s, bn) {
			c = append(c, filepath.Join(dn, s))
		}
	}
	return
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

func Glossary(l []string, arg string) (c []string) {
	for _, s := range l {
		if len(arg) == 0 || strings.HasPrefix(s, arg) {
			c = append(c, s)
		}
	}
	return
}

// Matching map keys or, if empty, all keys.
func MapKeys[V any](m map[string]V, arg string) (c []string) {
	for k := range m {
		if len(arg) == 0 || strings.HasPrefix(k, arg) {
			c = append(c, k)
		}
	}
	return
}

func Print(c []string) {
	sort.Strings(c)
	for _, s := range c {
		fmt.Println(s)
	}
}

func PrintFiles(arg string)                          { Print(Files(arg)) }
func PrintFlags(fs FlagSetter, arg string)           { Print(Flags(fs, arg)) }
func PrintGlossary(l []string, arg string)           { Print(Glossary(l, arg)) }
func PrintMapKeys[V any](m map[string]V, arg string) { Print(MapKeys(m, arg)) }
func PrintPrograms(arg string)                       { Print(Programs(arg)) }

func Programs(arg string) (c []string) {
	if strings.Contains(arg, "/") {
		c = Files(arg)
		return
	}
	path := os.Getenv("PATH")
	if len(path) == 0 {
		return
	}
	for _, dn := range filepath.SplitList(path) {
		dir, err := os.ReadDir(dn)
		if err != nil {
			continue
		}
		for _, de := range dir {
			fn := de.Name()
			if de.IsDir() {
				continue
			}
			if !strings.HasPrefix(fn, arg) {
				continue
			}
			pn := filepath.Join(dn, fn)
			if err = Access(pn, X_OK); err == nil {
				c = append(c, fn)
			}
		}
	}
	return
}

type FlagSetter interface{ VisitAll(fn func(*flag.Flag)) }
