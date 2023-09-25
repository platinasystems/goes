// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package fit

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestFit(t *testing.T) {
	itb, err := ioutil.ReadFile("testdata/platina-mk1-bmc-itb.bin")
	if err != nil {
		t.Fatal(err)
	}

	vmlinuz, err := ioutil.ReadFile("testdata/platina-mk1-bmc.vmlinuz")
	if err != nil {
		t.Fatal(err)
	}

	initramfs, err := ioutil.ReadFile("testdata/platina-mk1-bmc.cpio.xz")
	if err != nil {
		t.Fatal(err)
	}

	dtb, err := ioutil.ReadFile("testdata/platina-mk1-bmc-dtb.bin")
	if err != nil {
		t.Fatal(err)
	}

	fit := Parse(itb)

	config := fit.Configs[fit.DefaultConfig]

	if config == nil {
		t.Fatal("No default configuration")
	}

	for _, image := range config.ImageList {
		if image.Type == "kernel" {
			fmt.Printf("Found kernel\n")
			if !bytes.Equal(image.Data, vmlinuz) {
				fmt.Printf("  data mismatch")
			}
		}

		if image.Type == "ramdisk" {
			fmt.Printf("Found ramdisk\n")
			if !bytes.Equal(image.Data, initramfs) {
				fmt.Printf("  data mismatch")
			}
		}

		if image.Type == "flat_dt" {
			fmt.Printf("Found dtb\n")
			if !bytes.Equal(image.Data, dtb) {
				fmt.Printf("  data mismatch")
			}
		}
	}
}
