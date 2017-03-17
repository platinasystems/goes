// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package eeprom

import (
	"bytes"
	"fmt"
	"io"
)

type TlvMap map[Typer]interface{}

func (m TlvMap) Add(t Type) (v interface{}) {
	switch t {
	case BaseEthernetAddressType:
		v = make(EthernetAddress, 6)
	case DeviceVersionType, LabelRevisionType:
		v = new(Hex8)
	case NEthernetAddressType:
		v = new(Dec16)
	case VendorExtensionType:
		v = Vendor.New()
	case CrcType:
		v = new(Hex32)
	default:
		v = new(bytes.Buffer)
	}
	m[t] = v
	return
}

func (m TlvMap) Bytes() []byte {
	buf := new(bytes.Buffer)
	for t, v := range m {
		b := v.(Byteser).Bytes()
		buf.WriteByte(t.Byte())
		buf.WriteByte(byte(len(b)))
		buf.Write(b)
	}
	return buf.Bytes()
}

func (m TlvMap) Equal(nm TlvMap) error {
	for t, v := range m {
		nv, found := nm[t]
		if !found {
			return fmt.Errorf("missing comparable %s", t)
		}
		vByteser, found := v.(Byteser)
		if !found {
			return fmt.Errorf("%s has no Bytes()", t)
		}
		nvByteser, found := nv.(Byteser)
		if !found {
			return fmt.Errorf("comparable %s has no Bytes()", t)
		}
		b := vByteser.Bytes()
		nb := nvByteser.Bytes()
		if !bytes.Equal(b, nb) {
			return fmt.Errorf("%s:\n%#0x vs.\n%#0x", t, nb, b)
		}
	}
	return nil
}

func (m TlvMap) String() string {
	buf := new(bytes.Buffer)
	for t, v := range m {
		_, isBytesBuffer := v.(*bytes.Buffer)
		if t != VendorExtensionType || isBytesBuffer {
			fmt.Fprint(buf, "eeprom.", t, ": ", v, "\n")
		} else {
			buf.WriteString(v.(fmt.Stringer).String())
		}
	}
	return buf.String()
}

// Write buf into TlvMap
func (m TlvMap) Write(buf []byte) (n int, err error) {
	for len(buf) > 2 && err == nil {
		t := Type(buf[0])
		l := int(buf[1])
		buf = buf[2:]
		n += 2
		if l == 0 {
			continue
		}
		if l > len(buf) {
			l = len(buf)
		}
		n += l
		v := m[t]
		if v == nil {
			v = m.Add(t)
		}
		if method, found := v.(Reseter); found {
			method.Reset()
		}
		if method, found := v.(io.Writer); found {
			_, err = method.Write(buf[:l])
		} else {
			err = fmt.Errorf("%s(%T): missing Writer", t, v)
		}
		buf = buf[l:]
	}
	return
}
