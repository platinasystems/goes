// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	"fmt"
	"net/netip"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"golang.org/x/net/dns/dnsmessage"
)

type Question struct {
	Class
	Type
	Name UniqueString
}

type Resource interface {
	String() string
	Type() Type
}

type ResourceRecord struct {
	Name UniqueString
	Secs uint32
	Class
	Resource
}

func Decode(data []byte) (msg struct {
	ID uint16
	HF
	OpCode
	RCode
	Questions []Question
	Answers,
	Authorities,
	Additionals []ResourceRecord
}, err error) {
	var p dnsmessage.Parser
	h, err := p.Start(data)
	if err != nil {
		return
	}
	msg.ID = h.ID
	msg.HF = NewHeaderFlags(h)
	msg.OpCode = OpCode(h.OpCode)
	msg.RCode = RCode(h.RCode)
	if msg.Questions, err = gatherQuestions(p.Question); err != nil {
		return
	}
	if msg.Answers, err = gatherResources(p.Answer); err != nil {
		return
	}
	if msg.Authorities, err = gatherResources(p.Authority); err != nil {
		return
	}
	msg.Additionals, err = gatherResources(p.Additional)
	return
}

func gatherQuestions(f func() (dnsmessage.Question, error)) (
	qs []Question, err error,
) {
	for {
		if q, e := f(); e != nil {
			if e == dnsmessage.ErrSectionDone {
				break
			}
			err = e
			return
		} else {
			qs = append(qs, Question{
				Class: Class(q.Class),
				Type:  Type(q.Type),
				Name:  MakeUniqueString(q.Name.String()),
			})
		}
	}
	return
}

func gatherResources(f func() (dnsmessage.Resource, error)) (
	rrs []ResourceRecord, err error,
) {
	for {
		if r, e := f(); e != nil {
			if e != dnsmessage.ErrSectionDone {
				err = e
			}
			break
		} else if rr, e := convert(r); e != nil {
			err = e
			break
		} else {
			rrs = append(rrs, rr)
		}
	}
	return
}

func convert(r dnsmessage.Resource) (rr ResourceRecord, err error) {
	rr.Name = MakeUniqueString(r.Header.Name.String())
	rr.Secs = r.Header.TTL
	rr.Class = Class(r.Header.Class)
	switch r.Header.Type {
	case dnsmessage.TypeA:
		a := r.Body.(*dnsmessage.AResource).A
		rr.Resource = A{netip.AddrFrom4(a)}
	case dnsmessage.TypeNS:
		s := r.Body.(*dnsmessage.NSResource).NS.String()
		rr.Resource = NS{MakeUniqueString(s)}
	case dnsmessage.TypeCNAME:
		s := r.Body.(*dnsmessage.CNAMEResource).CNAME.String()
		rr.Resource = CNAME{MakeUniqueString(s)}
	case dnsmessage.TypeSOA:
		soa := r.Body.(*dnsmessage.SOAResource)
		rr.Resource = SOA{
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
		rr.Resource = PTR{MakeUniqueString(s)}
	case dnsmessage.TypeMX:
		mx := r.Body.(*dnsmessage.MXResource)
		rr.Resource = MX{
			Preference: mx.Pref,
			Exchange:   MakeUniqueString(mx.MX.String()),
		}
	case dnsmessage.TypeTXT:
	case dnsmessage.TypeAAAA:
		aaaa := r.Body.(*dnsmessage.AAAAResource).AAAA
		rr.Resource = AAAA{netip.AddrFrom16(aaaa)}
	case dnsmessage.TypeSRV:
		srv := r.Body.(*dnsmessage.SRVResource)
		rr.Resource = SRV{
			Priority: srv.Priority,
			Weight:   srv.Weight,
			Port:     srv.Port,
			Target:   MakeUniqueString(srv.Target.String()),
		}
	case dnsmessage.TypeOPT:
		opts := r.Body.(*dnsmessage.OPTResource).Options
		rr.Resource = OPT(opts)
	case dnsmessage.Type(TypeSVCB):
		var svcb SVCB
		data := r.Body.(*dnsmessage.UnknownResource).Data
		err = svcb.UnmarshalBinary(data)
		rr.Resource = svcb
	case dnsmessage.Type(TypeHTTPS):
		var https HTTPS
		data := r.Body.(*dnsmessage.UnknownResource).Data
		err = https.UnmarshalBinary(data)
		rr.Resource = https
	case dnsmessage.Type(TypeCAA):
		var caa CAA
		data := r.Body.(*dnsmessage.UnknownResource).Data
		err = caa.UnmarshalBinary(data)
		rr.Resource = caa
	default:
		err = xerrors.Unsupported("type", Type(r.Header.Type))
	}
	return
}

func LineWrapResource(r Resource, indent int) {
	lns := strings.Split(r.String(), "\n")
	switch len(lns) {
	case 0:
		fmt.Println()
	case 1:
		fmt.Println(lns[0])
	default:
		fmt.Print("(", lns[0])
		for _, s := range lns[1:] {
			fmt.Printf("\n%*s%s", indent+1, "", s)
		}
		fmt.Println(")")
	}
}

// Format new query.
// If cap(buf) > 512, add EDNS0 OPT.
// Returns ID and binary request.
func NewQuestion(
	buf []byte,
	name string,
	t Type,
	c Class,
	hf HF,
) (uint16, []byte, error) {
	var rh dnsmessage.ResourceHeader
	var q dnsmessage.Question
	if buf == nil {
		buf = make([]byte, 0, 512)
	}
	qhdr := dnsmessage.Header{
		ID:                 NewID(),
		Authoritative:      (hf & HFAuthoritative) != 0,
		Truncated:          (hf & HFTruncated) != 0,
		RecursionDesired:   (hf & HFRecursionDesired) != 0,
		RecursionAvailable: (hf & HFRecursionAvailable) != 0,
		AuthenticData:      (hf & HFAuthenticData) != 0,
		CheckingDisabled:   (hf & HFCheckingDisabled) != 0,
	}
	mb := dnsmessage.NewBuilder(buf, qhdr)
	err := mb.StartQuestions()
	if err != nil {
		return qhdr.ID, nil, err
	}
	q.Name.Length = uint8(copy(q.Name.Data[:], name))
	q.Type = dnsmessage.Type(t)
	q.Class = dnsmessage.Class(c)
	err = mb.Question(q)
	if err != nil {
		return qhdr.ID, nil, err
	}
	if cap(buf) > 512 {
		err = mb.StartAdditionals()
		if err != nil {
			return qhdr.ID, nil, err
		}
		err = rh.SetEDNS0(MaxPacketSize, dnsmessage.RCodeSuccess, false)
		if err != nil {
			return qhdr.ID, nil, err
		}
		err = mb.OPTResource(rh, dnsmessage.OPTResource{})
		if err != nil {
			return qhdr.ID, nil, err
		}
	}
	fin, err := mb.Finish()
	return qhdr.ID, fin, err
}
