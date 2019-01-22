// Copyright © 2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package buildid

import (
	"bytes"
	"debug/elf"
	"fmt"
)

const (
	ElfNoteNameSzLength = 4
	ElfNoteDescSzLength = 4
	ElfNoteTypeLength   = 4
	GoElfNoteNameLength = 4

	ElfNoteNameSzIndex = 0
	ElfNoteDescSzIndex = ElfNoteNameSzIndex + ElfNoteNameSzLength
	ElfNoteTypeIndex   = ElfNoteDescSzIndex + ElfNoteDescSzLength
	ElfNoteNameIndex   = ElfNoteTypeIndex + ElfNoteTypeLength
	GoElfNoteDescIndex = ElfNoteNameIndex + GoElfNoteNameLength

	GoElfNoteNameEnd = ElfNoteNameIndex + GoElfNoteNameLength

	GoElfNoteType = 4
)

var GoElfNoteName = [GoElfNoteNameLength]byte{'G', 'o', 0, 0}

func New(fn string) (string, error) {
	f, err := elf.Open(fn)
	if err != nil {
		return "", err
	}
	defer f.Close()
	section := f.Section(".note.go.buildid")
	if section == nil {
		return "", fmt.Errorf("%s: no note section", fn)
	}
	data, err := section.Data()
	if err != nil {
		return "", err
	}
	noteNameSz := f.ByteOrder.Uint32(data)
	if noteNameSz != GoElfNoteNameLength {
		return "", fmt.Errorf("%s: invalid namesz: %d", fn, noteNameSz)
	}
	noteDescSz := f.ByteOrder.Uint32(data[ElfNoteDescSzIndex:])
	if GoElfNoteDescIndex+noteDescSz > uint32(len(data)) {
		return "", fmt.Errorf("invalid decsz: %d", noteDescSz)
	}
	noteType := f.ByteOrder.Uint32(data[ElfNoteTypeIndex:])
	if noteType != GoElfNoteType {
		return "", fmt.Errorf("invalid type: %#x", noteType)
	}
	noteName := data[ElfNoteNameIndex:GoElfNoteNameEnd]
	if !bytes.Equal(noteName, GoElfNoteName[:]) {
		return "", fmt.Errorf("invalid name: %q", noteName)
	}
	end := GoElfNoteDescIndex + noteDescSz
	return string(data[GoElfNoteDescIndex:end]), nil
}
