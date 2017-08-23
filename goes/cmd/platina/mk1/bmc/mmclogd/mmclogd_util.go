// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mmclogd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/platinasystems/go/internal/log"
)

const (
	nl = "\n"
	sp = " "
	lt = "<"
	gt = ">"
	lb = "["
	rb = "]"
)

type logFile struct {
	Name  string
	Size  int64
	Exist bool
	SeqNo int64
}

var MMCdir = "/mnt"
var FileA = "dmesg_log0"
var FileB = "dmesg_log1"
var MaxSize int64 = 512 * 1024 * 1024 //512MiB

var Active = logFile{Name: "", Size: 0, Exist: false}
var Second = logFile{Name: "", Size: 0, Exist: false}

func InitLogging() error {
	exists, err := DetectMMC()
	if err != nil {
		return err
	}
	if !exists {
		err = fmt.Errorf("No MMC Card, log disabled: %s", err)
		return err
	}

	if err = MountMMC(); err != nil {
		return err
	}

	if err = SetActive(); err != nil {
		return err
	}

	if err = StartTicker(); err != nil {
		return err
	}
	return nil
}

func UpdateMMC() error { //TODO
	if err := SetActive(); err != nil {
		return err
	}
	//GRAB LATEST DMESG BLOB
	//CHECK LAST DMESG SEQ IN ACTIVE IF EXISTS
	//WHEN WRITING, IF NOT EXIST, CREATE FIRST
	//WRITE THE LATEST BLOB TO FILE
	return nil
}

func DetectMMC() (bool, error) {
	exists := false
	files, err := ioutil.ReadDir("/dev")
	if err != nil {
		return false, err
	}
	for _, f := range files {
		if !f.IsDir() {
			if strings.Contains(f.Name(), "mmcblk0") {
				exists = true
			}
		}
	}
	return exists, nil
}

func MountMMC() error { //TODO
	return nil
}

func SetActive() error {
	aExist := false
	bExist := false
	aSize := int64(0)
	bSize := int64(0)
	f, err := os.Stat(MMCdir + FileA)
	if !os.IsNotExist(err) {
		aExist = true
		aSize = f.Size()
	}
	f, err = os.Stat(MMCdir + FileB)
	if !os.IsNotExist(err) {
		bExist = true
		bSize = f.Size()
	}
	//CASE 1: one full, one small or doesn't exist => Active=smaller
	if aSize >= MaxSize && bSize < MaxSize {
		Active = logFile{Name: FileB, Size: bSize, Exist: bExist}
		Second = logFile{Name: FileA, Size: aSize, Exist: aExist}
	}
	if bSize >= MaxSize && aSize < MaxSize {
		Active = logFile{Name: FileA, Size: aSize, Exist: aExist}
		Second = logFile{Name: FileB, Size: bSize, Exist: bExist}
	}
	//CASE 2: neither exist => ACTIVE=A
	if !aExist && !bExist {
		Active = logFile{Name: FileA, Size: aSize, Exist: aExist}
		Second = logFile{Name: FileB, Size: bSize, Exist: bExist}
	}
	//CASE 3: both exist, neither full => ACTIVE=NEWER
	if aSize < MaxSize && bSize < MaxSize { //TODO
		Active = logFile{Name: FileA, Size: aSize, Exist: aExist}
		Second = logFile{Name: FileB, Size: bSize, Exist: bExist}
	}
	//CASE 4: A exists being filled, B not exist => ACTIVE=A
	if aExist && !bExist && aSize < MaxSize {
		Active = logFile{Name: FileA, Size: aSize, Exist: aExist}
		Second = logFile{Name: FileB, Size: bSize, Exist: bExist}
	}
	//CASE 5: B exists being filled, A not exist => ACTIVE=B
	if bExist && !aExist && bSize < MaxSize {
		Active = logFile{Name: FileB, Size: bSize, Exist: bExist}
		Second = logFile{Name: FileA, Size: aSize, Exist: aExist}
	}
	//CASE 6: both are full, => ERASE OLDER, ACTIVE=OLDER
	//rmFile(MMCdir + OLDER) TODO

	return nil
}

func Status() error {
	exists, err := DetectMMC()
	if err != nil {
		return err
	}
	if exists == false {
		fmt.Println("MMC card does not exist")
	} else {
		fmt.Println("MMC card exists")
	}

	if _, err := os.Stat("/tmp/mmclog_enable"); os.IsNotExist(err) {
		fmt.Println("Ticker disabled")
	} else {
		fmt.Println("Ticker enabled")
	}

	fmt.Println("Active =", MMCdir+Active.Name, ", Size =", Active.Size, ", Exists =", Active.Exist)
	fmt.Println("Second =", MMCdir+Second.Name, ", Size =", Second.Size, ", Exists =", Second.Exist)
	return nil
}

func LogDmesg(n int) error {
	if n < 1 || n > 100000 {
		return fmt.Errorf("value must be between 1 - 100,000")
	}
	for i := 0; i < n; i++ {
		log.Print("MMC card test message 100k: 123456789012345678 ", i)
	}
	return nil
}

func StartTicker() error {
	f, err := os.Create("/tmp/mmclog_enable")
	if err != nil {
		return nil
	}
	f.Close()
	return nil
}

func StopTicker() error {
	if err := rmFile("/tmp/mmclog_enable"); err != nil {
		return err
	}
	return nil
}

func ListMMC() error {
	files, _ := ioutil.ReadDir(MMCdir)
	for _, f := range files {
		fmt.Println(f.Name())
	}
	return nil
}

func GetDmesgInfo() error {
	var kmsg log.Kmsg

	f, err := os.Open("/dev/kmsg")
	if err != nil {
		return err
	}
	defer f.Close()
	if err = syscall.SetNonblock(int(f.Fd()), true); err != nil {
		return err
	}
	buf := make([]byte, 4096)
	defer func() { buf = buf[:0] }()
	var si syscall.Sysinfo_t
	if err = syscall.Sysinfo(&si); err != nil {
		return err
	}

	fo, err := os.Create("/tmp/dat2")
	if err != nil {
		return err
	}
	defer f.Close()

	//get starting seq number -- function to find out the older file
	//get ending seq number -- to avoid dups
	//only tack onto the end to file (no dups)
	for i := 0; i < 400; i++ {
		n, err := f.Read(buf)
		if err != nil {
			break
		}
		kmsg.Parse(buf[:n])
		ksq := strconv.Itoa(int(kmsg.Seq))
		kst := strconv.Itoa(int(kmsg.Stamp)) //convert this to time
		fs := ksq + sp + lb + kst + rb + sp + kmsg.Msg + nl
		_, err = fo.Write([]byte(fs))
	}

	fo.Sync()

	return nil
}

func ReadFile() error {
	dat, err := ioutil.ReadFile("/tmp/dat2")
	if err != nil {
		return err
	}
	fmt.Print(string(dat))
	fmt.Print(len(dat))
	fmt.Print(dat[0])

	return nil
}

func printFirstTimestamp() error {
	return nil
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
