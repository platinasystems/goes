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
	"sort"
	"strings"
	"time"

	"github.com/platinasystems/go/internal/redis"
)

const DefaultNumvfs = 32

type Mac net.HardwareAddr
type Vf uint

type Pf struct {
	net.Interface
	numvfs int
}

type Pfs []Pf

type Virtfns []string

func (p Pfs) Len() int      { return len(p) }
func (p Pfs) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p Pfs) Less(i, j int) bool {
	ni := uint32(p[i].Interface.HardwareAddr[3]) << 16
	ni |= uint32(p[i].Interface.HardwareAddr[4]) << 8
	ni |= uint32(p[i].Interface.HardwareAddr[5])
	nj := uint32(p[j].Interface.HardwareAddr[3]) << 16
	nj |= uint32(p[j].Interface.HardwareAddr[4]) << 8
	nj |= uint32(p[j].Interface.HardwareAddr[5])
	return ni < nj
}

func (v Virtfns) Len() int      { return len(v) }
func (v Virtfns) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v Virtfns) Less(i, j int) bool {
	ni, _ := getVfi(v[i])
	nj, _ := getVfi(v[j])
	return ni < nj
}

func (vf Vf) Port() uint    { return uint(vf >> PortShift) }
func (vf Vf) SubPort() uint { return uint((vf & (Port1 - 1)) >> SubPortShift) }
func (vf Vf) Vlan() uint    { return uint(vf & (SubPort1 - 1)) }

func (vf Vf) String() string { return VfName(vf.Port(), vf.SubPort()) }

// Machines may customize VfName for 1 based ports
var VfName = func(port, subport uint) string {
	return fmt.Sprintf("eth-%d-%d", port, subport)
}

// Machines must customize VfMac to allocate the next HardwareAddr
var VfMac = func() net.HardwareAddr { panic("FIXME") }

func Del(vfs [][]Vf) error {
	pfs, err := getPfs(len(vfs))
	if err != nil {
		return err
	}
	for _, pf := range pfs {
		if terr := setNumvfs(pf.Name, 0); terr != nil && err == nil {
			err = terr
		}
	}
	return err
}

func New(vfs [][]Vf) error {
	numvfs, err := getNumvfs()
	if err != nil {
		return err
	}
	pfs, err := getPfs(len(vfs))
	if err != nil {
		return err
	}
	for pfi, pf := range pfs {
		var virtfns Virtfns

		if pf.numvfs != numvfs {
			// First set to zero to avoid device busy error on second setNumvfs.
			if numvfs != 0 {
				if err = setNumvfs(pf.Name, 0); err != nil {
					return err
				}
			}
			if err = setNumvfs(pf.Name, numvfs); err != nil {
				return err
			}
		}

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
			err = ifset(pf.Name, "vf", vfi, "mac", VfMac(), "vlan",
				vf.Vlan())
			if err != nil {
				return err
			}
			vfname, err := getVfname(virtfn)
			if err != nil {
				return err
			}
			want := vf.String()
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
		// Setting VEPA bridge mode (instead of default VEB) for the pf allows
		// external loopback cable pings to work (while not hurting regular
		// connectivity).
		err = bridgemodeset(pf.Name, "hwmode", "vepa")
		if err != nil {
			return err
		}
	}
	return err
}

func (mac Mac) Plus(u uint) {
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

func (mac Mac) VfMac() net.HardwareAddr {
	vfmac := make(net.HardwareAddr, len(mac))
	copy(vfmac[:], mac[:])
	mac.Plus(1)
	return vfmac
}

func getPfs(numpfs int) (Pfs, error) {
	pfs := make(Pfs, 0, numpfs)
	all, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, dev := range all {
		var numvfs int
		fn := filepath.Join("/sys/class/net", dev.Name,
			"device/sriov_numvfs")
		f, terr := os.Open(fn)
		if terr != nil {
			continue
		}
		_, terr = fmt.Fscan(f, &numvfs)
		f.Close()
		if terr != nil {
			continue
		}
		if dev.Flags&net.FlagUp != net.FlagUp {
			if err = ifset(dev.Name, "up"); err != nil {
				return nil, err
			}
		}
		pfs = append(pfs, Pf{dev, numvfs})
		if pfs.Len() == numpfs {
			sort.Sort(pfs)
			return pfs, nil
		}
	}
	return nil, fmt.Errorf("have %d vs. %d pfs", pfs.Len(), numpfs)
}

func getNumvfs() (numvfs int, err error) {
	numvfs = DefaultNumvfs
	if s, _ := redis.Hget(redis.DefaultHash, "sriov.numvfs"); len(s) > 0 {
		if _, err = fmt.Sscan(s, &numvfs); err != nil {
			err = fmt.Errorf("sriov.numvfs: %v", err)
		}
	}
	return
}

func setNumvfs(ifname string, n int) error {
	fn := filepath.Join("/sys/class/net", ifname, "device/sriov_numvfs")
	f, err := os.OpenFile(fn, os.O_WRONLY|os.O_TRUNC, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err = fmt.Fprintln(f, n); err != nil {
		err = fmt.Errorf("%s: %v", fn, err)
	}
	return err
}

func getTotalvfs(ifname string) (totalvfs int, err error) {
	fn := filepath.Join("/sys/class/net", ifname, "device/sriov_totalvfs")
	f, err := os.Open(fn)
	if err != nil {
		return
	}
	defer f.Close()
	_, err = fmt.Fscan(f, &totalvfs)
	if err != nil {
		err = fmt.Errorf("%s: %v", fn, err)
	}
	return
}

func pfvirtfns(pfname string, numvfs int) (Virtfns, error) {
	var virtfns Virtfns
	pat := filepath.Join("/sys/class/net", pfname, "device/virtfn*")
	for tries := 0; true; tries++ {
		matches, err := filepath.Glob(pat)
		if err == nil && len(matches) >= numvfs {
			virtfns = Virtfns(matches)
			break
		}
		if tries == 5 {
			return nil, fmt.Errorf("%s: vf t/o", pfname)
		}
		time.Sleep(time.Second)
	}
	sort.Sort(virtfns)
	return virtfns, nil
}

func getVfi(virtfn string) (vfi int, err error) {
	s := strings.TrimPrefix(filepath.Base(virtfn), "virtfn")
	_, err = fmt.Sscan(s, &vfi)
	return
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

func bridgemodeset(name string, args ...interface{}) error {
	cmd := exec.Command("/sbin/bridge", "link", "set", "dev", name)
	for _, arg := range args {
		cmd.Args = append(cmd.Args, fmt.Sprint(arg))
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("%v: %v - %s", cmd.Args, err, output)
	}
	return err
}
