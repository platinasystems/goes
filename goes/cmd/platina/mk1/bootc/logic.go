// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// bootc.go Module - GOES-BOOT and GOES
//   provides Client REST messages to server
//   provides state machine for executing wipe/re-install in goesboot
//   executes Register(), receives Client struct, script
//
// bootd.go Daemon - GOES
//   provides Server REST messages to respond to master
//
// pushd.go Daemon - GOES
//   executes wipe request from bootd
//   pushes boot state, install state, etc. to master

package bootc

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/platinasystems/redis"
)

const (
	minCoreVer   = 0.41
	minCoreCom   = 299
	verNum       = "1.13"
	goesBootCfg  = "/mountd/sda1/bootc.cfg"
	sda1Cfg      = "/bootc.cfg"
	sda6Cfg      = "/sda1/bootc.cfg"
	goesBootPath = "/mountd/sda1/"
	sda1Path     = "/"
	sda6Path     = "/sda1/"
	mount        = "/sda1"
	devSda1      = "/dev/sda1"
	devSda6      = "/dev/sda6"
	tmpFile      = "/tmp/EEOF"
	sda1Etc      = "/sda1/etc"
	fstype       = "ext4"
	zero         = uintptr(0)
	sda1         = "sda1"
	sda6         = "sda6"
	goesBoot     = "coreboot"
	cbSda1       = "/mountd/sda1/"
	cbSda6       = "/mountd/sda6/"
	tarFile      = "postinstall.tar.gz"
	scriptFile   = "rc.local"
	waitMs       = 100
	timeoutCnt   = 50
	Mach         = "mk1"
	Machine      = "platina-" + Mach
	CorebootName = "coreboot-" + Machine + ".rom"
	sda6cntEnb   = false
)

var BootcCfgFile string
var uuid1 string
var uuid1w string
var uuid6 string
var uuid6w string
var kexec0 string
var kexec1 string
var kexec6 string

var fileList = [...]string{
	"debian-8.10.0-amd64-DVD-1.iso",
	"preseed.cfg",
	"hd-media/preseed.cfg",
	"boot/vmlinuz",
	"boot/initrd.gz",
	"usr/local/sbin/flashrom",
	"usr/local/share/flashrom/layouts",
	"rc.local",
	"postinstall.tar.gz",
}

func Bootc() []string {
	for i := 0; ; i++ {
		time.Sleep(time.Millisecond * time.Duration(waitMs))
		if i > timeoutCnt {
			return []string{""}
		}
		p, err := ioutil.ReadFile("/proc/partitions")
		if err != nil {
			continue
		}
		parts := string(p)
		m, err := ioutil.ReadFile("/proc/mounts")
		if err != nil {
			continue
		}
		mounts := string(m)
		if err := readCfg(); err != nil {
			continue
		} else {
			if err := writeCfg(); err != nil { // updates bootc format
				fmt.Println("Error: writing bootc.cfg, drop into grub...\n")
				return []string{""}
			}
		}
		if Cfg.Disable {
			return []string{""}
		}

		fixSda1Swap()

		if err := fixNewroot(); err != nil {
			fmt.Println("Error: can't fix newroot, drop into grub...")
			return []string{""}
		}

		// sda1 utility mode
		if Cfg.BootSda1 && strings.Contains(mounts, "sda1") {
			if err := formKexec1(); err != nil {
				fmt.Println("Error: can't form kexec string, drop into grub...")
				return []string{""}
			}
			if err := clrSda1Flag(); err != nil {
				fmt.Println("Error: can't clear sda1 flag, drop into grub...")
				return []string{""}
			}
			// boot sda1 for recovery
			return []string{"kexec", "-k", Cfg.Sda1K,
				"-i", Cfg.Sda1I, "-c", kexec1, "-e"}
		}

		// install
		if Cfg.Install && strings.Contains(mounts, "sda1") && !strings.Contains(parts, "sda6") {
			if err := formKexec1(); err != nil {
				fmt.Println("Error: can't form install kexec, drop into grub...")
				return []string{""}
			}
			if err := clrInstall(); err != nil {
				fmt.Println("Error: can't clear install bit, drop into grub...")
				return []string{""}
			}
			//TODO Comment out until post script is fixed TODO if err := setPostInstall(); err != nil {
			//	fmt.Println("Error: can't set postinstall bit, drop into grub...")
			//	return []string{""}
			//}
			// boot installer from sda1
			return []string{"kexec", "-k", Cfg.ReInstallK, "-i",
				Cfg.ReInstallI, "-c", kexec0, "-e"}
		}

		// postinstall
		if Cfg.PostInstall && strings.Contains(mounts, sda6) && strings.Contains(mounts, sda1) {
			if err := clrPostInstall(); err != nil {
				fmt.Println("Error: post install copy failed, drop into grub...")
				return []string{""}
			}
			// FIXME update this per new architecture
			if err := Copy(cbSda1+tarFile, cbSda6+tarFile); err != nil {
				fmt.Println("Error: post install copy failed, drop into grub...")
				return []string{""}
			}
			if err := Copy(cbSda1+scriptFile, cbSda6+"etc/"+scriptFile); err != nil {
				fmt.Println("Error: post install copy failed, drop into grub...")
				return []string{""}
			}
		}

		// DHCP, ZTP, PCC
		if true {
			var i = 0
			for start := time.Now(); time.Since(start) < 10*time.Second; {
				i++
			}

			// SETUP PHASE - set dhcp bits, static IP
			//
			// D10 3 SEC LOOP wait for console interrupt --> YES, E10 goes-boot shell
			//
			// D20 static IP configured? (bootc.cfg)     --> NO,  goto D60 check if PCC enabled
			//                                                    skip ZTP & DHCP
			//                                                    P10 set DHCP option 55 and option 61
			//
			// D30a Is ZTP Enabled?                      --> YES, P20 set Option 43 (ZTP script)

			// DHCP PHASE
			//
			// D40 INFINITE LOOP DHCP server found?      --> NO, check for ESC --> E20a shell
			//
			// ... DO DHCP

			// ZTP PHASE
			//
			// D30b Is ZTP Enabled?                      --> NO, goto D60, PCC Enb check
			//
			// D50 INFINITE LOOP ZTP server found?       --> NO, check for ESC --> E20b shell
			//
			// P30 set "pixie boot" bit in bootc.cfg, download images, boot them, clr pixie bit

			// PCC PHASE
			//
			// D60 Is PCC Enabled                        --> NO, goto P50 boot sda6
			//
			// D70 INFINITE LOOP, PCC server found?      --> ESC - goes-boot shell
			//
			// P40 register with PCC
			//
			// P50 boot sda6

			// SDA6 BOOT
			// D80 Is PCC Enabled                        --> NO, DONE
			//
			// P60 check in with PCC (register) goes register
			//
			// keep alives, status updates, control etc. from PCC...
			//
			// Transparent to invader: PCC will push additional configs via ansible
			//
			// Transparent to invader: keep alives, status updates, control etc. from PCC
		}

		// sda6 normal
		if (Cfg.BootSda6Cnt > 0 || !sda6cntEnb) &&
			strings.Contains(parts, sda6) && strings.Contains(mounts, sda6) {
			if err := fixPaths(); err != nil {
				fmt.Println("Error: can't fix paths, drop into grub...")
				return []string{""}
			}
			if err := formKexec6(); err != nil {
				fmt.Println("Error: can't form sda6 kexec, drop into grub...")
				return []string{""}
			}
			if sda6cntEnb {
				if err := decBootSda6Cnt(); err != nil {
					fmt.Println("Error: can't decrement sda6cnt, drop into grub...")
					return []string{""}
				}
			}
			// boot sda6
			return []string{"kexec", "-k", Cfg.Sda6K,
				"-i", Cfg.Sda6I, "-c", kexec6, "-e"}
		}

		// non-partitioned
		if !strings.Contains(parts, sda6) && strings.Contains(mounts, sda1) {
			if err := formKexec1(); err != nil {
				fmt.Println("Error: can't form kexec string, drop into grub...")
				return []string{""}
			}
			if err := clrSda1Flag(); err != nil {
				fmt.Println("Error: can't clear sda1 flag, drop into grub...")
				return []string{""}
			}
			// boot sda1 if we are not partitioned
			return []string{"kexec", "-k", Cfg.Sda1K,
				"-i", Cfg.Sda1I, "-c", kexec1, "-e"}
		}
	}

	fmt.Println("Error: bootc can't boot, drop into grub...")
	return []string{""}
}

func checkFiles() bool {
	context, err := getContext()
	if err != nil {
		return false
	}
	path := ""
	switch context {
	case goesBoot:
		//fmt.Println("TEMP: goes-boot context.")
		path = goesBootPath
	case sda1:
		//fmt.Println("TEMP: goes sda1 context.")
		path = sda1Path
	case sda6:
		//fmt.Println("TEMP: goes sda6 context.")
		path = sda6Path
		if err := mountSda1(); err != nil {
			//fmt.Println("TEMP: goes sda6 mount-sda1 fail.")
			return false
		}
		//fmt.Println("TEMP: goes sda6 mount-sda1 success.")
	default:
		fmt.Println("ERROR: could not determine context.")
		return false
	}

	good := true
	//FIXME change names of .iso's to original.iso, latest.iso
	for _, f := range fileList {
		fmt.Println("Checking", path+f, "exists...")
		if _, err := os.Stat(path + f); os.IsNotExist(err) {
			fmt.Println("ERROR:	file", path+f, "does not exist")
			good = false
		}
	}
	// 1. check for goes running
	// 2. check for sda6cnt == 3
	// 3. check the sda6 k string, see it matches /boot
	// 4. check the sda6 i string, see it matches /boot
	if good {
		fmt.Println("PASSED: wipe/reinstall is configured properly.")
		return true
	}
	fmt.Println("FAILED: wipe/reinstall is NOT configured properly.")
	return false
}

func getFiles() error {
	context, err := getContext()
	if err != nil {
		return err
	}
	//path := ""
	switch context {
	case goesBoot:
		//fmt.Println("TEMP: goes-boot context.")
		//path = goesBootPath
	case sda1:
		//fmt.Println("TEMP: goes sda1 context.")
		//path = sda1Path
	case sda6:
		//fmt.Println("TEMP: goes sda6 context.")
		//path = sda6Path
		if err := mountSda1(); err != nil {
			//fmt.Println("TEMP: goes sda6 mount-sda1 fail.")
			return err
		}
		//fmt.Println("TEMP: goes sda6 mount-sda1 success.")
	default:
		fmt.Println("ERROR: could not determine context.")
		return nil
	}

	//fmt.Println("TEMP: sda1 path = ", path)
	return nil
}

func (c *Command) bootc() {
	if kexec := Bootc(); len(kexec) > 1 {
		err := c.Main(kexec...)
		fmt.Println(err)
	}
	return
}

func initCfg() error {
	Cfg = BootcConfig{
		Install:         false,
		BootSda1:        false,
		BootSda6Cnt:     3,
		EraseSda6:       false,
		IAmMaster:       false,
		MyIpAddr:        "192.168.101.129",
		MyGateway:       "192.168.101.1",
		MyNetmask:       "255.255.255.0",
		MasterAddresses: []string{"198.168.101.142"},
		ReInstallK:      "/mountd/sda1/boot/vmlinuz",
		ReInstallI:      "/mountd/sda1/boot/initrd.gz",
		ReInstallC:      `netcfg/get_hostname=platina netcfg/get_domain=platinasystems.com interface=auto auto locale=en_US preseed/file=/hd-media/preseed.cfg`,
		Sda1K:           "/mountd/sda1/boot/vmlinuz-3.16.0-4-amd64",
		Sda1I:           "/mountd/sda1/boot/initrd.img-3.16.0-4-amd64",
		Sda1C:           "::eth0:none",
		Sda6K:           "/mountd/sda6/boot/vmlinuz-3.16.0-4-amd64",
		Sda6I:           "/mountd/sda6/boot/initrd.img-3.16.0-4-amd64",
		Sda6C:           "::eth0:none",
		ISO1Name:        "debian-8.10.0-amd64-DVD-1.iso",
		ISO1Desc:        "Jessie debian-8.10.0",
		ISO2Name:        "",
		ISO2Desc:        "",
		ISOlastUsed:     1,
		PostInstall:     false,
		Disable:         false,
		PccEnb:          false,
		PccIP:           "",
		PccPort:         "",
		PccSN:           "",
	}
	if err := writeCfg(); err != nil {
		return err
	}
	return nil
}

func getContext() (context string, err error) {
	mach := redis.DefaultHash
	if mach == goesBoot {
		return goesBoot, nil
	}
	if mach == "platina-mk1" {
		cmd := exec.Command("df")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", err
		}
		outs := strings.Split(string(out), "\n")
		for _, m := range outs {
			if strings.Contains(m, devSda1) {
				return sda1, nil
			}
			if strings.Contains(m, devSda6) {
				return sda6, nil
			}
		}
	}
	return "", fmt.Errorf("Error: root directory not found")
}

func mountSda1() error {
	if _, err := os.Stat(mount); os.IsNotExist(err) {
		err := os.Mkdir(mount, os.FileMode(0755))
		if err != nil {
			fmt.Printf("Error mkdir: %v", err)
			return err
		}
	}
	if _, err := os.Stat(sda1Etc); os.IsNotExist(err) {
		if err := syscall.Mount(devSda1, mount, fstype, zero, ""); err != nil {
			fmt.Printf("Error mounting: %v", err)
			return err
		}
	}
	return nil
}

func setBootcCfgFile() error {
	context, err := getContext()
	if err != nil {
		return err
	}
	switch context {
	case goesBoot:
		BootcCfgFile = goesBootCfg
	case sda1:
		BootcCfgFile = sda1Cfg
	case sda6:
		BootcCfgFile = sda6Cfg
		if err := mountSda1(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Error: unknown machine/partition")
	}
	return nil
}

func writeCfg() error {
	if err := setBootcCfgFile(); err != nil {
		return err
	}
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	return nil
}

func readCfg() error {
	if err := setBootcCfgFile(); err != nil {
		return err
	}
	if _, err := os.Stat(BootcCfgFile); os.IsNotExist(err) {
		if err = initCfg(); err != nil {
			return err
		}
	}
	dat, err := ioutil.ReadFile(BootcCfgFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(dat, &Cfg)
	if err != nil {
		return err
	}
	return nil
}

func formKexec1() (err error) {
	uuid1, err = readUUID("sda1")
	if err != nil {
		return err
	}
	kexec0 = "root=UUID=" + uuid1 + " console=ttyS0,115200 " + Cfg.ReInstallC
	kexec1 = "root=UUID=" + uuid1 + " console=ttyS0,115200 "
	kexec1 += Cfg.Sda1C
	return nil
}

func formKexec6() (err error) {
	uuid6, err = readUUID("sda6")
	if err != nil {
		return err
	}
	kexec6 = "root=UUID=" + uuid6 + " console=ttyS0,115200"
	kexec6 += Cfg.Sda6C
	return nil
}

func readUUID(partition string) (uuid string, err error) {
	dat, err := ioutil.ReadFile("/mountd/" + partition + "/etc/fstab")
	if err != nil {
		return "", err
	}
	dat1 := strings.Split(string(dat), "UUID=")
	dat2 := strings.Split(dat1[2], "/")
	dat3 := []byte(dat2[0])
	len3 := len(dat3) - 1
	dat4 := string(dat3[0:len3])
	uuid = string(dat4)
	return uuid, nil
}

func ResetSda6Count() error {
	c, err := getContext()
	if err != nil {
		return err
	}
	if c != sda6 {
		return nil
	}
	if err = readCfg(); err != nil {
		return err
	}
	Cfg.BootSda6Cnt = 3
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	return nil
}

func setSda6Count(x string) error {
	if err := readCfg(); err != nil {
		return err
	}
	i, err := strconv.Atoi(x)
	if err != nil {
		return err
	}
	Cfg.BootSda6Cnt = i
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	return nil
}

func clrSda6Count() error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.BootSda6Cnt = 0
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	return nil
}

func decBootSda6Cnt() error {
	if err := readCfg(); err != nil {
		return err
	}
	x := Cfg.BootSda6Cnt
	if x > 0 {
		x--
	}
	Cfg.BootSda6Cnt = x
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	return nil
}

func setInstall() error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.Install = true
	Cfg.BootSda6Cnt = 3
	Cfg.PostInstall = false
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	return nil
}

func clrInstall() error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.BootSda1 = false
	Cfg.Install = false
	Cfg.PostInstall = false
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	return nil
}

func setPostInstall() error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.PostInstall = true
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	return nil
}

func clrPostInstall() error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.PostInstall = false
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	return nil
}

func setDisable() error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.Disable = true
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	return nil
}

func clrDisable() error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.Disable = false
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	return nil
}

func setPccEnb() error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.PccEnb = true
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	pccInit()
	return nil
}

func clrPccEnb() error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.PccEnb = false
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	pccInit()
	return nil
}

func setPccIP(x string) error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.PccIP = x
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	pccInit()
	return nil
}

func setPccPort(x string) error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.PccPort = x
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	pccInit()
	return nil
}

func setPccSN(x string) error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.PccSN = x
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	pccInit()
	return nil
}

func setSda1Flag() error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.BootSda1 = true
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	return nil
}

func clrSda1Flag() error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.BootSda1 = false
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	return nil
}

func setIp(x string) error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.MyIpAddr = x
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	return nil
}

func setNetmask(x string) error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.MyNetmask = x
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	return nil
}

func setGateway(x string) error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.MyGateway = x
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	return nil
}

func SetSda1K(x string) error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.Sda1K = x
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	return nil
}

func SetSda1I(x string) error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.Sda1I = x
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	return nil
}

func SetSda6K(x string) error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.Sda6K = x
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	return nil
}

func SetSda6I(x string) error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.Sda6I = x
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	return nil
}

func UpdateBootcCfg(k, i string) error {
	k = cbSda6 + "boot/" + k
	if err := SetSda6K(k); err != nil {
		return err
	}
	i = cbSda6 + "boot/" + i
	if err := SetSda6I(i); err != nil {
		return err
	}
	return nil
}

func Wipe(dryrun bool) error {
	if dryrun {
		fmt.Println("Start wipe dryrun.  Does not write to disk.")
	}
	fmt.Println("Making sure we booted from sda1 or sda6...")
	context, err := getContext()
	if context != sda6 && context != sda1 {
		return fmt.Errorf("Not booted from sda6 or sda1, can't wipe.")
	}
	if !dryrun {
		if err := clrInstall(); err != nil {
			return err
		}
	}

	fmt.Println("Making sure sda6 exists...")
	d1 := []byte("#!/bin/bash\necho -e " + `"p\nq\n"` + " | /sbin/fdisk /dev/sda\n")
	if err = ioutil.WriteFile(tmpFile, d1, 0755); err != nil {
		return err
	}
	cmd := exec.Command(tmpFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("fdisk: %v, %v\n", out, err)
		return err
	}
	outs := strings.Split(string(out), "\n")
	n := 0
	for _, m := range outs {
		if strings.Contains(m, devSda6) {
			n = 1
		}
	}
	if n == 0 {
		return fmt.Errorf("Error: /dev/sda6 not in partition table, aborting")
	}

	fmt.Println("Making sure sda1 has all the re-install files...")
	if passed := checkFiles(); !passed {
		return fmt.Errorf("Not all files exist on sda1, aborting...")
	}

	fmt.Println("Check coreboot version...")
	var im IMGINFO
	if im, err = getCorebootInfo(); err != nil {
		fmt.Println("Coreboot could not be read, aborting...")
		return err
	}
	if len(im.Extra) == 0 {
		fmt.Println("Coreboot git version doesn't exist in rom, aborting")
		return fmt.Errorf("Couldn't determine coreboot version")
	}
	y := strings.Split(im.Extra, "-")
	z := strings.Split(y[0], "v")
	w, err := strconv.Atoi(y[1])
	if err != nil {
		return fmt.Errorf("Couldn't determine coreboot version, aborting...")
	}
	v, err := strconv.ParseFloat(z[1], 64)
	if err != nil {
		return fmt.Errorf("Couldn't determine coreboot version, aborting...")
	}
	if v < minCoreVer {
		return fmt.Errorf("Coreboot needs upgraded.")
	}
	if v == minCoreVer && w < minCoreCom {
		return fmt.Errorf("Coreboot needs upgraded.")
	}
	fmt.Printf("Coreboot version ok, ver = %d, subver = %d\n", v, w)

	if !dryrun {
		fmt.Println("Deleting sda6 from the partition table...")
		d1 = []byte("#!/bin/bash\necho -e " + `"d\n6\nw\n"` + " | /sbin/fdisk /dev/sda\n")
		if err = ioutil.WriteFile(tmpFile, d1, 0755); err != nil {
			return err
		}
		cmd = exec.Command(tmpFile)
		out, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("fdisk: %v, %v\n", out, err)
		}
	}

	if !dryrun {
		fmt.Println("Verify sda6 is actually gone...")
		d1 = []byte("#!/bin/bash\necho -e " + `"p\nq\n"` + " | /sbin/fdisk /dev/sda\n")
		if err = ioutil.WriteFile(tmpFile, d1, 0755); err != nil {
			return err
		}
		cmd = exec.Command(tmpFile)
		out, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("fdisk: %v, %v\n", out, err)
			return err
		}
		outs = strings.Split(string(out), "\n")
		for _, m := range outs {
			if strings.Contains(m, devSda6) {
				return fmt.Errorf("Error: /dev/sda6 not deleted, aborting wipe")
			}
		}
		fmt.Println("Verified")
	}

	if !dryrun {
		fmt.Println("Setting Install bit for coreboot...")
		if err := setInstall(); err != nil {
			return err
		}
	}

	if dryrun {
		fmt.Println("Wipe dryrun passed...")
	}
	return nil
}

func fixPaths() error { //FIXME Temporary remove by 9/30/2018
	files, err := ioutil.ReadDir(cbSda6 + "boot")
	if err != nil {
		return err
	}
	k := ""
	kn := ""
	for _, f := range files {
		if strings.Contains(f.Name(), "vmlinuz") {
			fn := strings.Replace(f.Name(), ".", "", 1)
			if fn > kn {
				k = f.Name()
				kn = fn
			}
		}
	}
	i := strings.Replace(k, "vmlinuz", "initrd.img", 1)
	err = UpdateBootcCfg(k, i)
	if err != nil {
		return err
	}
	return nil
}

func fixNewroot() error { // FIXME Temporary remove by 9/30/2018
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.ReInstallK = strings.Replace(Cfg.ReInstallK, "newroot", "mountd", 1)
	Cfg.ReInstallI = strings.Replace(Cfg.ReInstallI, "newroot", "mountd", 1)
	Cfg.Sda1K = strings.Replace(Cfg.Sda1K, "newroot", "mountd", 1)
	Cfg.Sda1I = strings.Replace(Cfg.Sda1I, "newroot", "mountd", 1)
	Cfg.Sda6K = strings.Replace(Cfg.Sda6K, "newroot", "mountd", 1)
	Cfg.Sda6I = strings.Replace(Cfg.Sda6I, "newroot", "mountd", 1)
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	return nil
}

func reboot() error {
	fmt.Print("\nWILL REBOOT NOW!!!\n")
	u, err := exec.Command("shutdown", "-r", "now").Output()
	fmt.Println(u)
	if err != nil {
		return err
	}
	return nil
}

func Copy(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func getCorebootInfo() (im IMGINFO, err error) {
	im.Name = CorebootName
	_, err = exec.Command("/usr/local/sbin/flashrom", "-p",
		"internal:boardmismatch=force", "-l",
		"/usr/local/share/flashrom/layouts/platina-mk1.xml",
		"-i", "bios", "-r", "cb.rom", "-A", "-V").Output()
	if err != nil {
		return im, err
	}
	a, err := ioutil.ReadFile("cb.rom")
	if err != nil {
		return im, err
	}
	temp := strings.Split(string(a), "\n")
	for _, j := range temp {
		if strings.Contains(j, "COREBOOT_VERSION ") {
			x := strings.Split(j, " ")
			im.Tag = strings.Replace(x[2], `"`, "", 2)
		}
		if strings.Contains(j, "COREBOOT_EXTRA_VERSION ") {
			x := strings.Split(j, " ")
			im.Extra = strings.Replace(x[2], `"`, "", 2)
		}
		if strings.Contains(j, "COREBOOT_ORIGIN_GIT_REVISION ") {
			x := strings.Split(j, " ")
			im.Commit = strings.Replace(x[2], `"`, "", 2)
		}
		if strings.Contains(j, "COREBOOT_BUILD ") {
			x := strings.SplitAfterN(j, " ", 3)
			im.Build = strings.Replace(x[2], `"`, "", 2)
		}
	}
	im.User = ""
	fi, err := os.Stat("cb.rom")
	if err != nil {
		return im, err
	}
	im.Size = fmt.Sprintf("%d", fi.Size())
	return im, nil
}

func fixSda1Swap() error {
	context, err := getContext()
	if err != nil {
		return fmt.Errorf("ERROR: cound not detemine context.")
	}
	if context != goesBoot {
		return nil
	}

	if _, err := os.Stat(cbSda6 + "etc/fstab"); os.IsNotExist(err) {
		fmt.Println("ERROR:	file", cbSda6+"etc/fstab", "does not exist")
		return nil
	}
	d6, err := ioutil.ReadFile(cbSda6 + "etc/fstab")
	if err != nil {
		return err
	}
	ds6 := strings.Split(string(d6), "\n")

	if _, err := os.Stat(cbSda1 + "etc/fstab"); os.IsNotExist(err) {
		fmt.Println("ERROR:	file", cbSda1+"etc/fstab", "does not exist")
		return fmt.Errorf("ERROR: cound not read sda1 /etc/fstab.")
	}
	d1, err := ioutil.ReadFile(cbSda1 + "etc/fstab")
	if err != nil {
		return err
	}
	ds1 := strings.Split(string(d1), "\n")
	uuid6 = ""
	uuid6w = ""
	for _, j := range ds6 {
		if strings.Contains(j, "swap") {
			if strings.Contains(j, "UUID") {
				k := strings.Split(j, " ")
				kk := strings.Split(k[0], "=")
				uuid6 = kk[1]
				uuid6w = j
			}
		}
	}
	uuid1 = ""
	uuid1w = ""
	for _, j := range ds1 {
		if strings.Contains(j, "swap") {
			if strings.Contains(j, "UUID") {
				k := strings.Split(j, " ")
				kk := strings.Split(k[0], "=")
				uuid1 = kk[1]
				uuid1w = j
			}
		}
	}
	if uuid1 == "" || uuid6 == "" {
		fmt.Println("Error: no uuids")
		return fmt.Errorf("ERROR: no UUID for swap in /etc/fstab.")
	}
	if uuid6 == uuid1 {
		return nil
	}
	for i, j := range ds1 {
		if strings.Contains(j, "swap") {
			if strings.Contains(j, "UUID") {
				ds1[i] = uuid6w
			}
		}
	}
	mm := strings.Join(ds1, "\n")
	m := []byte(mm)
	err = ioutil.WriteFile(cbSda1+"etc/fstab", m, 0644)
	if err != nil {
		return err
	}
	return nil
}
