// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package upgrade

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/platinasystems/go/internal/kexec"
	"github.com/platinasystems/go/internal/url"
)

func upgradeX(s string, v string, t bool, x string, f bool) error {
	fmt.Printf("Update %s\n", x)
	if !f {
		vr := getVer(x)
		vs, err := getSrvVer(s, v, t, x)
		if err != nil {
			return err
		}
		fmt.Printf("    %s version currently:  %s\n", x, vr)
		fmt.Printf("    %s version on server:  %s\n", x, vs)
		if vr == vs {
			fmt.Print("    Versions match, skipping %s upgrade\n\n", x)
			return nil
		}
		if len(vr) == 0 || len(vs) == 0 {
			fmt.Print("    No tag found, aborting %s upgrade\n\n", x)
			return nil
		}
	}
	if err := writeImageX(x); err != nil {
		return err
	}
	return nil
}

func getVer(x string) string { //FIXME
	switch x {
	case "ubo":
		return ""
	case "dtb":
		return ""
	case "env":
		return ""
	case "ker":
		return ""
	case "ini":
		return ""
	}
	return ""
}

func getSrvVer(s string, v string, t bool, x string) (string, error) { //FIXME
	switch x {
	case "ubo":
		return "", nil
	case "dtb":
		return "", nil
	case "env":
		return "", nil
	case "ker":
		return "", nil
	case "ini":
		return "", nil
	}
	return "", nil
}

func prVer(v []string, vs []string) {
	fmt.Print("\n")
	fmt.Print("Currently running:\n")
	fmt.Printf("    U-boot version     : %s\n", v[0])
	fmt.Printf("    Device tree version: %s\n", v[1])
	fmt.Printf("    Environment version: %s\n", v[2])
	fmt.Printf("    Kernel version     : %s\n", v[3])
	fmt.Printf("    Initrd/Goes version: %s\n", v[4])
	fmt.Print("\n")
	fmt.Print("Version on server:\n")
	fmt.Printf("    U-boot version     : %s\n", vs[0])
	fmt.Printf("    Device tree version: %s\n", vs[1])
	fmt.Printf("    Environment version: %s\n", vs[2])
	fmt.Printf("    Kernel version     : %s\n", vs[3])
	fmt.Printf("    Initrd/Goes version: %s\n", vs[4])
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
