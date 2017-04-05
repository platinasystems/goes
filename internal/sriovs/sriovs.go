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

	"github.com/platinasystems/go/internal/assert"
)

const DefaultNumvfs = 16

type MacByIfindex map[int][6]byte
type Mac [6]byte
type Vf uint

func (vf Vf) Port() uint    { return uint((vf &^ (Port(1) - 1)) >> 20) }
func (vf Vf) SubPort() uint { return uint((vf & 0xf0000) >> 16) }
func (vf Vf) Vlan() uint    { return uint(vf & 0xffff) }

func Port(u uint) Vf    { return Vf(u << 20) }
func SubPort(u uint) Vf { return Vf((u & 0xf) << 16) }
func Vlan(u uint) Vf    { return Vf(u & 0xffff) }

func Mksriovs(porto uint, vfs ...[]Vf) (MacByIfindex, error) {
	var mac Mac
	err := assert.Root()
	if err != nil {
		return nil, err
	}
	numvfsFns, err := NumvfsFns()
	if err != nil {
		return nil, err
	}
	macByIfindex := make(MacByIfindex)
pfloop:
	for pfi, numvfsFn := range numvfsFns {
		var numvfs, totalvfs uint
		var virtfns []string
		if pfi > len(vfs) {
			break pfloop
		}

		pfname := filepath.Base(filepath.Dir(filepath.Dir(numvfsFn)))
		if !strings.HasPrefix(pfname, "pf") {
			newname := fmt.Sprint("pf", pfi)
			cmd := exec.Command("ip", "link", "set", pfname,
				"name", newname)
			if err = cmd.Run(); err != nil {
				err = fmt.Errorf("%v: %v", cmd.Args, err)
				break pfloop
			}
			numvfsFn = filepath.Join("/sys/class/net", newname,
				"device/sriov_numvfs")
			pfname = newname
		}

		if _, err = FnScan(numvfsFn, &numvfs); err != nil {
			err = fmt.Errorf("%s: numvfs: %v", pfname, err)
			break pfloop
		}

		totalvfsFn := filepath.Join(filepath.Dir(numvfsFn),
			"sriov_totalvfs")
		if _, err = FnScan(totalvfsFn, &totalvfs); err != nil {
			err = fmt.Errorf("%s: totalvfs: %v", pfname, err)
			break pfloop
		}

		var pfdev *net.Interface
		pfdev, err = net.InterfaceByName(pfname)
		if err != nil {
			break pfloop
		}
		if pfdev.Flags&net.FlagUp != net.FlagUp {
			cmd := exec.Command("ip", "link", "set", pfname, "up")
			if err = cmd.Run(); err != nil {
				err = fmt.Errorf("%v: %v", cmd.Args, err)
				break pfloop
			}
		}

		copy(mac[:], pfdev.HardwareAddr)
		mac.Plus(uint(len(numvfsFns)-pfi) + (uint(pfi) * totalvfs))

		virtfnPat := filepath.Join(filepath.Dir(numvfsFn), "virtfn*")
		if numvfs == 0 {
			numvfs = DefaultNumvfs
			if s := os.Getenv("NUMVFS"); len(s) > 0 {
				_, err = fmt.Sscan(s, &numvfs)
				if err != nil {
					err = fmt.Errorf("NUMVFS: %v", err)
					break pfloop
				}
			}
			if n := uint(len(vfs[pfi])); n < numvfs {
				numvfs = n
			}
			if err = FnPrintln(numvfsFn, numvfs); err != nil {
				err = fmt.Errorf("set %s: %v", numvfsFn, err)
				break pfloop
			}
			for tries := 0; true; tries++ {
				virtfns, err = filepath.Glob(virtfnPat)
				if err == nil && uint(len(virtfns)) == numvfs {
					break
				}
				if tries == 5 {
					err = fmt.Errorf("%s: vf t/o", pfname)
					break pfloop
				}
				time.Sleep(time.Second)
			}
		} else if virtfns, err = filepath.Glob(virtfnPat); err != nil {
			break pfloop
		}

		for _, virtfn := range virtfns {
			var vfi uint
			base := filepath.Base(virtfn)
			svfi := strings.TrimPrefix(base, "virtfn")
			if _, err = fmt.Sscan(svfi, &vfi); err != nil {
				break pfloop
			}
			if vfi >= uint(len(vfs[pfi])) {
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
				err = fmt.Errorf("%v: %v", cmd.Args, err)
				break pfloop
			}

			var vfdev *net.Interface
			vfdev, err = net.InterfaceByName(vfname)
			if err != nil {
				break pfloop
			}
			macByIfindex[vfdev.Index] = mac
			mac.Plus(1)

			var match []string
			match, err = filepath.Glob(filepath.Join(virtfn,
				"net/*"))
			if err != nil {
				err = fmt.Errorf("glob %s/net*: %v", virtfn,
					err)
				break pfloop
			}
			if len(match) == 0 {
				err = fmt.Errorf("%s has no virtfns", pfname)
				break pfloop
			}
			if name := filepath.Base(match[0]); name != vfname {
				cmd = exec.Command("ip", "link", "set", name,
					"name", vfname)
				if err = cmd.Run(); err != nil {
					err = fmt.Errorf("%v: %v", cmd.Args,
						err)
					break pfloop
				}
			}

			// bounce the vf to reload its mac from the pf
			cmd = exec.Command("ip", "link", "set", vfname, "up")
			if err = cmd.Run(); err != nil {
				err = fmt.Errorf("%v: %v", cmd.Args, err)
				break pfloop
			}
			cmd = exec.Command("ip", "link", "set", vfname, "down")
			if err = cmd.Run(); err != nil {
				err = fmt.Errorf("%v: %v", cmd.Args, err)
				break pfloop
			}
		}
	}

	if false {
		idxs := make([]int, 0, len(macByIfindex))
		for idx := range macByIfindex {
			idxs = append(idxs, idx)
		}
		sort.Ints(idxs)
		for _, idx := range idxs {
			fmt.Print("if[", idx, "].mac: ",
				Mac(macByIfindex[idx]), "\n")
		}
	}

	return macByIfindex, err
}

func (macs MacByIfindex) Cast() map[int][6]byte {
	return map[int][6]byte(macs)
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
