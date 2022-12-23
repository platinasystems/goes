// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package complete

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Print the sorted results of args[len(args)-1] completed by:
//
//	string - glob files
//	*flag.Flagset
//	[]net.Interface
//	[]string
func Last(w io.Writer, args []string, completers ...any) {
	var last string
	if len(args) > 0 {
		last = args[len(args)-1]
	}
	var c []string
	for _, v := range completers {
		switch t := v.(type) {
		case string:
			c = append(c, Glob(t, last)...)
		case *flag.FlagSet:
			c = append(c, Flags(t, last)...)
		case []net.Interface:
			c = append(c, NetDevs(t, last)...)
		case []string:
			c = append(c, Glossary(t, last)...)
		}
	}
	sort.Strings(c)
	for _, s := range c {
		fmt.Fprintln(w, s)
	}
}

func Glob(pat, arg string) (c []string) {
	ps := string(os.PathSeparator)
	c, _ = filepath.Glob(fmt.Sprint(arg, pat))
	for i, fn := range c {
		if fi, err := os.Stat(fn); err == nil {
			if fi.IsDir() {
				c[i] = fmt.Sprint(fn, ps)
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

func Flags(fs *flag.FlagSet, arg string) (c []string) {
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

func NetDevs(itfs []net.Interface, arg string) (c []string) {
	for _, itf := range itfs {
		if len(arg) == 0 || strings.HasPrefix(itf.Name, arg) {
			c = append(c, itf.Name)
		}
	}
	return
}

func Glossary(l []string, arg string) (c []string) {
	for _, s := range l {
		if len(s) == 0 {
			continue
		}
		if len(arg) == 0 || strings.HasPrefix(s, arg) {
			c = append(c, s)
		}
	}
	return
}
