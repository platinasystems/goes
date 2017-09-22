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
	logAfn   = "dmesg-a.txt"
	logBfn   = "dmesg-b.txt"
	logEfn   = "dmesg-err.txt"
	activeFn = "active.txt"
	nl       = "\n"
	sp       = " "
	lt       = "<"
	gt       = ">"
	lb       = "["
	rb       = "]"
)

func initLogging(c *Info) error {
	exists, err := detectMMC()
	if err != nil {
		return err
	}
	if !exists {
		err = fmt.Errorf("No MMC Card, log disabled: %s", err)
		return err
	}

	if err = mountMMC(c); err != nil {
		return err
	}

	if err = setActive(c); err != nil {
		return err
	}

	if err = startTicker(); err != nil {
		return err
	}
	status(c)
	return nil
}

func updateLogs(c *Info) error {
	if err = initFileInfo(c); err != nil {
		return err
	}
	if err := nextBatch(); err != nil {
		return err
	}
	if err := createAppend(); err != nil {
		return err
	}
	return nil
}

func detectMMC() (bool, error) {
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

func mountMMC(c *Info) error { //TODO
	c.active++
	return nil
}

func initFileInfo(c *Info) error {
	c.logA.Name = MMCdir + "/" + logAfn
	c.logA.Exst = false
	if fi, err := os.Stat(c.logA.Name); !os.IsNotExist(err) {
		c.logA.Exst = false
		c.logA.Size = fi.Size()
	}
	c.logB.Name = MMCdir + "/" + logBfn
	c.logB.Exst = false
	if fi, err := os.Stat(c.logB.Name); !os.IsNotExist(err) {
		c.logB.Exst = false
		c.logB.Size = fi.Size()
	}
	c.logE.Name = MMCdir + "/" + logEfn
	c.logE.Exst = false
	if fi, err := os.Stat(c.logE.Name); !os.IsNotExist(err) {
		c.logE.Exst = false
		c.logE.Size = fi.Size()
	}
	c.actv.Name = MMCdir + "/" + activeFn
	c.actv.Exst = false
	if fi, err := os.Stat(c.actv.Name); !os.IsNotExist(err) {
		c.actv.Exst = false
		c.actv.Size = fi.Size()
	}
	c.logA.SeqN, err = latestSeqNum(c.logA.Name, c.logA.Exst)
	if err != nil {
		return err
	}
	c.logB.SeqN, err = latestSeqNum(c.logB.Name, c.logB.Exst)
	if err != nil {
		return err
	}
	return nil
}

func setActive(c *Info) error {
	if c.actv.Exst {
		dat, err := ioutil.ReadFile(c.actv.Name)
		if err != nil {
			return err
		}
	}
	if !c.logB.Exst {
		return activeA()
	}
	if !c.logA.Exst && c.logB.Exst {
		return activeB()
	}
	x, err := latestSeqNum(c.logA.Name, c.logA.Exst)
	if err != nil {
		return err
	}
	y, err := latestSeqNum(c.logB.Name, c.logB.Exst)
	if err != nil {
		return err
	}
	if x > y {
		return activeA()
	} else {
		return activeB()
	}
	return nil
}

func activeA(c *Info) error {
	c.active = "A"
	d := []byte(c.active)
	if err := ioutil.WriteFile(c.actv, d, 0644); err != nil {
		return err
	}
	return nil
}

func activeB(c *Info) error {
	c.active = "B"
	d := []byte(c.active)
	if err := ioutil.WriteFile(c.actv, d, 0644); err != nil {
		return err
	}
	return nil
}

func latestSeqNum(fn string, ex bool) (x int64, err error) { //TODO
	return 0, nil
}

func startTicker() error {
	f, err := os.Create("/tmp/mmclog_enable")
	if err != nil {
		return nil
	}
	f.Close()
	return nil
}

func stopTicker() error {
	if err := rmFile("/tmp/mmclog_enable"); err != nil {
		return err
	}
	return nil
}

func nextBatch() error { //TODO
	//if just booted grab entire dmesg
	//if 2nd time or more, only grab the new parts, maintain a last timestamp in RAM
	return nil
}

func createAppend() error { //TODO
	//check if buff will fit, if so write it
	//else, erase non-active file if exists, make it active, write there
	return nil
}

func logDmesg(n int) error {
	if n < 1 || n > 100000 {
		return fmt.Errorf("value must be between 1 - 100,000")
	}
	for i := 0; i < n; i++ {
		log.Print("MMC card test message 100k: 123456789012345678 ", i)
	}
	return nil
}

func listMMC() error {
	files, _ := ioutil.ReadDir(MMCdir)
	for _, f := range files {
		fmt.Println(f.Name())
	}
	return nil
}

func getDmesgInfo() error { //FIXME FIGURE THIS OUT
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

func rmFile(f string) error {
	if _, err := os.Stat(f); err != nil {
		return err
	}
	if err := os.Remove(f); err != nil {
		return err
	}
	return nil
}
