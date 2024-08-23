// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This package provides end-to-end data security with a ciphered box that has
// a separately ciphered label.
package box

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"sync/atomic"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/gcm"
)

const SizeofEthPayload = 1500
const SizeofIP6 = 4 + 2 + 1 + 1 + 16 + 16
const SizeofUDP = 2 + 2 + 2 + 2
const BoxAndLabelCap = (SizeofEthPayload - SizeofIP6 - SizeofUDP) &^ (8 - 1)

const (
	From = iota
	_
	_
	_
	Via
	_
	_
	_
	To
	_
	_
	_
	ZipCode
	_
	_
	_
	Stamp
)

const Content = Stamp + gcm.Overhead
const ContentMTU = BoxAndLabelCap - Content - gcm.Overhead

type Box struct {
	data [BoxAndLabelCap]byte
	next *Box
	netip.AddrPort
	Contents []byte
}

type Id = uint32

const SizeofId = int(unsafe.Sizeof(Id(0)))

type Opener interface {
	Open(dst, inv, ciphertext, data []byte) ([]byte, error)
}

type Sealer interface {
	Seal(dst, inv, plaintext, data []byte) []byte
}

var (
	GetId  = binary.BigEndian.Uint32
	SetId  = binary.BigEndian.PutUint32
	SetZip = binary.BigEndian.PutUint32
)

var (
	ErrEmpty      = errors.New("empty")
	ErrIncomplete = errors.New("incomplete")
	ErrOverrun    = errors.New("overrun")
	ErrUnderrun   = errors.New("underrun")
)

var inventory atomic.Pointer[Box]

func New() *Box {
	for box := inventory.Load(); box != nil; box = inventory.Load() {
		if inventory.CompareAndSwap(box, box.next) {
			box.Contents = box.data[Content:Content]
			return box
		}
	}
	box := new(Box)
	box.Contents = box.data[Content:Content]
	return box
}

func NewReadContents(r io.Reader) (*Box, error) {
	box := New()
	n, err := r.Read(box.data[Content:])
	if err != nil {
		box.Return()
		box = nil
	} else {
		box.Contents = box.data[Content : Content+n]
	}
	return box, err
}

// Receive labelled box.
func NewRx(udp *net.UDPConn) (*Box, error) {
	box := New()
	n, ap, err := udp.ReadFromUDPAddrPort(box.data[:])
	if err != nil {
		box.Return()
		box = nil
	} else if n < Content {
		box.Return()
		box = nil
		err = ErrUnderrun
	} else {
		box.AddrPort = ap
		box.Contents = box.data[Content:n]
	}
	return box, err
}

func (box *Box) Clone() *Box {
	clone := New()
	n := Content + len(box.Contents)
	copy(clone.data[:n], box.data[:n])
	clone.Contents = clone.data[Content:n]
	return clone
}

// Seal box content with a random ZipCode (aka incremental nonce value) added
// to the local nonce.
func (box *Box) CloseWith(v Sealer) {
	if len(box.Contents) == 0 {
		return
	}
	box.RandZipCode()
	zip := box.data[ZipCode:Stamp]
	box.Contents = v.Seal(box.Contents[:0], zip, box.Contents, nil)
}

func (box *Box) From(from Id) { SetId(box.data[From:], from) }
func (box *Box) FromWhom() Id { return GetId(box.data[From:]) }

// Size of label plus length of contents.
func (box *Box) Len() int {
	return Content + len(box.Contents)
}

// Non-blocking put to channel. If channel is full, return to inventory.
func (box *Box) NonBlockingPut(ch chan<- *Box) {
	select {
	case ch <- box:
	default:
		box.Return()
	}
}

// Open box contents with label's ZipCode (aka. incremental nonce value.)
func (box *Box) OpenWith(v Opener) error {
	var err error
	if len(box.Contents) == 0 {
		return ErrEmpty
	}
	if cap(box.Contents) < BoxAndLabelCap-Content {
		return fmt.Errorf("read offset Contents")
	}
	zip := box.data[ZipCode:Stamp]
	box.Contents, err = v.Open(box.Contents[:0], zip, box.Contents, nil)
	return err
}

//go:linkname runtime_randn runtime.randn
func runtime_randn(n uint32) uint32

func (box *Box) RandZipCode() {
	SetZip(box.data[ZipCode:], runtime_randn(4))
}

// Fill data with Contents then advance Contents.
func (box *Box) Read(data []byte) (int, error) {
	n := copy(data, box.Contents)
	box.Contents = box.Contents[n:]
	if len(box.Contents) == 0 {
		box.Contents = box.data[Content:Content]
	}
	return n, nil
}

func (box *Box) Return() {
	for box.next = inventory.Load(); !inventory.
		CompareAndSwap(box.next, box); box.next = inventory.Load() {
	}
}

// Restore read Contents.
func (box *Box) Rewind() {
	n := (BoxAndLabelCap - cap(box.Contents)) + len(box.Contents)
	box.Contents = box.data[Content:n]
}

// Seal box label with the shared key and nonce.
func (box *Box) SealWith(v Sealer) {
	v.Seal(box.data[To:To], nil, box.data[To:Stamp], nil)
}

func (box *Box) To(to Id)   { SetId(box.data[To:], to) }
func (box *Box) ToWhom() Id { return GetId(box.data[To:]) }

// Send closed box with sealed label.
func (box Box) Tx(udp *net.UDPConn) (int, error) {
	n := Content + len(box.Contents)
	return udp.WriteToUDPAddrPort(box.data[:n], box.AddrPort)
}

// Unseal box label with the shared cipher key and nonce.
func (box *Box) UnsealWith(v Opener) error {
	if len(box.data) < Content {
		return ErrUnderrun
	}
	_, err := v.Open(box.data[To:To], nil, box.data[To:Content], nil)
	return err
}

// This appends data to Contents.
func (box *Box) Write(data []byte) (int, error) {
	if cap(box.Contents) < BoxAndLabelCap-Content {
		box.Contents = box.data[Content:Content]
	}
	n := len(data)
	if n > cap(box.Contents) {
		return 0, ErrOverrun
	}
	box.Contents = append(box.Contents, data...)
	return n, nil
}

// Write Contents w/o label.
func (box *Box) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(box.Contents)
	box.Contents = box.data[Content:Content]
	return int64(n), err
}

func (box *Box) Via(via Id)  { SetId(box.data[Via:], via) }
func (box *Box) ViaWhom() Id { return GetId(box.data[Via:]) }
