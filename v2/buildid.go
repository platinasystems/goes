// Copyright © 2015-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"bytes"
	"debug/elf"
	"fmt"
)

const (
	elfNoteNameSzLength = 4
	elfNoteDescSzLength = 4
	elfNoteTypeLength   = 4
	goElfNoteNameLength = 4

	elfNoteNameSzIndex = 0
	elfNoteDescSzIndex = elfNoteNameSzIndex + elfNoteNameSzLength
	elfNoteTypeIndex   = elfNoteDescSzIndex + elfNoteDescSzLength
	elfNoteNameIndex   = elfNoteTypeIndex + elfNoteTypeLength
	goElfNoteDescIndex = elfNoteNameIndex + goElfNoteNameLength

	goElfNoteNameEnd = elfNoteNameIndex + goElfNoteNameLength

	goElfNoteType = 4
)

var goElfNoteName = [goElfNoteNameLength]byte{'G', 'o', 0, 0}

type BuildId string

func (proc BuildId) MarshalText() ([]byte, error) {
	fn := string(proc)
	f, err := elf.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	section := f.Section(".note.go.buildid")
	if section == nil {
		return nil, fmt.Errorf("%s: no note section", fn)
	}
	data, err := section.Data()
	if err != nil {
		return nil, err
	}
	noteNameSz := f.ByteOrder.Uint32(data)
	if noteNameSz != goElfNoteNameLength {
		return nil, fmt.Errorf("%s: invalid namesz: %d", fn, noteNameSz)
	}
	noteDescSz := f.ByteOrder.Uint32(data[elfNoteDescSzIndex:])
	if goElfNoteDescIndex+noteDescSz > uint32(len(data)) {
		return nil, fmt.Errorf("invalid decsz: %d", noteDescSz)
	}
	noteType := f.ByteOrder.Uint32(data[elfNoteTypeIndex:])
	if noteType != goElfNoteType {
		return nil, fmt.Errorf("invalid type: %#x", noteType)
	}
	noteName := data[elfNoteNameIndex:goElfNoteNameEnd]
	if !bytes.Equal(noteName, goElfNoteName[:]) {
		return nil, fmt.Errorf("invalid name: %q", noteName)
	}
	end := goElfNoteDescIndex + noteDescSz
	return data[goElfNoteDescIndex:end], nil
}

func (proc BuildId) String() string {
	text, err := proc.MarshalText()
	if err != nil {
		return err.Error()
	}
	return string(text)
}
