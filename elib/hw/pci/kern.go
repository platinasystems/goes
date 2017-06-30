// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !uio_pci_dma,!vfio

package pci

type wrapperDevice struct {
	Device
}

type wrapperBus struct {
	busCommon
}

var DefaultBus = &wrapperBus{}

func (wrapperBus) NewDevice() BusDevice  { return &wrapperDevice{} }
func (wrapperBus) Validate() (err error) { return }

func (d *wrapperDevice) GetDevice() *Device { return &d.Device }
func (d *wrapperDevice) Open() error        { return nil }
