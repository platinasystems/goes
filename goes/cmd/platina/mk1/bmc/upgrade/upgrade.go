// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//TODO SERVER -s, HASH -v, LIST -l with parm,flag
//TODO AUTH WGET
//TODO BLOB / DEBLOB
//TODO upgrade to be called automatically if enabled, register with bootsrvr
//TODO generate ubo.bin, env.bin

package upgrade

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/cavaliercoder/grab"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/kexec"
	"github.com/platinasystems/go/internal/parms"
	"github.com/platinasystems/go/internal/url"
)

const (
	Name    = "upgrade"
	Apropos = "upgrade images"
	Usage   = "upgrade [LATEST | -v HASH] [-l] [-s SERVER]"
	Man     = `
DESCRIPTION
	The upgrade command upgrades firmware images in flash.

	The upgrade command can upgrade any or all of the images
	in flash, depending what is in the image blob.  Images blobs
	can be downloaded either from the Platina webserver on the
	internet, or, a local server in the enterprise.
	Image versions can be either LATEST, or a particular past
	version can be specified with a version hash.

	For the BMC the independently erasable and replacable images are:
	   1. ubo:  QSPI header, u-boot bootloader
	   2. dtb:  device tree file
	   3. env:  u-boot envvar block
	   4. ker:  linux kernel
	   5. ini:  initrd  filesystem containing goes

OPTIONS
	LATEST		upgrades flash to platina-mk1-bmc-LATEST
	-v [HASH]	upgrades flash to platina-mk1-bmc-[HASH]
	-l		lists available upgrade hashes
	-s [SERVER]     specifies SERVER, overrides default www.platina.com`

	DfltMod = 0755
	MmcDir  = "/mmc"
	MmcDev  = "/dev/mmcblk0p1"
	DfltSrv = "192.168.101.127" //FIXME invader7 for now
	PlatSvr = "www.platina.com"
	Machine = "platina-mk1-bmc"
)

var imageNames = []string{"ubo", "dtb", "env", "ker", "ini"}

type Interface interface {
	Apropos() lang.Alt
	Main(...string) error
	Man() lang.Alt
	String() string
	Usage() string
}

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }

func (cmd) Main(args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("Missing version:  LATEST or hash")
	}
	v := ""
	s := "http://" + DfltSrv
	t := args[0]
	if t == "LATEST" {
		v = "LATEST"
	}
	if len(args) == 3 { //FIXME
		u := args[1]
		if u == "-s" {
			s = "http://" + args[2]
		}
	}
	err := doUpgrade(s, v)
	if err != nil {
		return err
	}
	return nil
}

func (cmd) Man() lang.Alt  { return man }
func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Usage }

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)

func doUpgrade(s string, v string) error {
	err := mountMmc()
	if err != nil {
		return err
	}
	err = getFiles(s, v)
	if err != nil {
		return err
	}
	//err = deBlob() //FIXME
	err = copyFiles(v)
	if err != nil {
		return err
	}
	err = rmFiles(v)
	if err != nil {
		return err
	}
	reboot()
	return nil
}

func reboot() error {
	kexec.Prepare()
	_ = syscall.Reboot(syscall.LINUX_REBOOT_CMD_KEXEC)
	return syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
}

func getFiles(s string, v string) error {
	for _, f := range imageNames {
		files := []string{
			s + "/" + Machine + "/" + Machine +
				"-" + f + "-" + v + ".bin",
			s + "/" + Machine + "/" + Machine +
				"-" + f + "-" + v + ".crc",
		}
		err := wgetFiles(files)
		if err != nil {
			return err
		}
	}
	return nil
}

func wgetFiles(urls []string) error {
	reqs := make([]*grab.Request, 0)
	for _, url := range urls {
		req, err := grab.NewRequest(url)
		if err != nil {
			return err
		}
		reqs = append(reqs, req)
	}

	successes, err := url.FetchReqs(0, reqs)
	if successes == 0 && err != nil {
		return err
	}
	return nil
}

func copyFiles(v string) error {
	for _, f := range imageNames {
		err := copyFile("/"+Machine+"-"+f+"-"+v+
			".bin", MmcDir+"/"+f+".bin")
		if err != nil {
			return err
		}
		err = copyFile("/"+Machine+"-"+f+"-"+v+
			".crc", MmcDir+"/"+f+".crc")
		if err != nil {
			return err
		}
	}
	return nil
}

func copyFile(f string, d string) error {
	sFile, err := os.Open(f)
	if err != nil {
		return err
	}
	defer sFile.Close()

	eFile, err := os.Create(d)
	if err != nil {
		return err
	}
	defer eFile.Close()

	_, err = io.Copy(eFile, sFile)
	if err != nil {
		return err
	}

	err = eFile.Sync()
	if err != nil {
		return err
	}
	return nil
}

func rmFiles(v string) error {
	for _, f := range imageNames {
		err := rmFile("/" + Machine + "-" + f + "-" + v + ".bin")
		if err != nil {
			return err
		}
		err = rmFile("/" + Machine + "-" + f + "-" + v + ".crc")
		if err != nil {
			return err
		}
	}
	return nil
}

func rmFile(f string) error {
	_, err := os.Stat(f)
	if err != nil {
		return err
	}

	if err = os.Remove(f); err != nil {
		return err
	}
	return nil
}

func mountMmc() error {
	var perm os.FileMode = DfltMod

	dn := MmcDir
	mdev := MmcDev
	f := os.MkdirAll

	if err := f(dn, perm); err != nil {
		return err
	}

	err := os.Chdir("/")
	if err != nil {
		return err
	}

	args := []string{" ", " "}
	flag, args := flags.New(args,
		"--fake",
		"-v",
		"-a",
		"-F",
		"-defaults",
		"-p",
		"-r",
		"-read-write",
		"-suid",
		"-no-suid",
		"-dev",
		"-no-dev",
		"-exec",
		"-no-exec",
		"-synchronous",
		"-no-synchronous",
		"-remount",
		"-mand",
		"-no-mand",
		"-dirsync",
		"-no-dirsync",
		"-atime",
		"-no-atime",
		"-diratime",
		"-no-diratime",
		"-bind",
		"-move",
		"-silent",
		"-loud",
		"-posixacl",
		"-no-posixacl",
		"-bindable",
		"-unbindable",
		"-private",
		"-slave",
		"-shared",
		"-relatime",
		"-no-relatime",
		"-iversion",
		"-no-iversion",
		"-strictatime",
		"-no-strictatime")
	parm, args := parms.New(args, "-match", "-o", "-t")
	parm["-t"] = "ext4"

	fs, err := getFilesystems()
	if err != nil {
		return err
	}

	fs.mountone(parm["-t"], mdev, dn, flag, parm)

	return nil
}

func (fs *filesystems) mountone(t, dev, dir string, flag flags.Flag, parm parms.Parm) *MountResult {
	var flags uintptr
	if flag["-defaults"] {
		//  rw, suid, dev, exec, auto, nouser, async
		flags &^= syscall.MS_RDONLY
		flags &^= syscall.MS_NOSUID
		flags &^= syscall.MS_NODEV
		flags &^= syscall.MS_NOEXEC
		if t == "" {
			t = "auto"
		}
		flags |= MS_NOUSER
		flags |= syscall.MS_ASYNC
	}
	for _, x := range translations {
		if flag[x.name] {
			if x.set {
				flags |= x.bits
			} else {
				flags &^= x.bits
			}
		}
	}
	if flag["--fake"] {
		return &MountResult{nil, dev, t, dir, flag}
	}

	tryTypes := []string{t}
	nodev := false
	if t == "auto" {
		tryTypes = fs.autoList
	} else {
		nodev = fs.isNoDev[t]
	}

	if !nodev {
		_, err := readSuperBlock(dev)
		if err != nil {
			return &MountResult{err, dev, t, dir, flag}
		}
	}

	var err error
	for _, t := range tryTypes {
		for i := 0; i < 5; i++ {
			err = syscall.Mount(dev, dir, t, flags, parm["-o"])
			if err == nil {
				return &MountResult{err, dev, t, dir, flag}
			}
			if err == syscall.EBUSY {
				time.Sleep(1 * time.Second)
				continue
			}
			break
		}
	}

	return &MountResult{err, dev, t, dir, flag}
}

// hack around syscall incorrect definition
const MS_NOUSER uintptr = (1 << 31)
const procFilesystems = "/proc/filesystems"

type fstabEntry struct {
	fsSpec  string
	fsFile  string
	fsType  string
	mntOpts string
}

type fsType struct {
	name  string
	nodev bool
}

type filesystems struct {
	isNoDev  map[string]bool
	autoList []string
}

var translations = []struct {
	name string
	bits uintptr
	set  bool
}{
	{"-read-only", syscall.MS_RDONLY, true},
	{"-read-write", syscall.MS_RDONLY, false},
	{"-suid", syscall.MS_NOSUID, false},
	{"-no-suid", syscall.MS_NOSUID, true},
	{"-dev", syscall.MS_NODEV, false},
	{"-no-dev", syscall.MS_NODEV, true},
	{"-exec", syscall.MS_NOEXEC, false},
	{"-no-exec", syscall.MS_NOEXEC, true},
	{"-synchronous", syscall.MS_SYNCHRONOUS, true},
	{"-no-synchronous", syscall.MS_SYNCHRONOUS, true},
	{"-remount", syscall.MS_REMOUNT, true},
	{"-mand", syscall.MS_MANDLOCK, true},
	{"-no-mand", syscall.MS_MANDLOCK, false},
	{"-dirsync", syscall.MS_DIRSYNC, true},
	{"-no-dirsync", syscall.MS_DIRSYNC, false},
	{"-atime", syscall.MS_NOATIME, false},
	{"-no-atime", syscall.MS_NOATIME, true},
	{"-diratime", syscall.MS_NODIRATIME, false},
	{"-no-diratime", syscall.MS_NODIRATIME, true},
	{"-bind", syscall.MS_BIND, true},
	{"-move", syscall.MS_MOVE, true},
	{"-silent", syscall.MS_SILENT, true},
	{"-loud", syscall.MS_SILENT, false},
	{"-posixacl", syscall.MS_POSIXACL, true},
	{"-no-posixacl", syscall.MS_POSIXACL, false},
	{"-bindable", syscall.MS_UNBINDABLE, false},
	{"-unbindable", syscall.MS_UNBINDABLE, true},
	{"-private", syscall.MS_PRIVATE, true},
	{"-slave", syscall.MS_SLAVE, true},
	{"-shared", syscall.MS_SHARED, true},
	{"-relatime", syscall.MS_RELATIME, true},
	{"-no-relatime", syscall.MS_RELATIME, false},
	{"-iversion", syscall.MS_I_VERSION, true},
	{"-no-iversion", syscall.MS_I_VERSION, false},
	{"-strictatime", syscall.MS_STRICTATIME, true},
	{"-no-strictatime", syscall.MS_STRICTATIME, false},
}

type MountResult struct {
	err    error
	dev    string
	fstype string
	dir    string
	flag   flags.Flag
}

func (r *MountResult) String() string {
	if r.err != nil {
		return fmt.Sprintf("%s: %v", r.dev, r.err)
	}
	if r.flag["--fake"] {
		return fmt.Sprintf("Would mount %s type %s at %s", r.dev, r.fstype, r.dir)
	}
	if r.flag["-v"] {
		return fmt.Sprintf("Mounted %s type %s at %s", r.dev, r.fstype, r.dir)
	}
	return ""
}

func (r *MountResult) ShowResult() {
	s := r.String()
	if s != "" {
		fmt.Println(s)
	}
}

type superBlock interface {
}

type unknownSB struct {
}

const (
	ext234SMagicOffL = 0x438
	ext234SMagicOffM = 0x439
	ext234SMagicValL = 0x53
	ext234SMagicValM = 0xef
)

type ext234 struct {
}

func readSuperBlock(dev string) (superBlock, error) {
	f, err := os.Open(dev)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fsHeader := make([]byte, 4096)
	_, err = f.Read(fsHeader)
	if err != nil {
		return nil, err
	}

	if fsHeader[ext234SMagicOffL] == ext234SMagicValL &&
		fsHeader[ext234SMagicOffM] == ext234SMagicValM {
		sb := &ext234{}
		return sb, nil
	}

	return &unknownSB{}, nil
}

func getFilesystems() (fsPtr *filesystems, err error) {
	f, err := os.Open(procFilesystems)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var fs filesystems
	fs.isNoDev = make(map[string]bool)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		nodev := false
		if strings.HasPrefix(line, "nodev") {
			nodev = true
			line = strings.TrimPrefix(line, "nodev")
		}
		line = strings.TrimSpace(line)
		fs.isNoDev[line] = nodev
		if !nodev {
			fs.autoList = append(fs.autoList, line)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "scan:", procFilesystems, err)
		return nil, err
	}
	return &fs, nil
}
