// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package eeprom

// Each machine main must assign the following Vendor parameters
var Vendor struct {
	New       func() VendorExtension
	ReadBytes func() ([]byte, error)
	Write     func([]byte) (int, error)
}
