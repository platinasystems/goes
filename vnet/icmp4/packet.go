// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package icmp4

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ip"

	"fmt"
	"unsafe"
)

type Header struct {
	Type     Type
	Code     uint8
	Checksum vnet.Uint16
}

const HeaderBytes = 4

type Type uint8

const (
	Echo_reply                  = 0
	Destination_unreachable     = 3
	Source_quench               = 4
	Redirect                    = 5
	Alternate_host_address      = 6
	Echo_request                = 8
	Router_advertisement        = 9
	Router_solicitation         = 10
	Time_exceeded               = 11
	Parameter_problem           = 12
	Timestamp_request           = 13
	Timestamp_reply             = 14
	Information_request         = 15
	Information_reply           = 16
	Address_mask_request        = 17
	Address_mask_reply          = 18
	Traceroute                  = 30
	Datagram_conversion_error   = 31
	Mobile_host_redirect        = 32
	Ip6_where_are_you           = 33
	Ip6_i_am_here               = 34
	Mobile_registration_request = 35
	Mobile_registration_reply   = 36
	Domain_name_request         = 37
	Domain_name_reply           = 38
	Skip                        = 39
	Photuris                    = 40
)

var typeStrings = [...]string{
	Echo_reply:                  "echo-reply",
	Destination_unreachable:     "destination-unreachable",
	Source_quench:               "source-quench",
	Redirect:                    "redirect",
	Alternate_host_address:      "alternate-host-address",
	Echo_request:                "echo-request",
	Router_advertisement:        "router-advertisement",
	Router_solicitation:         "router-solicitation",
	Time_exceeded:               "time-exceeded",
	Parameter_problem:           "parameter-problem",
	Timestamp_request:           "timestamp-request",
	Timestamp_reply:             "timestamp-reply",
	Information_request:         "information-request",
	Information_reply:           "information-reply",
	Address_mask_request:        "address-mask-request",
	Address_mask_reply:          "address-mask-reply",
	Traceroute:                  "traceroute",
	Datagram_conversion_error:   "datagram-conversion-error",
	Mobile_host_redirect:        "mobile-host-redirect",
	Ip6_where_are_you:           "ip6-where-are-you",
	Ip6_i_am_here:               "ip6-i-am-here",
	Mobile_registration_request: "mobile-registration-request",
	Mobile_registration_reply:   "mobile-registration-reply",
	Domain_name_request:         "domain-name-request",
	Domain_name_reply:           "domain-name-reply",
	Skip:                        "skip",
	Photuris:                    "photuris",
}

func (t Type) String() string {
	return elib.StringerHex(typeStrings[:], int(t))
}

var typeMap = parse.NewStringMap(typeStrings[:])

func (t *Type) Parse(in *parse.Input) {
	var v uint8
	if !in.Parse("%v", typeMap, &v) {
		panic(parse.ErrInput)
	}
	*t = Type(v)
}
func (h *Header) String() string { return "ICMP4 " + h.Type.String() }

// 4 byte icmp header
type header32 struct {
	d32 [1]uint32
}

func (h *Header) checksum(payload []byte) vnet.Uint16 {
	i := (*header32)(unsafe.Pointer(h))
	c := ip.Checksum(i.d32[0])
	c = c.AddBytes(payload)
	return ^c.Fold()
}

func (h *Header) Len() uint                       { return HeaderBytes }
func (h *Header) Read(b []byte) vnet.PacketHeader { return (*Header)(vnet.Pointer(b)) }
func (h *Header) Write(b []byte) {
	h.Checksum = 0
	h.Checksum = h.checksum(b[HeaderBytes:])

	type t struct{ data [HeaderBytes]byte }
	i := (*t)(unsafe.Pointer(h))
	copy(b[:], i.data[:])
}

type EchoRequest struct {
	Id       vnet.Uint16
	Sequence vnet.Uint16
}

const EchoRequestBytes = 4

func (h *EchoRequest) String() string {
	return fmt.Sprintf("id %d seq %d", h.Id.ToHost(), h.Sequence.ToHost())
}
func (h *EchoRequest) Len() uint                       { return EchoRequestBytes }
func (h *EchoRequest) Read(b []byte) vnet.PacketHeader { return (*EchoRequest)(vnet.Pointer(b)) }
func (h *EchoRequest) Write(b []byte) {
	type t struct{ data [EchoRequestBytes]byte }
	i := (*t)(unsafe.Pointer(h))
	copy(b[:], i.data[:])
}
