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
	"strconv"
	"strings"
	"syscall"

	"github.com/platinasystems/go/internal/kexec"
	"github.com/platinasystems/go/internal/url"
)

func getFile(s string, v string, t bool, fn string) (int, error) {
	rmFile(fn)
	urls := "http://" + s + "/" + v + "/" + fn
	if t {
		urls = "tftp://" + s + "/" + v + "/" + fn
	}
	r, err := url.Open(urls)
	if err != nil {
		return 0, nil
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

func getBootedQSPI() (int, error) {
	dat, err := ioutil.ReadFile("/tmp/qspi")
	if err != nil {
		return -1, err
	}
	if strings.Contains(string(dat), "QSPI0") {
		return 0, nil
	}
	if strings.Contains(string(dat), "QSPI1") {
		return 1, nil
	}
	return -1, nil
}

func getQSPIversion(q bool) (string, error) {
	selectQSPI1(q)
	_, b, err := readFlash(Qfmt["ver"].off, Qfmt["ver"].siz)
	if err != nil {
		return "", err
	}
	return string(b[VERSION_OFF:VERSION_LEN]), nil
}

func getInstalledVersions() ([]string, error) {
	iv := make([]string, 2)
	selectQSPI1(false)
	_, b, err := readFlash(Qfmt["ver"].off, Qfmt["ver"].siz)
	if err != nil {
		return nil, err
	}
	iv[0] = string(b[VERSION_OFF:VERSION_LEN])

	selectQSPI1(true)
	_, b, err = readFlash(Qfmt["ver"].off, Qfmt["ver"].siz)
	if err != nil {
		return nil, err
	}
	iv[1] = string(b[VERSION_OFF:VERSION_LEN])
	return iv, nil
}

func getServerVersion(s string, v string, t bool) (string, error) {
	fn := "LIST"
	n, err := getFile(s, v, t, fn)
	if err != nil {
		return "", err
	}
	if n == 0 {
		return "", nil
	}
	l, err := ioutil.ReadFile(fn)
	if err != nil {
		return "", err
	}
	sv, err := findLatestVer(l)
	return sv, nil
}

func printVersions(s string, v string, iv []string, sv string, qspi int) {
	fmt.Print("\n")
	fmt.Print("Installed versions in QSPI flash:\n")
	if qspi == 0 {
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
	if len(sv) > 0 {
		fmt.Print("\n")
		fmt.Print("Version on server:\n")
		fmt.Printf("    Requested server  : %s\n", s)
		fmt.Printf("    Requested version : %s\n", v)
		fmt.Printf("    Found version     : %s\n", sv)
	}
	fmt.Print("\n")
}

func findLatestVer(b []byte) (string, error) {
	s := ""
	latestNum := 0.01
	latestVer := "0.01"
	for _, j := range b {
		if j != 10 {
			s += string(j)
		} else {
			s = strings.TrimSpace(s)
			ss := strings.Replace(s, "v", "", -1)
			f, _ := strconv.ParseFloat(ss, 64)
			if f > latestNum {
				latestNum = f
				latestVer = s
			}
			s = ""
		}
	}
	if latestVer == "" {
		err := fmt.Errorf("findLatestVer error")
		return "", err
	}
	return latestVer, nil
}

func isVersionNewer(cur string, x string) (n bool, err error) {
	var c float64 = 0.0
	var f float64 = 0.0
	cur = strings.TrimSpace(cur)
	cur = strings.Replace(cur, "v", "", -1)
	c, err = strconv.ParseFloat(cur, 64)
	if err != nil {
		c = 0.0
	}
	x = strings.TrimSpace(x)
	x = strings.Replace(x, "v", "", -1)
	f, err = strconv.ParseFloat(x, 64)
	if err != nil {
		f = 0.0
	}
	if f > c {
		return true, nil
	}
	return false, nil
}
