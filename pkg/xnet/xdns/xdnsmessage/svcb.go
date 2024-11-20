// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net/netip"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"golang.org/x/net/dns/dnsmessage"
)

//go:generate stringer -type SVCBKey -trimprefix SVCBKey -linecomment
type SVCBKey uint16

const (
	SVCBKeyMandatory SVCBKey = iota // mandatory

	SVCBKeyALPN          // alpn
	SVCBKeyNoDefaultALPN // no-default-alpn
	SVCBKeyPort          // port
	SVCBKeyIPV4Hint      // ipv4hint
	SVCBKeyECH           // ech
	SVCBKeyIPV6Hint      // ipv6hint
	SVCBKeyDOHPath       // doh-path
)

// https://datatracker.ietf.org/doc/rfc9460/
type TypeSVCBResource struct {
	Pri    uint16
	Name   UniqueString
	Params []Param
}

type TypeHTTPSResource struct{ TypeSVCBResource }
type TypeCAAResource struct{ TypeSVCBResource }

func (v TypeSVCBResource) construct(
	mb *dnsmessage.Builder, h dnsmessage.ResourceHeader,
) error {
	r := dnsmessage.UnknownResource{
		Type: h.Type,
	}
	r.Data = binary.BigEndian.AppendUint16(r.Data, v.Pri)
	r.Data = v.Name.AppendTo(r.Data)
	for _, param := range v.Params {
		_ = param // FIXME r.Data = param.AppendTo(r.Data)
	}
	return mb.UnknownResource(h, r)
}

func (svcb TypeSVCBResource) String() string {
	w := new(strings.Builder)
	fmt.Fprintf(w, "%d", svcb.Pri)
	fmt.Fprint(w, " ", svcb.Name)
	for _, param := range svcb.Params {
		fmt.Fprint(w, " ", param)
	}
	return w.String()
}

type SVCBALPNS []string

func (alpns SVCBALPNS) String() string {
	return fmt.Sprintf("%q", strings.Join(alpns, ","))
}

type SVCBPort uint16

func (port SVCBPort) String() string {
	return fmt.Sprintf("%d", uint16(port))
}

func (svcb *TypeSVCBResource) UnmarshalBinary(msg []byte) error {
	var err error
	if len(msg) < 2 {
		return xerrors.Incomplete("SVCB")
	}
	svcb.Pri = binary.BigEndian.Uint16(msg)
	msg = msg[2:]
	msg, err = svcb.Name.Decode(msg)
	if err != nil {
		return err
	}
	if svcb.Pri == 0 {
		// Alias mode [rfc9460 section 2.4.2]
		return nil
	}
	for len(msg) >= 4 {
		k := SVCBKey(binary.BigEndian.Uint16(msg))
		param := Param{Key: k}
		msg = msg[2:]
		l := int(binary.BigEndian.Uint16(msg))
		msg = msg[2:]
		switch param.Key {
		case SVCBKeyMandatory:
			// FIXME txt
		case SVCBKeyALPN:
			var alpns SVCBALPNS
			for i, n := 0, 0; i < l; i += 1 + n {
				n = int(msg[i])
				if 1+n > l {
					return xerrors.Underrun("SVCB-ALPN")
				}
				alpn := string(bytes.Clone(msg[i+1 : i+1+n]))
				alpns = append(alpns, alpn)
			}
			param.Value = alpns
		case SVCBKeyNoDefaultALPN:
			// empty
		case SVCBKeyPort:
			param.Value = SVCBPort(binary.BigEndian.Uint16(msg))
		case SVCBKeyIPV4Hint:
			if len(msg) < 4 {
				return xerrors.Underrun("SVCB-IPV4HINT")
			}
			param.Value, _ = netip.AddrFromSlice(msg[:4])
		case SVCBKeyECH:
			return xerrors.FIXME("SVCB-ECH: decode base64")
		case SVCBKeyIPV6Hint:
			if len(msg) < 16 {
				return xerrors.Underrun("SVCB-IPV6HINT")
			}
			param.Value, _ = netip.AddrFromSlice(msg[:16])
		case SVCBKeyDOHPath:
			// FIXME decode DOH path
		default:
			// FIXME param.Value = msg[:l]
		}
		svcb.Params = append(svcb.Params, param)
		msg = msg[l:]
	}
	return nil
}

func (r *TypeHTTPSResource) UnmarshalBinary(msg []byte) error {
	return r.TypeSVCBResource.UnmarshalBinary(msg)
}

func (r *TypeCAAResource) UnmarshalBinary(msg []byte) error {
	return r.TypeSVCBResource.UnmarshalBinary(msg)
}
