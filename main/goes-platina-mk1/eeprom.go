// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/platinasystems/go/internal/i2c"
	"github.com/platinasystems/go/internal/optional/platina/eeprom"
)

const (
	busIndex   = 0
	busAddress = 0x51
	busDelay   = 10 * time.Millisecond
)

const MinMacs = 134

var OUI = []byte{0x02, 0x46, 0x8a}

func init() {
	eeprom.Bus.Index = busIndex
	eeprom.Bus.Address = busAddress
	eeprom.Bus.Delay = busDelay
}

func pubEeprom(pub chan<- string) error {
	buf, err := readEeprom()
	if err != nil {
		return err
	}

	p := eeprom.NewEeprom(buf)

	for _, s := range strings.Split(p.String(), "\n") {
		if len(s) > 0 {
			pub <- s
		}
	}

	if true { // FIXME false this after debug
		cerr := p.Equal(eeprom.NewEeprom(p.Bytes()))
		if cerr != nil {
			pub <- fmt.Sprintf("eeprom.verify: ", cerr)
		} else {
			pub <- "eeprom.verify: OK"
		}
	}

	n, found := p.Tlv[eeprom.NEthernetAddressType]
	if !found {
		return fmt.Errorf("eeprom: %s: not found",
			eeprom.NEthernetAddressType.String())
	}
	if n.(eeprom.Dec16) < MinMacs {
		return fmt.Errorf("%d < %d MAC addresses", n, MinMacs)
	}

	ev, found := p.Tlv[eeprom.BaseEthernetAddressType]
	if !found {
		return fmt.Errorf("eeprom: %s: not found",
			eeprom.BaseEthernetAddressType.String())
	}
	ea := ev.(*eeprom.EthernetAddress)
	if !bytes.Equal(ea[:3], OUI) {
		return fmt.Errorf("eeprom: %s: wrong O.U.I.",
			ea.String())
	}
	return nil
}

func readEeprom() ([]byte, error) {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock()

	bus, err := i2c.New(busIndex, busAddress)
	if err != nil {
		return nil, fmt.Errorf("eeprom: bus open: %v", err)
	}
	defer bus.Close()

	lbuf, err := bus.ReadBlock(eeprom.LenOffset, 2, busDelay)
	if err != nil {
		return nil, fmt.Errorf("eeprom: Read DataLen: %v", err)
	}
	n := eeprom.HeaderSz + int(binary.BigEndian.Uint16(lbuf))
	buf, err := bus.ReadBlock(0, n, busDelay)
	if err != nil {
		err = fmt.Errorf("eeprom: Read Data: %v", err)
	}
	return buf, err
}
