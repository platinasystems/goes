// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package contains support to parse /proc/iomem and /proc/ioports
// and anything else of similar structure

package memmap

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type Region struct {
	What	string
	Ranges	[]*Range
}

type Range struct {
	Start	uintptr
	End	uintptr
}

type RegionMap map[string]Region

func (r Region) String() string {
	return fmt.Sprintf("%s: %v", r.What, r.Ranges)
}

func (r Range) String() string {
	return fmt.Sprintf("%x-%x", r.Start, r.End)
}

func ReaderToMap(r io.Reader) (regionMap RegionMap, err error) {

	regionMap = make(map[string]Region)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fields := strings.SplitAfterN(scanner.Text(), ":", 2)
		var start, end uintptr
		_, _ = fmt.Sscanf(fields[0], "%x-%x", &start, &end)
		// ensure i is 2 and err is nil

		key := strings.TrimSpace(fields[1])
		reg := regionMap[key]
		reg.What = key
		
		rng := &Range{}
		rng.Start = start
		rng.End = end
		
		reg.Ranges = append(reg.Ranges, rng)
		regionMap[key] = reg

	}
	return regionMap, nil
}

func FileToMap(s string) (regionMap RegionMap, err error) {
	f, err := os.OpenFile(s, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ReaderToMap(f)
}
