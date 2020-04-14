// Copyright © 2019-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This package gets a GO program's BuildId.  Usage,
//
//	s, err := buildid.New("/proc/self/exe")
//	if err != nil {
//		fmt.Frintln(os.Stderr, err)
//	} else {
//		fmt.Println(s)
//	}
//
// Which is equivalent to,
//
//	$ go tool buildid PROGRAM
package buildid

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
	if noteNameSz != goElfNoteNameLength {
		return "", fmt.Errorf("%s: invalid namesz: %d", fn, noteNameSz)
	}
	noteDescSz := f.ByteOrder.Uint32(data[elfNoteDescSzIndex:])
	if goElfNoteDescIndex+noteDescSz > uint32(len(data)) {
		return "", fmt.Errorf("invalid decsz: %d", noteDescSz)
	}
	noteType := f.ByteOrder.Uint32(data[elfNoteTypeIndex:])
	if noteType != goElfNoteType {
		return "", fmt.Errorf("invalid type: %#x", noteType)
	}
	noteName := data[elfNoteNameIndex:goElfNoteNameEnd]
	if !bytes.Equal(noteName, goElfNoteName[:]) {
		return "", fmt.Errorf("invalid name: %q", noteName)
	}
	end := goElfNoteDescIndex + noteDescSz
	return string(data[goElfNoteDescIndex:end]), nil
}
