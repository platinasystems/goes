// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build darwin

package uptime

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xexec"
	"github.com/platinasystems/goes/v2/pkg/xos/sysctl"
)

const (
	Day    = 24 * time.Hour
	FSHIFT = 11
)

type Sysctl struct {
	Name string
	MIB  []int32
}

var (
	KernBoottime = Sysctl{"kern.boottime", []int32{1, 21}}
	VmLoadavg    = Sysctl{"vm.loadavg", []int32{2, 2}}
)

func uptime() error {
	bkbt, err := sysctl.Get(KernBoottime.MIB...)
	if err != nil {
		return err
	}
	if len(bkbt) < 8+4 {
		return xerrors.Invalid(KernBoottime.Name)
	}
	blav, err := sysctl.Get(VmLoadavg.MIB...)
	if err != nil {
		return err
	}
	if len(blav) < (4 * +4) {
		return xerrors.Invalid(VmLoadavg.Name)
	}

	now := time.Now()
	nusers := xexec.NUsers() - 1 // Why -1?

	upSec := binary.NativeEndian.Uint64(bkbt)
	upUsec := binary.NativeEndian.Uint32(bkbt[8:])
	upTime := time.Unix(int64(upSec), int64(upUsec)*1000)

	m1 := float64(binary.NativeEndian.Uint32(blav))
	m5 := float64(binary.NativeEndian.Uint32(blav[4:]))
	m15 := float64(binary.NativeEndian.Uint32(blav[8:]))
	fscale := float64(binary.NativeEndian.Uint32(blav[12:]))
	if fscale == 0 {
		fscale = 1 << FSHIFT
	}
	m1 /= fscale
	m5 /= fscale
	m15 /= fscale

	nowHour, nowMin, _ := now.Clock()
	fmt.Printf("%2d:%02d  up", nowHour, nowMin)

	dur := now.Sub(upTime)
	if days := uint(dur / Day); days != 0 {
		fmt.Print(" ", days, " days,")
		dur -= time.Duration(days) * Day
	}

	dur = dur.Round(time.Minute)
	h := uint(dur.Hours())
	dur -= time.Duration(h) * time.Hour
	m := uint(dur.Minutes())

	fmt.Printf(" %d:%d, %d users, load averages: %1.2f %1.2f %1.2f\n",
		h, m, nusers, m1, m5, m15)
	return nil
}
