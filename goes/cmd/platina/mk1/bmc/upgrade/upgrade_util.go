// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package upgrade

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/platinasystems/go/internal/kexec"
	"github.com/platinasystems/go/internal/url"
)

func getInstalledVersions() ([]string, error) {
	iv := make([]string, 2)
	_, b, err := readFlash(off[5], siz[5])
	if err != nil {
		return nil, err
	}
	iv[0] = string(b[VERSION_OFF:VERSION_LEN])

	_, b, err = readFlash(altoff, altsiz)
	if err != nil {
		return nil, err
	}
	iv[1] = string(b[VERSION_OFF:VERSION_LEN])
	return iv, nil
}

func getServerVersion(s string, v string, t bool) (string, error) {
	fn := "LIST"
	if _, err := getFile(s, v, t, fn); err != nil {
		return "", err
	}
	_, err := ioutil.ReadFile(fn) //l
	if err != nil {
		return "", err
	}
	//LOOP THROUGH L for LATEST, THEN TAKE PART RIGHT OF -, CHOPPED //FIXME

	return "20170901", nil
}

func bootFromBackup() (bool, error) { //FIXME GET QSPI from DMESG
	return false, nil
}

func printVersions(s string, v string, iv []string, sv string, backup bool) {
	fmt.Print("\n")
	fmt.Print("Installed versions in QSPI flash:\n")
	if !backup {
		fmt.Printf("  * QSPI0 version: %s\n", iv[0])
		fmt.Printf("    QSPI1 version: %s\n", iv[1])
		fmt.Print("\n")
		fmt.Print("Booted from QSPI0\n")
	} else {
		fmt.Printf("    QSPI0 version: %s\n", iv[0])
		fmt.Printf("  * QSPI1 version: %s\n", iv[1])
		fmt.Print("\n")
		fmt.Print("Booted from QSPI1\n")
	}
	fmt.Print("\n")
	fmt.Print("Version on server:\n")
	fmt.Printf("    Requested server  : %s\n", s)
	fmt.Printf("    Requested version : %s\n", v)
	fmt.Printf("    Found version     : %s\n", sv)
	fmt.Print("\n")
}

func getFile(s string, v string, t bool, fn string) (int, error) {
	rmFile(fn)
	urls := "http://" + s + "/" + v + "/" + fn
	if t {
		urls = "tftp://" + s + "/" + v + "/" + fn
	}
	r, err := url.Open(urls)
	if err != nil {
		return 0, err
	}
	f, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, DfltMod)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	n, err := io.Copy(f, r)
	if err != nil {
		return 0, err
	}
	syscall.Fsync(int(os.Stdout.Fd()))
	return int(n), nil
}

func rmFile(f string) error {
	if _, err := os.Stat(f); err != nil {
		return err
	}
	if err := os.Remove(f); err != nil {
		return err
	}
	return nil
}

func reboot() error {
	kexec.Prepare()
	_ = syscall.Reboot(syscall.LINUX_REBOOT_CMD_KEXEC)
	return syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
}

func unzip() error {
	archive := ArchiveName
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	target := "."
	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}
		r, err := file.Open()
		if err != nil {
			return err
		}
		defer r.Close()
		t, err := os.OpenFile(
			path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer t.Close()
		if _, err := io.Copy(t, r); err != nil {
			return err
		}
	}
	return nil
}
