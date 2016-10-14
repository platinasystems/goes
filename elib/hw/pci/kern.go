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
