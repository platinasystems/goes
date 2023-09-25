// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package platina_eeprom

import "fmt"

func ReadBytes() ([]byte, error) {
	return readbytes()
}

func Write([]byte) (int, error) {
	return 0, fmt.Errorf("FIXME")
}
