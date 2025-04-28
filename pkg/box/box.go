// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This package provides end-to-end data security with a ciphered box that has
// a separately ciphered label.
package box

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"os"
	"runtime"
	"sync/atomic"
	"time"
	_ "unsafe"

	"github.com/platinasystems/goes/v2/pkg/gcm"
	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

const IP6MTU = netph.ETHMTU - netph.IP6Size
const UDPMTU = IP6MTU - netph.UDPSize
const Cap = UDPMTU &^ (4 - 1)

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

const SizeofZipCode = 4
const Content = Stamp + gcm.Overhead
const ContentMTU = Cap - Content - gcm.Overhead

var EncodeZip = xnet.Encode32[uint32]
var zap netip.AddrPort

type Box struct {
	data [Cap]byte
	next *Box
	netip.AddrPort
	Contents []byte
}

type Opener interface {
	Open(dst, inv, ciphertext, data []byte) ([]byte, error)
}

type Sealer interface {
	Seal(dst, inv, plaintext, data []byte) []byte
}

var (
	ErrEmpty      = errors.New("empty")
	ErrIncomplete = errors.New("incomplete")
	ErrOverrun    = errors.New("overrun")
	ErrUnderrun   = errors.New("underrun")
)

var inventory atomic.Pointer[Box]

func New() (box *Box) {
	for box = inventory.Load(); box != nil; box = inventory.Load() {
		if inventory.CompareAndSwap(box, box.next) {
			break
		}
	}
	if box == nil {
		box = new(Box)
	}
	box.Contents = box.data[Content:Content]
	box.AddrPort = zap
	return
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
func NewRx(ctx context.Context, udp *net.UDPConn) (*Box, error) {
	box := New()
	for dur := nextRxDur(0); ctx.Err() == nil; dur = nextRxDur(dur) {
		err := udp.SetReadDeadline(time.Now().Add(dur))
		if err != nil {
			box.Return()
			return nil, err
		}
		n, ap, err := udp.ReadFromUDPAddrPort(box.data[:])
		if err != nil {
			if hasExceededDeadline(err) {
				runtime.Gosched()
				continue
			}
			box.Return()
			return nil, err
		}
		if n < Content {
			box.Return()
			return nil, ErrUnderrun
		}
		box.AddrPort = ap
		box.Contents = box.data[Content:n]
		break
	}
	return box, nil
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

// Reset contents.
func (box *Box) Empty() {
	box.Contents = box.data[Content:Content]
}

func (box *Box) Format(w fmt.State, verb rune) {
	from := box.FromWhom()
	to := box.ToWhom()
	via := box.ViaWhom()
	fmt.Fprint(w, "box")
	if box.AddrPort.Addr().IsValid() {
		fmt.Fprint(w, " ", box.AddrPort, ",")
	}
	fmt.Fprint(w, " ", to.Index(), ".", to.Version())
	fmt.Fprint(w, " <- ", via.Index(), ".", via.Version())
	fmt.Fprint(w, " <- ", from.Index(), ".", from.Version())
}

func (box *Box) From(from Id) { from.Encode(box.data[From:]) }
func (box *Box) FromWhom() Id { return DecodeId(box.data[From:]) }

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
	if cap(box.Contents) < Cap-Content {
		return fmt.Errorf("read offset Contents")
	}
	zip := box.data[ZipCode:Stamp]
	box.Contents, err = v.Open(box.Contents[:0], zip, box.Contents, nil)
	return err
}

//go:linkname runtime_randn runtime.randn
func runtime_randn(n uint32) uint32

func (box *Box) RandZipCode() {
	EncodeZip(box.data[ZipCode:], runtime_randn(SizeofZipCode))
}

// Keep trying to add box to channel until successful or context is done.
func (box *Box) Queue(ctx context.Context, ch chan<- *Box) {
	for {
		select {
		case <-ctx.Done():
			box.Return()
			return
		case ch <- box:
			return
		default:
			runtime.Gosched()
		}
	}
}

func (box *Box) Return() {
	for box.next = inventory.Load(); !inventory.
		CompareAndSwap(box.next, box); box.next = inventory.Load() {
	}
}

// Seal box label with the shared key and nonce.
func (box *Box) SealWith(v Sealer) {
	v.Seal(box.data[To:To], nil, box.data[To:Stamp], nil)
}

func (box *Box) To(to Id)   { to.Encode(box.data[To:]) }
func (box *Box) ToWhom() Id { return DecodeId(box.data[To:]) }

// Send closed box with sealed label.
func (box Box) Tx(ctx context.Context, udp *net.UDPConn) (int, error) {
	const dur = 50 * time.Millisecond
	n := Content + len(box.Contents)
	for ctx.Err() == nil {
		err := udp.SetReadDeadline(time.Now().Add(dur))
		if err != nil {
			return 0, err
		}
		_, err = udp.WriteToUDPAddrPort(box.data[:n], box.AddrPort)
		if err == nil {
			break
		} else if operr, ok := err.(*net.OpError); ok {
			if !operr.Timeout() {
				return 0, operr
			}
		} else if !errors.Is(err, os.ErrDeadlineExceeded) {
			return 0, err
		}
		runtime.Gosched()
	}
	return n, nil
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
	if cap(box.Contents) < Cap-Content {
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

func (box *Box) Via(via Id)  { via.Encode(box.data[Via:]) }
func (box *Box) ViaWhom() Id { return DecodeId(box.data[Via:]) }

func nextRxDur(dur time.Duration) time.Duration {
	const min = 10 * time.Millisecond
	if dur < min {
		return min
	}
	max := 250 * time.Millisecond
	if runtime.NumCPU() <= 1 {
		max = 50 * time.Millisecond
	}
	if dur *= 2; dur > max {
		dur = max
	}
	return dur
}

func hasExceededDeadline(err error) bool {
	if operr, isOpErr := err.(*net.OpError); isOpErr {
		return operr.Timeout()
	}
	return errors.Is(err, os.ErrDeadlineExceeded)
}
