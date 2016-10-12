// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package diagutils/ipmigod provides an IPMI protocol daemon.
package ipmigod

import (
	"strconv"

	"github.com/platinasystems/go/flags"
	"github.com/platinasystems/go/ipmigod"
	"github.com/platinasystems/go/parms"
	"github.com/platinasystems/go/redis"
)

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return "ipmigod" }
func (cmd) Usage() string  { return "ipmigod [OPTIONS]..." }

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "ipmigod daemon",
	}
}

func (cmd) Daemon() int { return 1 }

func (cmd) Main(args ...string) error {
	flag, args := flags.New(args, "-mm")
	parm, args := parms.New(args, "-lc")
	for k, v := range map[string]string{
		"-lc": "0",
	} {
		if len(parm[k]) == 0 {
			parm[k] = v
		}
	}

	cardSlot, err := redis.Hget("cmdline", "card/slot")
	if err != nil {
		return err
	}
	cardType, err := redis.Hget("cmdline", "card/type")
	if err != nil {
		return err
	}

	mmCard := flag["-mm"] || cardType == "management"

	cardNum, err := strconv.ParseInt(parm["-lc"], 0, 0)
	if err != nil {
		return err
	}

	if cardNum == 0 && len(cardSlot) > 0 {
		cardNum, err = strconv.ParseInt(cardSlot, 0, 0)
		if err != nil {
			return err
		}
	}

	ipmigod.Ipmigod(mmCard, int(cardNum))
	return nil
}
