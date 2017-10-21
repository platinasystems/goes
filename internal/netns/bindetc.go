// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netns

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
)

func BindEtc(name string) error {
	dn := filepath.Join("/etc/netns", name)
	_, err := os.Stat(dn)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	dir, err := ioutil.ReadDir(dn)
	if err != nil {
		return err
	}
	for _, fi := range dir {
		fn := fi.Name()
		if fn == "." || fn == ".." {
			continue
		}
		dst := filepath.Join("/etc", fn)
		src := filepath.Join("/etc/netns", name, fn)
		err := syscall.Mount(src, dst, "none", syscall.MS_BIND, "")
		if err != nil {
			return err
		}
	}
	return nil
}
