// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package sleep

import (
	"fmt"
	"strconv"
	"time"
)

type sleep struct{}

func New() sleep { return sleep{} }

func (sleep) String() string { return "sleep" }
func (sleep) Usage() string  { return "sleep SECONDS" }

func (sleep) Main(args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("SECONDS: missing")
	}

	t, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		return fmt.Errorf("%s: %v", args[0], err)
	}

	time.Sleep(time.Second * time.Duration(t))
	return nil
}

func (sleep) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "suspend execution for an interval of time",
	}
}

func (sleep) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	sleep - suspend execution for an interval of time

SYNOPSIS
	sleep SECONDS

DESCRIPTION
	The sleep command suspends execution for a number of SECONDS.`,
	}
}
