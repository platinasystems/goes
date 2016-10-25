// package fit is for parsing flattened image tree binaries
package fit

import (
	"fmt"
	"github.com/platinasystems/go/kexec"
)

func (f *Fit) KexecLoadConfig(conf *Config, offset uintptr) (err error) {
	var segments []kexec.KexecSegment

	segments = make([]kexec.KexecSegment, len(conf.ImageList),
		len(conf.ImageList))

	for i, image := range conf.ImageList {
		segments[i].Buf = &image.Data[0]
		bufsz := uint(len(image.Data))
		segments[i].Bufsz = bufsz
		memsz := (bufsz + 4095) &^ 4095
		segments[i].Mem = uintptr(image.LoadAddr) + offset
		segments[i].Memsz = memsz
	}
	if f.Debug {
		fmt.Printf("Segments: %+v\n", segments)
	}
	err = kexec.SegmentLoad(conf.BaseAddr, &segments, uintptr(0))

	return
}
