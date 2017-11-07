// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mmclog

import (
	//"fmt"
	//"strconv"

	"github.com/platinasystems/go/goes/cmd/platina/mk2/lc1-bmc/mmclogd"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "mmclog"
	Apropos = "MMC persistant dmesg log status"
	Usage   = "mmclog"
	Man     = `
DESCRIPTION
        The mmclog command shows status of MMC dmesg logs`
)

type Interface interface {
	Apropos() lang.Alt
	Main(...string) error
	Man() lang.Alt
	String() string
	Usage() string
}

type cmd struct{}

func New() Interface { return cmd{} }

func (cmd) Apropos() lang.Alt { return apropos }
func (cmd) Man() lang.Alt     { return man }
func (cmd) String() string    { return Name }
func (cmd) Usage() string     { return Usage }

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)

func (cmd) Main(args ...string) (err error) {
	/*var a uint64 = 0
	if len(args) != 0 {
		a, err = strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			return fmt.Errorf("%s: %v", args[0], err)
		}
	}

	switch a {
	case 1:
		return mmclogd.LogDmesg(1)
	case 2:
		return mmclogd.LogDmesg(100000)
	case 3:
		return mmclogd.StartTicker()
	case 4:
		return mmclogd.StopTicker()
	case 5:
		return mmclogd.MountMMC()
	case 6:
		return mmclogd.ListMMC()
	case 7:
		return mmclogd.InitLogging()
	case 8:
		return mmclogd.GetDmesgInfo()
	case 9:
		return mmclogd.ReadFile()
	default:
		return mmclogd.Status()
	return nil
	}*/

	return mmclogd.Status()
}
