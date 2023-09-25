// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// package fit is for parsing flattened image tree binaries
package fit

import (
	"github.com/platinasystems/goes/internal/kexec"
)

func (f *Fit) KexecLoadConfig(conf *Config, cmdline string) (err error) {
	kdat := []byte{}
	idat := []byte{}

	for _, image := range conf.ImageList {
		if image.Type == "kernel" {
			kdat = image.Data
		}

		if image.Type == "ramdisk" {
			idat = image.Data
		}
	}

	return kexec.LoadSlices(kdat, idat, cmdline, 0)
}
