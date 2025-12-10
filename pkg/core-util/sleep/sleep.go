// Copyright © 2015-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package sleep

import (
	"context"
	"flag"
	"strconv"
	"strings"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
)

const SleepUsage = `
usage: {{.Name}} <number[unit]>
Suspend execution for time interval.

Units
  ns	Nanoseconds
  us	Microseconds
  ms	Milliseconds
  s	Seconds (default)
  m	Minutes
  d	Hours
  d	Days
`

func Sleep(ctx context.Context, args []string) error {
	xflag.TemplateUsage(SleepUsage)
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	args = flag.Args()
	if len(args) == 0 {
		return xerrors.Incomplete("interval")
	}
	s := args[0]
	unit := time.Second
	if strings.HasSuffix(s, "ns") {
		s = strings.TrimSuffix(s, "ns")
		unit = time.Nanosecond
	} else if strings.HasSuffix(s, "us") {
		s = strings.TrimSuffix(s, "us")
		unit = time.Microsecond
	} else if strings.HasSuffix(s, "ms") {
		s = strings.TrimSuffix(s, "ms")
		unit = time.Millisecond
	} else if strings.HasSuffix(s, "s") {
		s = strings.TrimSuffix(s, "s")
		unit = time.Second
	} else if strings.HasSuffix(s, "m") {
		s = strings.TrimSuffix(s, "m")
		unit = time.Minute
	} else if strings.HasSuffix(s, "h") {
		s = strings.TrimSuffix(s, "h")
		unit = time.Hour
	} else if strings.HasSuffix(s, "d") {
		s = strings.TrimSuffix(s, "d")
		unit = 24 * time.Hour
	}

	t, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case <-time.After(time.Duration(t) * unit):
	}
	return err
}
