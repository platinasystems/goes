// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// make vfs for each given pf
package sriovs

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/platinasystems/go/internal/assert"
	"github.com/platinasystems/go/internal/redis"
)

const DefaultNumvfs = 16

type Mac [6]byte
type Vf uint

func (vf Vf) Port() uint    { return uint((vf &^ (Port(1) - 1)) >> 20) }
func (vf Vf) SubPort() uint { return uint((vf & 0xf0000) >> 16) }
func (vf Vf) Vlan() uint    { return uint(vf & 0xffff) }

func Port(u uint) Vf    { return Vf(u << 20) }
func SubPort(u uint) Vf { return Vf((u & 0xf) << 16) }
func Vlan(u uint) Vf    { return Vf(u & 0xffff) }

func Mksriovs(porto uint, vfs ...[]Vf) error {
	err := assert.Root()
	if err != nil {
		return err
	}
	numpfs := len(vfs)
	numvfs := DefaultNumvfs
	if s, _ := redis.Hget(redis.DefaultHash, "sriov.numvfs"); len(s) > 0 {
		_, err = fmt.Sscan(s, &numvfs)
		if err != nil {
			return fmt.Errorf("sriov.numvfs: %v", err)
		}
	}
	pfs, totalvfs, err := pfinit(numpfs, numvfs)
	if err != nil {
		return err
	}
	for pfi, pf := range pfs {
		var mac Mac
		var virtfns []string

		copy(mac[:], pf.HardwareAddr)
		mac.Plus(uint(len(pfs) - pfi + (pfi * totalvfs)))

		virtfns, err = pfvirtfns(pf.Name, numvfs)
		if err != nil {
			return err
		}
		for _, virtfn := range virtfns {
			vfi, err := getVfi(virtfn)
			if err != nil {
				return err
			}
			if vfi >= len(vfs[pfi]) {
				continue
			}
			vf := vfs[pfi][vfi]
			err = ifset(pf.Name, "vf", vfi, "mac", mac, "vlan",
				vf.Vlan())
			if err != nil {
				return err
			}
			mac.Plus(1)
			vfname, err := getVfname(virtfn)
			if err != nil {
				return err
			}
			want := fmt.Sprintf("eth-%d-%d", vf.Port()+porto,
				vf.SubPort()+porto)
			if vfname != want {
				err = ifset(vfname, "name", want)
				if err != nil {
					return err
				}
				vfname = want
			}
			// bounce vf to reload its mac from the pf
			if err = ifset(vfname, "up"); err != nil {
				return err
			}
			if err = ifset(vfname, "down"); err != nil {
				return err
			}
		}
	}
	return err
}

func (mac *Mac) Plus(u uint) {
	base := mac[5]
	mac[5] += byte(u)
	if mac[5] < base {
		base = mac[4]
		mac[4] += 1
		if mac[4] < base {
			mac[3] += 1
		}
	}
}

func (mac Mac) String() string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%2x",
		mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
}

func pfinit(numpfs, numvfs int) (devs []net.Interface, total int, err error) {
	all, err := net.Interfaces()
	if err != nil {
		return
	}
	for _, dev := range all {
		var cur int
		fn := filepath.Join("/sys/class/net", dev.Name,
			"device/sriov_numvfs")
		f, terr := os.Open(fn)
		if terr != nil {
			continue
		}
		_, terr = fmt.Fscan(f, &cur)
		f.Close()
		if terr != nil {
			continue
		}
		if dev.Flags&net.FlagUp != net.FlagUp {
			if err = ifset(dev.Name, "up"); err != nil {
				return
			}
		}
		devs = append(devs, dev)
		if cur < numvfs {
			f, err = os.OpenFile(fn, os.O_WRONLY|os.O_TRUNC, 0)
			if err != nil {
				return
			}
			_, err = fmt.Fprintln(f, numvfs)
			f.Close()
			if err != nil {
				return
			}
		}
	}
	if len(devs) == 0 {
		err = fmt.Errorf("no sriovs")
	}
	f, err := os.Open(filepath.Join("/sys/class/net", devs[0].Name,
		"device/sriov_totalvfs"))
	if err != nil {
		return
	}
	_, err = fmt.Fscan(f, &total)
	f.Close()
	return
}

func pfvirtfns(pfname string, numvfs int) (virtfns []string, err error) {
	pat := filepath.Join("/sys/class/net", pfname, "device/virtfn*")
	for tries := 0; true; tries++ {
		virtfns, err = filepath.Glob(pat)
		if err == nil && len(virtfns) >= numvfs {
			break
		}
		if tries == 5 {
			err = fmt.Errorf("%s: vf t/o", pfname)
			break
		}
		time.Sleep(time.Second)
	}
	return
}

func getVfi(virtfn string) (int, error) {
	var vfi int
	base := filepath.Base(virtfn)
	s := strings.TrimPrefix(base, "virtfn")
	_, err := fmt.Sscan(s, &vfi)
	return vfi, err
}

func getVfname(virtfn string) (string, error) {
	dn := filepath.Join(virtfn, "net")
	for tries := 0; true; tries++ {
		dir, err := ioutil.ReadDir(dn)
		if err == nil {
			if len(dir) == 0 {
				return "", fmt.Errorf("%s: empty", dn)
			}
			return dir[0].Name(), nil
		}
		if tries == 5 {
			return "", fmt.Errorf("%s: vf t/o", dn)
		}
		time.Sleep(time.Second)
	}
	panic("oops")
}

func getVfdev(vfname string) (*net.Interface, error) {
	for tries := 0; true; tries++ {
		vfdev, err := net.InterfaceByName(vfname)
		if err == nil {
			return vfdev, nil
		}
		if tries == 5 {
			return nil, fmt.Errorf("%s: t/o", vfname)
		}
		time.Sleep(time.Second)
	}
	panic("oops")
}

func ifset(name string, args ...interface{}) error {
	cmd := exec.Command("ip", "link", "set", name)
	for _, arg := range args {
		cmd.Args = append(cmd.Args, fmt.Sprint(arg))
	}
	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("%v: %v", cmd.Args, err)
	}
	return err
}
