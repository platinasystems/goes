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
	"time"

	"github.com/platinasystems/log"
)

const (
	nl = "\n"
	sp = " "
	lt = "<"
	gt = ">"
	lb = "["
	rb = "]"

	MaxEpollEvents = 32
	KB             = 1024 * 64
)

func initLogging(c *Info) error {
	c.logA = MMCDIR + "/" + LOGA
	c.logB = MMCDIR + "/" + LOGB
	c.seq_end = 0

	exists, err := detectMMC()
	if err != nil {
		return err
	}
	if !exists {
		err = fmt.Errorf("No MMC Card, log disabled: %s", err)
		return err
	}
	if _, err := os.Stat(MMCDIR); os.IsNotExist(err) {
		err := os.Mkdir(MMCDIR, 0755)
		if err != nil {
			return fmt.Errorf("mkdir %s: %s", MMCDIR, err)
		}
	}
	err = syscall.Mount("/dev/mmcblk0p1", "/mnt", "ext4", uintptr(0), "")
	if err != nil {
		return err
	}
	if err = startTicker(); err != nil {
		return err
	}
	return nil
}

func updateLogs(c *Info) (err error) {
	err, msg := getNewDmesg(c)
	if err != nil {
		return err
	}
	if err := createAppend(c, msg); err != nil {
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

func getNewDmesg(c *Info) (error, []string) {
	var event syscall.EpollEvent
	var events [MaxEpollEvents]syscall.EpollEvent
	var buf [KB]byte
	var kmsg log.Kmsg
	var si syscall.Sysinfo_t
	msg := make([]string, MAXMSG)
	defer func() { msg = msg[:0] }()

	f, err := os.Open("/dev/kmsg")
	if err != nil {
		return err, nil
	}
	defer f.Close()

	fd := int(f.Fd())
	if err = syscall.SetNonblock(fd, true); err != nil {
		return err, nil
	}
	epfd, err := syscall.EpollCreate1(0)
	if err != nil {
		return err, nil
	}
	defer syscall.Close(epfd)

	event.Events = syscall.EPOLLIN
	event.Fd = int32(fd)
	err = syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, fd, &event)
	if err != nil {
		return err, nil
	}
	nevents, err := syscall.EpollWait(epfd, events[:], -1)
	if err != nil {
		return err, nil
	}
	for ev := 0; ev < nevents; ev++ {
		i := 0
		for {
			nbytes, err := syscall.Read(int(events[ev].Fd), buf[:])
			if nbytes > 0 {
				kmsg.Parse(buf[:nbytes])
				ksq := strconv.Itoa(int(kmsg.Seq))

				now := time.Now()
				tim := time.Time(kmsg.Stamp.Time(now, int64(si.Uptime)))
				kst := fmt.Sprintln(tim)
				ksu := strings.Split(kst, ".")

				kmg := fmt.Sprint(sp, lb, kmsg.Stamp, rb, sp)
				if uint64(kmsg.Seq) > c.seq_end {
					fs := ksq + sp + lb + ksu[0] + rb + kmg + kmsg.Msg
					msg[i] = fmt.Sprintln(fs)
					i++
					c.seq_end = uint64(kmsg.Seq)
				}
			}
			if err != nil {
				break
			}
		}
	}
	return nil, msg
}

func createAppend(c *Info, msg []string) error {
	mode := os.O_CREATE | os.O_APPEND | os.O_WRONLY
	f, err := os.OpenFile(c.logA, mode, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	for i, _ := range msg {
		_, err = f.Write([]byte(msg[i]))
		if err != nil {
			return err
		}
	}
	f.Sync()
	f.Close()
	fi, err := os.Stat(c.logA)
	if err != nil {
		return err
	}
	if fi.Size() > MAXSIZE {
		rmFile(c.logB)
		err := os.Rename(c.logA, c.logB)
		if err != nil {
			return err
		}
	}
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

func listMMC(c *Info) error {
	files, _ := ioutil.ReadDir(MMCDIR)
	for _, f := range files {
		fmt.Println(f.Name())
	}
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
