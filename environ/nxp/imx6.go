// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package nxp provides access to the NXP iMX6 ARM CPU
package imx6

import (
	"fmt"
	"io/ioutil"
	"strconv"
)

type Cpu struct {
}

func (h *Cpu) ReadTemp() float64 {
	tmp, _ := ioutil.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	tmp2 := fmt.Sprintf("%.4s", string(tmp[:]))
	tmp3, _ := strconv.Atoi(tmp2)
	tmp4 := float64(tmp3)
	return float64(tmp4 / 100.0)
}
