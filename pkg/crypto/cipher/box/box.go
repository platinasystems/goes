// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This package provides end-to-end data security with a ciphered box that has
// a separately ciphered label.
package box

import (
	"crypto/rand"
	"errors"

	"github.com/platinasystems/goes/v2/pkg/crypto/cipher/box/label"
	"github.com/platinasystems/goes/v2/pkg/crypto/cipher/gcm"
)

type Box []byte

type Opener interface {
	Open(dst, iv, ciphertext, data []byte) ([]byte, error)
}

type Sealer interface {
	Seal(dst, iv, plaintext, data []byte) []byte
}

const (
	BeginFrom = 0
	EndFrom   = BeginFrom + label.Size

	BeginVia = EndFrom
	EndVia   = BeginVia + label.Size

	BeginTo = EndVia
	EndTo   = BeginTo + label.Size

	IVSize = 4

	BeginIV = EndTo
	EndIV   = BeginIV + IVSize

	BeginContent = EndIV + gcm.Overhead

	Overhead = BeginContent + gcm.Overhead

	Size = 2 << 10
)

var (
	ErrIncomplete = errors.New("incomplete")
	ErrOverrun    = errors.New("overrun")
	ErrUnderrun   = errors.New("underrun")
)

func New() Box {
	return make(Box, Size, Size)
}

func (box Box) Clone() Box {
	clone := New()
	clone = clone[:len(box)]
	copy(clone, box)
	return clone
}

// Seal box content with a randomized IV added to the local nonce; then label
// the sealed content length and IV.
func (box Box) CloseWith(v Sealer) Box {
	contents := box[BeginContent:]
	if len(contents) == 0 {
		return box
	}
	if niv, err := rand.Read(box[BeginIV:EndIV]); err != nil {
		panic(err)
	} else if niv != IVSize {
		panic(ErrIncomplete)
	}
	contents = v.Seal(contents[:0], box[BeginIV:EndIV], contents, nil)
	box = box[:BeginContent+len(contents)]
	return box
}

func (box Box) Contents() []byte {
	return box[BeginContent:]
}

func (box Box) Empty() Box {
	return box[:BeginContent]
}

func (box Box) Expand() Box {
	return box[:cap(box)]
}

func (box Box) From(from label.Label) Box {
	from.Put(box[BeginFrom:EndFrom])
	return box
}

func (box Box) FromWhom() label.Label {
	return label.With(box[BeginFrom:EndFrom])
}

// Open box contents with label's incremental nonce value.
func (box Box) OpenWith(v Opener) (Box, error) {
	var err error
	contents := box[BeginContent:]
	if len(contents) == 0 {
		return box, nil
	}
	iv := box[BeginIV:EndIV]
	contents, err = v.Open(contents[:0], iv, contents, nil)
	return box[:BeginContent+len(contents)], err
}

// Seal box label with the shared key and nonce.
func (box Box) SealWith(v Sealer) {
	v.Seal(box[BeginTo:BeginTo], nil, box[BeginTo:EndIV], nil)
}

// Shrink box to BeginContent+n
func (box Box) Shrink(n int) Box {
	return box[:BeginContent+n]
}

func (box Box) To(to label.Label) Box {
	to.Put(box[BeginTo:EndTo])
	return box
}

func (box Box) ToWhom() label.Label {
	return label.With(box[BeginTo:EndTo])
}

// Unseal box label with the shared cipher key and nonce.
func (box Box) UnsealWith(v Opener) error {
	if len(box) < BeginContent {
		return ErrUnderrun
	}
	_, err := v.Open(box[BeginTo:BeginTo], nil,
		box[BeginTo:BeginContent], nil)
	return err
}

func (box Box) Via(via label.Label) Box {
	via.Put(box[BeginVia:EndVia])
	return box
}

func (box Box) ViaWhom() label.Label {
	return label.With(box[BeginVia:EndVia])
}
