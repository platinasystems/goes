// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package platina_eeprom

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/platinasystems/goes/cmd/eeprom"
	"github.com/platinasystems/goes/external/i2c"
	"github.com/platinasystems/goes/external/redis/publisher"
)

func RedisdHook(pub *publisher.Publisher) {
	var p eeprom.Eeprom

	buf, err := readbytes()
	if err != nil {
		panic(err)
	}

	_, err = p.Write(buf)
	if err != nil {
		panic(err)
	}

	for _, s := range strings.Split(p.String(), "\n") {
		if len(s) > 0 {
			pub.Write([]byte(s))
		}
	}

	if config.minMacs > 0 {
		v, found := p.Tlv[eeprom.NEthernetAddressType]
		if !found {
			fmt.Printf("eeprom: %s: not found",
				eeprom.NEthernetAddressType)
		} else if n := int(*v.(*eeprom.Dec16)); n < config.minMacs {
			fmt.Printf("%d < %d MAC addresses",
				n, config.minMacs)
		}
	}

	if !bytes.Equal(config.oui[:], []byte{0, 0, 0}) {
		ev, found := p.Tlv[eeprom.BaseEthernetAddressType]
		if !found {
			fmt.Printf("eeprom: %s: not found",
				eeprom.BaseEthernetAddressType.String())
		} else {
			// all non-blank MAC addresses are allowed
			ea := ev.(eeprom.EthernetAddress)
			if bytes.Equal(ea[:], []byte{0, 0, 0, 0, 0, 0}) {
				fmt.Printf("eeprom: zero MAC BASE")
			}
		}
	}
}

func readbytes() ([]byte, error) {
	// eeprom reads are called early, by redis hook in start
	// i2cd is not up in start, so direct i2c calls are used
	var (
		bus  *i2c.Bus
		err  error
		lbuf []byte
	)
	// Try possible addresses one by one
	for _, address := range config.bus.addresses {
		fmt.Printf("try eeprom address 0x%x...", address)
		if bus, err = i2c.New(config.bus.index, address); err == nil {
			if lbuf, err = bus.ReadBlock(eeprom.LenOffset, 2, config.bus.delay); err == nil {
				fmt.Printf("success\n")
				break
			}
		}
		fmt.Printf("%v\n", err)
	}
	if err != nil {
		return nil, fmt.Errorf("eeprom: %v", err)
	}
	defer bus.Close()

	n := eeprom.HeaderSz + int(binary.BigEndian.Uint16(lbuf))
	buf, err := bus.ReadBlock(0, n, config.bus.delay)
	if err != nil {
		err = fmt.Errorf("eeprom: Read Data: %v", err)
	}
	return buf, err
}
