// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	_ "embed"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"golang.org/x/net/dns/dnsmessage"
)

// FIXME wrap, then replace these...
type Message = dnsmessage.Message
type Parser = dnsmessage.Parser

type Resource interface {
	String() string
	Type() Type
}

type ResourceRecord struct {
	Name UniqueString
	TTL  time.Duration
	V    Resource
}

func ConvertResource(r dnsmessage.Resource) (rr ResourceRecord, err error) {
	if r.Header.Class != dnsmessage.ClassINET {
		err = xerrors.Unsupported("class", Class(r.Header.Class))
		return
	}
	rr.Name = MakeUniqueString(r.Header.Name.String())
	rr.TTL = time.Second * time.Duration(r.Header.TTL)
	switch r.Header.Type {
	case dnsmessage.TypeA:
		a := r.Body.(*dnsmessage.AResource).A
		rr.V = A{netip.AddrFrom4(a)}
	case dnsmessage.TypeNS:
		s := r.Body.(*dnsmessage.NSResource).NS.String()
		rr.V = NS{MakeUniqueString(s)}
	case dnsmessage.TypeCNAME:
		s := r.Body.(*dnsmessage.CNAMEResource).CNAME.String()
		rr.V = CNAME{MakeUniqueString(s)}
	case dnsmessage.TypeSOA:
		soa := r.Body.(*dnsmessage.SOAResource)
		rr.V = SOA{
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
		rr.V = PTR{MakeUniqueString(s)}
	case dnsmessage.TypeMX:
		mx := r.Body.(*dnsmessage.MXResource)
		rr.V = MX{
			Preference: mx.Pref,
			Exchange:   MakeUniqueString(mx.MX.String()),
		}
	case dnsmessage.TypeTXT:
	case dnsmessage.TypeAAAA:
		aaaa := r.Body.(*dnsmessage.AAAAResource).AAAA
		rr.V = AAAA{netip.AddrFrom16(aaaa)}
	case dnsmessage.TypeSRV:
		srv := r.Body.(*dnsmessage.SRVResource)
		rr.V = SRV{
			Priority: srv.Priority,
			Weight:   srv.Weight,
			Port:     srv.Port,
			Target:   MakeUniqueString(srv.Target.String()),
		}
	case dnsmessage.Type(TypeSVCB):
		var svcb SVCB
		data := r.Body.(*dnsmessage.UnknownResource).Data
		err = svcb.UnmarshalBinary(data)
		rr.V = svcb
	case dnsmessage.Type(TypeHTTPS):
		var https HTTPS
		data := r.Body.(*dnsmessage.UnknownResource).Data
		err = https.UnmarshalBinary(data)
		rr.V = https
	case dnsmessage.Type(TypeCAA):
		var caa CAA
		data := r.Body.(*dnsmessage.UnknownResource).Data
		err = caa.UnmarshalBinary(data)
		rr.V = caa
	default:
		err = xerrors.Unsupported("type", Type(r.Header.Type))
	}
	return
}

func AnswerString(r dnsmessage.Resource) (s string) {
	switch Type(r.Header.Type) {
	case TypeA:
		rb := r.Body.(*dnsmessage.AResource)
		s = net.IP(rb.A[:]).String()
	case TypeNS:
		s = r.Body.(*dnsmessage.NSResource).NS.String()
	case TypeCNAME:
		s = r.Body.(*dnsmessage.CNAMEResource).CNAME.String()
	case TypeSOA:
		rb := r.Body.(*dnsmessage.SOAResource)
		s = fmt.Sprintf("ns %v, mbox %v, s/n %d",
			rb.NS, rb.MBox, rb.Serial)
	case TypePTR:
		s = r.Body.(*dnsmessage.PTRResource).PTR.String()
	case TypeMX:
		rb := r.Body.(*dnsmessage.MXResource)
		s = fmt.Sprintf("%v, pref %d", rb.MX, rb.Pref)
	case TypeTXT:
		s = strings.Join(r.Body.(*dnsmessage.TXTResource).TXT, " ")
	case TypeAAAA:
		rb := r.Body.(*dnsmessage.AAAAResource)
		s = net.IP(rb.AAAA[:]).String()
	case TypeSRV:
		rb := r.Body.(*dnsmessage.SRVResource)
		s = fmt.Sprintf("%v, port %d, pri %d, weight %d",
			rb.Target, rb.Port, rb.Priority, rb.Weight)
	case TypeSVCB:
		var svcb SVCB
		data := r.Body.(*dnsmessage.UnknownResource).Data
		if err := svcb.UnmarshalBinary(data); err != nil {
			s = err.Error()
		} else {
			s = svcb.String()
		}
	case TypeHTTPS:
		var https HTTPS
		data := r.Body.(*dnsmessage.UnknownResource).Data
		if err := https.UnmarshalBinary(data); err != nil {
			s = err.Error()
		} else {
			s = https.String()
		}
	case TypeCAA:
		var caa CAA
		data := r.Body.(*dnsmessage.UnknownResource).Data
		if err := caa.UnmarshalBinary(data); err != nil {
			s = err.Error()
		} else {
			s = caa.String()
		}
	default:
		data := r.Body.(*dnsmessage.UnknownResource).Data
		s = fmt.Sprintf("%#x", data)
	}
	return
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
