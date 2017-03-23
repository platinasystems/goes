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
)

const DefaultNumvfs = 16

type Mac []byte
type Vf uint

func (vf Vf) Port() uint    { return uint((vf &^ (Port(1) - 1)) >> 20) }
func (vf Vf) SubPort() uint { return uint((vf & 0xf0000) >> 16) }
func (vf Vf) Vlan() uint    { return uint(vf & 0xffff) }

func Port(u uint) Vf    { return Vf(u << 20) }
func SubPort(u uint) Vf { return Vf((u & 0xf) << 16) }
func Vlan(u uint) Vf    { return Vf(u & 0xffff) }

func Mksriovs(porto uint, vfs ...[]Vf) error {
	mac := make(Mac, 6)
	err := assert.Root()
	if err != nil {
		return err
	}
	numvfsFns, err := filepath.Glob("/sys/class/net/*/device/sriov_numvfs")
	if err != nil {
		return err
	}
	for pfi, numvfsFn := range numvfsFns {
		var numvfs, totalvfs uint
		var virtfns []string
		if pfi > len(vfs) {
			break
		}
		classNetPf := strings.TrimSuffix(numvfsFn,
			"/device/sriov_numvfs")
		pfname := filepath.Base(classNetPf)
		if _, err = FnScan(numvfsFn, &numvfs); err != nil {
			return fmt.Errorf("%s: numvfs: %v", pfname, err)
		}
		s := strings.Replace(numvfsFn, "numvfs", "totalvfs", 1)
		if _, err = FnScan(s, &totalvfs); err != nil {
			return fmt.Errorf("%s: totalvfs: %v", pfname, err)
		}

		if !strings.HasPrefix(pfname, "pf") {
			newname := fmt.Sprint("pf", pfi)
			cmd := exec.Command("ip", "link", "set", pfname,
				"name", newname)
			if err = cmd.Run(); err != nil {
				return fmt.Errorf("%v: %v", cmd.Args, err)
			}
			pfname = newname
			numvfsFn = filepath.Join("/sys/class/net", pfname, "device/sriov_numvfs")
		}

		pfdev, err := net.InterfaceByName(pfname)
		if err != nil {
			return err
		}
		if pfdev.Flags&net.FlagUp != net.FlagUp {
			cmd := exec.Command("ip", "link", "set", pfname, "up")
			if err = cmd.Run(); err != nil {
				return fmt.Errorf("%v: %v", cmd.Args, err)
			}
		}

		copy(mac, pfdev.HardwareAddr)
		mac.Plus(uint(len(numvfsFns)-pfi) + (uint(pfi) * totalvfs))

		s = filepath.Join("/sys/class/net", pfname, "device/virtfn*")
		if numvfs == 0 {
			numvfs = DefaultNumvfs
			if n := uint(len(vfs[pfi])); n < numvfs {
				numvfs = n
			}
			if err = FnPrintln(numvfsFn, numvfs); err != nil {
				return fmt.Errorf("set %s: %v", numvfsFn, err)
			}
			for tries := 0; true; tries++ {
				virtfns, err = filepath.Glob(s)
				if err == nil && uint(len(virtfns)) == numvfs {
					break
				}
				if tries == 5 {
					return fmt.Errorf("%s: vf t/o", pfname)
				}
				time.Sleep(time.Second)
			}
		} else if virtfns, err = filepath.Glob(s); err != nil {
			return err
		}

		for _, virtfn := range virtfns {
			var vfi uint
			s = strings.TrimPrefix(filepath.Base(virtfn), "virtfn")
			if _, err = fmt.Sscan(s, &vfi); err != nil {
				return err
			}
			if vfi > uint(len(vfs[pfi])) {
				continue
			}

			vf := vfs[pfi][vfi]
			vfname := fmt.Sprintf("eth-%d-%d", vf.Port()+porto,
				vf.SubPort())
			cmd := exec.Command("ip", "link", "set", pfname,
				"vf", fmt.Sprint(vfi),
				"mac", mac.String(),
				"vlan", fmt.Sprint(vf.Vlan()))
			if err = cmd.Run(); err != nil {
				return fmt.Errorf("%v: %v", cmd.Args, err)
			}
			mac.Plus(1)

			match, err := filepath.Glob(filepath.Join(virtfn,
				"net/*"))
			if err != nil {
				return fmt.Errorf("glob %s/net*: %v",
					virtfn, err)
			}
			if name := filepath.Base(match[0]); name != vfname {
				cmd = exec.Command("ip", "link", "set", name,
					"name", vfname)
				if err = cmd.Run(); err != nil {
					return fmt.Errorf("%v: %v", cmd.Args,
						err)
				}
			}

			// bounce the vf to reload its mac from the pf
			cmd = exec.Command("ip", "link", "set", vfname, "up")
			if err = cmd.Run(); err != nil {
				return fmt.Errorf("%v: %v", cmd.Args, err)
			}
			cmd = exec.Command("ip", "link", "set", vfname, "down")
			if err = cmd.Run(); err != nil {
				return fmt.Errorf("%v: %v", cmd.Args, err)
			}
		}

	}
	return nil
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

func (mac Mac) String() string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%2x",
		mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
}

func FnPrintln(fn string, values ...interface{}) error {
	f, err := os.OpenFile(fn, os.O_WRONLY|os.O_TRUNC, 0)
	if err == nil {
		defer f.Close()
		_, err = fmt.Fprintln(f, values...)
	}
	return err
}

func FnScan(fn string, a ...interface{}) (n int, err error) {
	b, err := ioutil.ReadFile(fn)
	if err == nil {
		n, err = fmt.Sscan(string(b), a...)
	}
	return
}
