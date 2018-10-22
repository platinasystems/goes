// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package platina_eeprom

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/platinasystems/go/goes/cmd/eeprom"
	"github.com/platinasystems/i2c"
	"github.com/platinasystems/redis/publisher"
)

func RedisdHook(pub *publisher.Publisher) {
	buf, err := readbytes()
	if pub.Error(err) != nil {
		return
	}

	var p eeprom.Eeprom

	_, err = p.Write(buf)
	if pub.Error(err) != nil {
		return
	}

	for _, s := range strings.Split(p.String(), "\n") {
		if len(s) > 0 {
			pub.WriteString(s)
		}
	}

	if config.minMacs > 0 {
		v, found := p.Tlv[eeprom.NEthernetAddressType]
		if !found {
			pub.Error(fmt.Errorf("eeprom: %s: not found",
				eeprom.NEthernetAddressType))
		} else if n := int(*v.(*eeprom.Dec16)); n < config.minMacs {
			pub.Error(fmt.Errorf("%d < %d MAC addresses",
				n, config.minMacs))
		}
	}

	if !bytes.Equal(config.oui[:], []byte{0, 0, 0}) {
		ev, found := p.Tlv[eeprom.BaseEthernetAddressType]
		if !found {
			pub.Error(fmt.Errorf("eeprom: %s: not found",
				eeprom.BaseEthernetAddressType.String()))
		} else {
			// all non-blank MAC addresses are allowed
			ea := ev.(eeprom.EthernetAddress)
			if bytes.Equal(ea[:], []byte{0, 0, 0, 0, 0, 0}) {
				pub.Error(fmt.Errorf("eeprom: zero MAC BASE"))
			}
		}
	}
}

func readbytes() ([]byte, error) {
	// eeprom reads are called early, by redis hook in start
	// i2cd is not up in start, so direct i2c calls are used
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
