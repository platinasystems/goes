// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This package provides end-to-end data security with a ciphered box that has
// a separately ciphered label.
package box

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/context/poll"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
)

type Box []byte

const (
	AddressBits = 25
	AddressMask = (1 << AddressBits) - 1

	ToBits   = AddressBits
	FromBits = AddressBits
	TypeBits = 3
	SizeBits = 11

	LabelBits = ToBits + FromBits + TypeBits + SizeBits

	IVBits = 64

	LabelSize = LabelBits / 8
	IVSize    = IVBits / 8
	Size      = 1 << SizeBits

	SizeBit = 0
	TypeBit = SizeBit + SizeBits
	FromBit = TypeBit + TypeBits
	ToBit   = FromBit + FromBits

	ToMask   = ((1 << ToBits) - 1) << ToBit
	FromMask = ((1 << FromBits) - 1) << FromBit
	TypeMask = ((1 << TypeBits) - 1) << TypeBit
	SizeMask = (Size - 1) << SizeBit

	BeginLabel   = 0
	EndLabel     = BeginLabel + LabelSize
	BeginIV      = EndLabel
	EndIV        = BeginIV + IVSize
	BeginContent = EndIV + CipherOverhead
)

var (
	ByteOrder   = binary.BigEndian
	ErrOverrun  = errors.New("overrun")
	ErrUnderrun = errors.New("underrun")
)

var pool = sync.Pool{New: func() any {
	return make(Box, Size, Size)
}}

func New() Box {
	return pool.Get().(Box)[:Size]
}

func (box Box) Append(v any) Box {
	switch t := v.(type) {
	case string:
		return append(box, t...)
	case []byte:
		return append(box, t...)
	default:
		panic(fmt.Errorf("can't append type %T", t))
	}
}

func (box Box) Clone() Box {
	clone := New()
	clone = clone[:len(box)]
	copy(clone, box)
	return clone
}

// Seal box content with a randomized IV added to the local nonce; then label
// the sealed content length and IV.
func (box Box) Close(c *Cipher) Box {
	contents := box[BeginContent:]
	if len(contents) > 0 {
		if niv, err := rand.Read(box[BeginIV:EndIV]); err != nil {
			panic(egress.Mark(err))
		} else if niv != IVSize {
			panic(egress.Mark(ErrIncomplete))
		}
		nonce := NoncePlusIV(c.nonce.local, box[BeginIV:EndIV])
		contents = c.gcm.Seal(contents[:0], nonce, contents, nil)
		box = box[:BeginContent+len(contents)]
	}
	lbl := box.label() &^ SizeMask
	lbl |= uint64(len(contents)) << SizeBit
	box.mark(lbl)
	return box
}

func (box Box) Contents() []byte { return box[BeginContent:] }
func (box Box) Empty() Box       { return box[:BeginContent] }
func (box Box) Expand() Box      { return box[:Size] }

func (box Box) From(from uint32) Box {
	from &= AddressMask
	box.mark((box.label() &^ FromMask) | (uint64(from) << FromBit))
	return box
}

func (box Box) FromWhom() uint32 {
	return uint32((box.label() & FromMask) >> FromBit)
}

func (box Box) IsToAll() bool {
	return (box.label() & ToMask) == ToMask
}

// Open box contents with IV added to the remote's nonce.
func (box Box) Open(c *Cipher) (Box, error) {
	var err error
	contents := box[BeginContent:]
	if len(contents) == 0 {
		return box, nil
	}
	nonce := NoncePlusIV(c.nonce.remote, box[BeginIV:EndIV])
	contents, err = c.gcm.Open(contents[:0], nonce, contents, nil)
	return box[:BeginContent+len(contents)], err
}

// Read and decipher box label from stream; then read contents.
func (box Box) Receive(ctx context.Context, conn net.Conn, bxc *Cipher) (
	Box, error,
) {
	box = box[:Size]
	ctxconn := poll.WithReader(ctx, conn)
	if _, ok := conn.(net.PacketConn); ok {
		n, err := ctxconn.Read(box)
		if err != nil {
			return box, egress.Mark(err)
		}
		if n < BeginContent {
			return box, egress.Mark(ErrUnderrun)
		}
		box = box[:n]
		if box, err = box.Unseal(bxc); err != nil {
			return box, egress.Mark(err)
		}
	} else {
		n, err := ctxconn.Read(box[:BeginContent])
		if err != nil {
			return box, egress.Mark(err)
		}
		if n != BeginContent {
			return box, egress.Mark(ErrUnderrun)
		}
		if box, err = box.Unseal(bxc); err != nil {
			return box, egress.Mark(err)
		}
		size := box.size()
		if size > Size-BeginContent {
			return box, egress.Mark(ErrOverrun)
		}
		box = box[:BeginContent+size]
		n, err = ctxconn.Read(box[BeginContent:])
		if err != nil {
			return box, egress.Mark(err)
		}
		if n != int(size) {
			return box, egress.Mark(ErrUnderrun)
		}
	}
	return box, nil
}

func (box Box) Recycle() { pool.Put(box) }

// Seal box label with the shared key and nonce.
func (box Box) Seal(c *Cipher) Box {
	c.gcm.Seal(box[:0], c.nonce.shared, box[:LabelSize+IVSize], nil)
	return box
}

func (box Box) To(to uint32) Box {
	to &= AddressMask
	box.mark((box.label() &^ ToMask) | (uint64(to) << ToBit))
	return box
}

func (box Box) ToAll() Box {
	box.mark(box.label() | ToMask)
	return box
}

func (box Box) ToWhom() uint32 {
	return uint32((box.label() & ToMask) >> ToBit)
}

func (box Box) Type() uint8 {
	return uint8((box.label() & TypeMask) >> TypeBit)
}

func (box Box) TypeCast(t uint32) Box {
	t &= (1 << TypeBits) - 1
	box.mark((box.label() &^ TypeMask) | (uint64(t) << TypeBit))
	return box
}

// Unseal box label with the shared key and nonce and return a box of
// the labelled size.
func (box Box) Unseal(c *Cipher) (Box, error) {
	_, err := c.gcm.Open(box[:0], c.nonce.shared, box[:BeginContent], nil)
	if err == nil {
		if size := box.size(); size > Size-BeginContent {
			err = ErrOverrun
		} else {
			box = box[:BeginContent+size]
		}
	}
	return box, err
}

func (box Box) label() uint64 {
	_ = box[LabelSize-1]
	return ByteOrder.Uint64(box)
}

func (box Box) mark(label uint64) {
	_ = box[LabelSize-1]
	ByteOrder.PutUint64(box, label)
}

func (box Box) size() uint32 {
	return uint32((box.label() & SizeMask) >> SizeBit)
}
