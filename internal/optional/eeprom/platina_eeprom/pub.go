// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package platina_eeprom

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/platinasystems/go/internal/i2c"
	"github.com/platinasystems/go/internal/optional/eeprom"
)

func RedisdHook(pub chan<- string) error {
	buf, err := readbytes()
	if err != nil {
		return err
	}

	var p eeprom.Eeprom

	_, err = p.Write(buf)
	if err != nil {
		return err
	}

	for _, s := range strings.Split(p.String(), "\n") {
		if len(s) > 0 {
			pub <- s
		}
	}

	if config.minMacs > 0 {
		v, found := p.Tlv[eeprom.NEthernetAddressType]
		if !found {
			return fmt.Errorf("eeprom: %s: not found",
				eeprom.NEthernetAddressType)
		}
		n := int(*v.(*eeprom.Dec16))
		if n < config.minMacs {
			return fmt.Errorf("%d < %d MAC addresses",
				n, config.minMacs)
		}
	}

	if !bytes.Equal(config.oui[:], []byte{0, 0, 0}) {
		ev, found := p.Tlv[eeprom.BaseEthernetAddressType]
		if !found {
			return fmt.Errorf("eeprom: %s: not found",
				eeprom.BaseEthernetAddressType.String())
		}

		// all non-blank MAC addresses are allowed
		ea := ev.(*eeprom.EthernetAddress)
		if bytes.Equal(ea[:], []byte{0, 0, 0, 0, 0, 0}) {
			return fmt.Errorf("eeprom: invalid MAC BASE: %0x", ea[:6])
		}
	}
	return nil
}

func readbytes() ([]byte, error) {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock()

	bus, err := i2c.New(config.bus.index, config.bus.address)
	if err != nil {
		return nil, fmt.Errorf("eeprom: bus open: %v", err)
	}
	defer bus.Close()

	lbuf, err := bus.ReadBlock(eeprom.LenOffset, 2, config.bus.delay)
	if err != nil {
		return nil, fmt.Errorf("eeprom: Read DataLen: %v", err)
	}
	n := eeprom.HeaderSz + int(binary.BigEndian.Uint16(lbuf))
	buf, err := bus.ReadBlock(0, n, config.bus.delay)
	if err != nil {
		err = fmt.Errorf("eeprom: Read Data: %v", err)
	}
	return buf, err
}
