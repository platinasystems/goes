// Copyright © 2015-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func LastArg(args []string) (s string) {
	if len(args) > 0 {
		s = args[len(args)-1]
	}
	return
}

func CompleteFiles(args []string) (c []string) {
	ps := string(os.PathSeparator)
	c, _ = filepath.Glob(fmt.Sprint(LastArg(args), "*"))
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

func CompleteFlags(fs *flag.FlagSet, args []string) (c []string) {
	arg := strings.TrimLeft(LastArg(args), "-")
	fs.VisitAll(func(f *flag.Flag) {
		if len(arg) == 0 || strings.HasPrefix(f.Name, arg) {
			c = append(c, fmt.Sprint("-", f.Name))
		}
	})
	return
}

func CompleteInterfaces(args []string) (c []string) {
	arg := LastArg(args)
	itfs, err := net.Interfaces()
	if err != nil {
		return
	}
	for _, itf := range itfs {
		if len(arg) == 0 || strings.HasPrefix(itf.Name, arg) {
			c = append(c, itf.Name)
		}
	}
	return
}

func CompleteStrings(l []string, args []string) (c []string) {
	arg := LastArg(args)
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
