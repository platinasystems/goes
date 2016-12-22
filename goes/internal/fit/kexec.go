// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// package fit is for parsing flattened image tree binaries
package fit

import (
	"fmt"

	"github.com/platinasystems/go/goes/internal/kexec"
)

func (f *Fit) KexecLoadConfig(conf *Config, offset uintptr) (err error) {
	segments := make([]kexec.KexecSegment, 0, len(conf.ImageList))

	for _, image := range conf.ImageList {
		segments = kexec.SliceAddSegment(segments, &image.Data,
			uintptr(image.LoadAddr)+offset)
	}
	if f.Debug {
		fmt.Printf("Segments: %+v\n", segments)
	}
	err = kexec.SegmentLoad(conf.BaseAddr, &segments, uintptr(0))

	return
}
