// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !uio_pci_dma

package pci

type wrapperDevice struct {
	Device
}

func NewDevice() Devicer {
	d := &wrapperDevice{}
	d.Devicer = d
	return d
}

func (d *wrapperDevice) GetDevice() *Device { return &d.Device }
func (d *wrapperDevice) Open() error        { return nil }
