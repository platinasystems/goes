// Copyright © 2024-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"golang.org/x/net/dns/dnsmessage"
)

const Cap = 1232

var (
	ErrOverrun  = errors.New("overrun")
	ErrUnderrun = errors.New("underrun")
)

type WireQuestion struct {
	Name UniqueString
	Class
	Type
}

type WireResource struct {
	Name     UniqueString
	Duration time.Duration
	Class    Class
	TypedResource
}

type TypedResource interface {
	Type() Type
	Resource
}

type Resource interface {
	fmt.Stringer
	construct(*dnsmessage.Builder, dnsmessage.ResourceHeader) error
}

func (wr WireResource) Resource() Resource {
	return wr.TypedResource
}

func (wr WireResource) Seconds() uint32 {
	return uint32(wr.Duration / time.Second)
}

func (wr WireResource) build(mb *dnsmessage.Builder) error {
	var h dnsmessage.ResourceHeader
	wr.Name.rename(&h.Name)
	h.Class = dnsmessage.Class(wr.Class)
	h.Type = dnsmessage.Type(wr.Type())
	h.TTL = wr.Seconds()
	return wr.construct(mb, h)
}

type Message struct {
	net.Addr
	ID uint16
	HF
	OpCode
	RCode
	Questions []WireQuestion
	Answers,
	Authorities,
	Additionals []WireResource
}

var MessagePool = sync.Pool{
	New: func() any { return new(Message) },
}

func NewMessage() *Message {
	return MessagePool.Get().(*Message)
}

func (m *Message) Free() {
	m.Addr = nil
	m.ID = 0
	m.HF = 0
	m.OpCode = 0
	m.RCode = 0
	if len(m.Questions) > 0 {
		m.Questions = m.Questions[:0]
	}
	if len(m.Answers) > 0 {
		m.Answers = m.Answers[:0]
	}
	if len(m.Authorities) > 0 {
		m.Answers = m.Authorities[:0]
	}
	if len(m.Additionals) > 0 {
		m.Answers = m.Additionals[:0]
	}
	MessagePool.Put(m)
}

// Append new query, or with any answers, response to given buffer.
// If there are no Additionals and cap(buf) > 512, add EDNS0 OPT.
// If “Message.ID“ is 0, set to random number.
func (m *Message) AppendTo(data []byte) ([]byte, error) {
	var err error
	var rh dnsmessage.ResourceHeader
	for m.ID == 0 {
		m.ID = NewID()
	}
	if data == nil {
		data = make([]byte, 0, 512)
	}
	mb := dnsmessage.NewBuilder(data, dnsmessage.Header{
		ID:                 m.ID,
		OpCode:             dnsmessage.OpCode(m.OpCode),
		RCode:              dnsmessage.RCode(m.RCode),
		Response:           (m.HF & HFResponse) != 0,
		Authoritative:      (m.HF & HFAuthoritative) != 0,
		Truncated:          (m.HF & HFTruncated) != 0,
		RecursionDesired:   (m.HF & HFRecursionDesired) != 0,
		RecursionAvailable: (m.HF & HFRecursionAvailable) != 0,
		AuthenticData:      (m.HF & HFAuthenticData) != 0,
		CheckingDisabled:   (m.HF & HFCheckingDisabled) != 0,
	})
	if len(m.Questions) > 0 {
		if err = mb.StartQuestions(); err != nil {
			return data[:0], err
		}
		for _, mq := range m.Questions {
			var q dnsmessage.Question
			mq.Name.rename(&q.Name)
			q.Class = dnsmessage.Class(mq.Class)
			q.Type = dnsmessage.Type(mq.Type)
			if err = mb.Question(q); err != nil {
				return data[:0], err
			}
		}
	}
	if len(m.Answers) > 0 {
		if err = mb.StartAnswers(); err != nil {
			return data[:0], err
		}
		for _, a := range m.Answers {
			if err = a.build(&mb); err != nil {
				return data[:0], err
			}
		}
	}
	if len(m.Additionals) > 0 {
		if err = mb.StartAdditionals(); err != nil {
			return data[:0], err
		}
		for _, a := range m.Additionals {
			if err = a.build(&mb); err != nil {
				return data[:0], err
			}
		}
	} else if cap(data) > 512 {
		err = mb.StartAdditionals()
		if err != nil {
			return nil, err
		}
		err = rh.SetEDNS0(cap(data), dnsmessage.RCodeSuccess, false)
		if err != nil {
			return nil, err
		}
		err = mb.OPTResource(rh, dnsmessage.OPTResource{})
		if err != nil {
			return nil, err
		}
	}
	if len(m.Authorities) > 0 {
		if err = mb.StartAuthorities(); err != nil {
			return data[:0], err
		}
		for _, a := range m.Authorities {
			if err = a.build(&mb); err != nil {
				return data[:0], err
			}
		}
	}
	return mb.Finish()
}

func (m *Message) UnmarshalBinary(data []byte) error {
	var p dnsmessage.Parser
	h, err := p.Start(data)
	if err != nil {
		return xerrors.Mark(err)
	}
	m.ID = h.ID
	m.HF = NewHeaderFlags(h)
	m.OpCode = OpCode(h.OpCode)
	m.RCode = RCode(h.RCode)
	if m.Questions, err = unpackQuestions(p.Question); err != nil {
		return xerrors.Mark(err)
	}
	if m.Answers, err = unpackResources(p.Answer); err != nil {
		return xerrors.Mark(err)
	}
	if m.Authorities, err = unpackResources(p.Authority); err != nil {
		return xerrors.Mark(err)
	}
	m.Additionals, err = unpackResources(p.Additional)
	return xerrors.Mark(err)
}

func unpackQuestions(unpack func() (dnsmessage.Question, error)) (
	[]WireQuestion, error,
) {
	var wqs []WireQuestion
	for i := 0; true; i++ {
		q, err := unpack()
		if err != nil {
			if errors.Is(err, dnsmessage.ErrSectionDone) {
				break
			}
			return wqs, fmt.Errorf("question[%d]: %w", i, err)
		}
		wqs = append(wqs, WireQuestion{
			Name:  MakeUniqueString(q.Name.String()),
			Class: Class(q.Class),
			Type:  Type(q.Type),
		})

	}
	return wqs, nil
}

func unpackResources(unpack func() (dnsmessage.Resource, error)) (
	[]WireResource, error,
) {
	var wrs []WireResource
	for {
		r, err := unpack()
		if err != nil {
			if errors.Is(err, dnsmessage.ErrSectionDone) {
				break
			}
			return wrs, err
		}
		wr, err := makeWireResource(r)
		if err != nil {
			return wrs, err
		}
		wrs = append(wrs, wr)
	}
	return wrs, nil
}

func makeWireResource(r dnsmessage.Resource) (WireResource, error) {
	var err error
	var v TypedResource
	switch r.Header.Type {
	case dnsmessage.TypeA:
		a := r.Body.(*dnsmessage.AResource).A
		v = TypeAResource{netip.AddrFrom4(a)}
	case dnsmessage.TypeNS:
		s := r.Body.(*dnsmessage.NSResource).NS.String()
		v = TypeNSResource{MakeUniqueString(s)}
	case dnsmessage.TypeCNAME:
		s := r.Body.(*dnsmessage.CNAMEResource).CNAME.String()
		v = TypeCNAMEResource{MakeUniqueString(s)}
	case dnsmessage.TypeSOA:
		soa := r.Body.(*dnsmessage.SOAResource)
		v = TypeSOAResource{
			MName:   MakeUniqueString(soa.MBox.String()),
			RName:   MakeUniqueString(soa.NS.String()),
			Serial:  soa.Serial,
			Refresh: soa.Refresh,
			Retry:   soa.Retry,
			Expire:  soa.Expire,
			Minimum: soa.MinTTL,
		}
	case dnsmessage.TypePTR:
		s := r.Body.(*dnsmessage.PTRResource).PTR.String()
		v = TypePTRResource{MakeUniqueString(s)}
	case dnsmessage.TypeMX:
		mx := r.Body.(*dnsmessage.MXResource)
		v = TypeMXResource{
			Preference: mx.Pref,
			Exchange:   MakeUniqueString(mx.MX.String()),
		}
	case dnsmessage.TypeTXT:
		txt := r.Body.(*dnsmessage.TXTResource)
		v = TypeTXTResource(txt.TXT)
	case dnsmessage.TypeAAAA:
		aaaa := r.Body.(*dnsmessage.AAAAResource).AAAA
		v = TypeAAAAResource{netip.AddrFrom16(aaaa)}
	case dnsmessage.Type(TypeLOC):
		var loc TypeLOCResource
		data := r.Body.(*dnsmessage.UnknownResource).Data
		err = loc.UnmarshalBinary(data)
		v = loc
	case dnsmessage.TypeSRV:
		srv := r.Body.(*dnsmessage.SRVResource)
		v = TypeSRVResource{
			Priority: srv.Priority,
			Weight:   srv.Weight,
			Port:     srv.Port,
			Target:   MakeUniqueString(srv.Target.String()),
		}
	case dnsmessage.TypeOPT:
		opts := r.Body.(*dnsmessage.OPTResource).Options
		v = TypeOPTResource(opts)
	case dnsmessage.Type(TypeSVCB):
		var svcb TypeSVCBResource
		data := r.Body.(*dnsmessage.UnknownResource).Data
		err = svcb.UnmarshalBinary(data)
		v = svcb
	case dnsmessage.Type(TypeHTTPS):
		var https TypeHTTPSResource
		data := r.Body.(*dnsmessage.UnknownResource).Data
		err = https.UnmarshalBinary(data)
		v = https
	case dnsmessage.Type(TypeCAA):
		var caa TypeCAAResource
		data := r.Body.(*dnsmessage.UnknownResource).Data
		err = caa.UnmarshalBinary(data)
		v = caa
	default:
		clone := bytes.Clone(r.Body.(*dnsmessage.UnknownResource).Data)
		v = TypeTBDResource{Type(r.Header.Type), clone}
	}
	return WireResource{
		MakeUniqueString(r.Header.Name.String()),
		time.Second * time.Duration(r.Header.TTL),
		Class(r.Header.Class),
		v,
	}, err
}

func LineWrap(w io.Writer, s string, indent int) {
	lns := strings.Split(s, "\n")
	switch len(lns) {
	case 0:
		fmt.Fprintln(w)
	case 1:
		fmt.Fprintln(w, lns[0])
	default:
		fmt.Fprint(w, "(", lns[0])
		for _, ln := range lns[1:] {
			fmt.Fprintf(w, "\n%*s%s", indent+1, "", ln)
		}
		fmt.Fprintln(w, ")")
	}
}

var BufferPool = sync.Pool{
	New: func() any { return make([]byte, Cap, Cap) },
}

func FreeBuffer(b []byte) {
	b = b[:cap(b)]
	BufferPool.Put(b)
}

func MakeBuffer() []byte { return BufferPool.Get().([]byte) }

func NewQuery(recursive bool, name UniqueString, c Class, t Type) *Message {
	var hf HF
	if recursive {
		hf |= HFRecursionDesired
	}
	if c == 0 {
		c = ClassINET
	}
	if t == 0 {
		t = TypeA
	}
	m := NewMessage()
	m.HF = hf
	m.ID = NewID()
	m.OpCode = OpCodeQuery
	if len(m.Questions) == 0 {
		m.Questions = make([]WireQuestion, 1)
	} else {
		m.Questions = m.Questions[:1]
	}
	m.Questions[0].Name = name
	m.Questions[0].Class = c
	m.Questions[0].Type = t
	return m
}
