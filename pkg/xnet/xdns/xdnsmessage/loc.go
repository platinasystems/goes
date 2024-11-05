// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	"fmt"
	"math"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
)

// RFC-1876
type LOC struct {
	Version uint8 // Must be 0
	Size,
	Horz,
	Vert LOCPrecision
	Lat,
	Lon LOCArc
	Alt LOCAlt
}

// Centimeters + 100,000m
type LOCAlt uint32

// Thousandths arc seconds
type LOCArc uint32

const (
	LOCArcHemBit = 31

	LOCArcHem = 1 << LOCArcHemBit
	LOCArcSec = 1000
	LOCArcMin = LOCArcSec * 60
	LOCArcDeg = LOCArcMin * 60
)

// Centimeters in nibbles of (v>>4)e(v&15)
type LOCPrecision uint8

func ParseLOC(tokens []string) (LOC, []string, error) {
	var (
		err error
		loc LOC
	)
	if tokens, err = loc.Lat.Parse(tokens); err != nil {
		return loc, tokens, fmt.Errorf("latitude: %w", err)
	}
	if tokens, err = loc.Lon.Parse(tokens); err != nil {
		return loc, tokens, fmt.Errorf("longitude: %w", err)
	}
	if tokens, err = loc.Alt.Parse(tokens); err != nil {
		return loc, tokens, fmt.Errorf("altitude: %w", err)
	}
	if len(tokens) == 0 {
		return loc, tokens, nil
	}
	if tokens, err = loc.Size.Parse(tokens); err != nil {
		return loc, tokens, fmt.Errorf("size: %w", err)
	}
	if len(tokens) == 0 {
		return loc, tokens, nil
	}
	if tokens, err = loc.Horz.Parse(tokens); err != nil {
		return loc, tokens, fmt.Errorf("horizontal: %w", err)
	}
	if len(tokens) == 0 {
		return loc, tokens, nil
	}
	tokens, err = loc.Vert.Parse(tokens)
	return loc, tokens, err
}

func (v LOC) String() string {
	w := new(strings.Builder)
	fmt.Fprintln(w, v.Lat, [2]rune{
		0: 'S',
		1: 'N',
	}[v.Lat.Hem()])
	fmt.Fprintln(w, v.Lon, [2]rune{
		0: 'W',
		1: 'E',
	}[v.Lon.Hem()])
	fmt.Fprint(w, v.Alt)
	if v.Size != 0 {
		fmt.Fprint(w, "\n", v.Size)
	}
	if v.Horz != 0 {
		fmt.Fprint(w, "\n", v.Horz)
	}
	if v.Vert != 0 {
		fmt.Fprint(w, "\n", v.Vert)
	}
	return w.String()
}

func (p *LOCAlt) Parse(tokens []string) ([]string, error) {
	var alt float32
	if len(tokens) == 0 {
		return tokens, xerrors.Incomplete("LOC")
	}
	_, err := fmt.Sscan(strings.TrimSuffix(tokens[0], "m"), &alt)
	if err == nil {
		*p = LOCAlt((alt * 100) + 100000)
		tokens = tokens[1:]
	}
	return tokens, nil
}

func (v LOCAlt) String() string {
	return fmt.Sprintf("%.2fm", (float64(v)-100000)/100)
}

func (p *LOCArc) Parse(tokens []string) ([]string, error) {
	var (
		arc LOCArc
		deg uint16
		min uint8
		sec float32
	)
	if len(tokens) < 2 {
		return tokens, xerrors.Incomplete("LOC")
	}
	_, err := fmt.Sscan(tokens[0], &deg)
	if err != nil {
		return tokens, err
	}
	arc = LOCArc(deg) * LOCArcDeg
	tokens = tokens[1:]
	switch tokens[0] {
	case "N", "E":
		arc |= LOCArcHem
		tokens = tokens[1:]
	case "S", "W":
		tokens = tokens[1:]
	default:
		if _, err = fmt.Sscan(tokens[0], &min); err != nil {
			return tokens, err
		}
		arc += LOCArc(min) * LOCArcMin
		tokens = tokens[1:]
		if len(tokens) == 0 {
			return tokens, xerrors.Incomplete("LOC")
		}
		switch tokens[0] {
		case "N", "E":
			arc |= LOCArcHem
			tokens = tokens[1:]
		case "S", "W":
			tokens = tokens[1:]
		default:
			if _, err = fmt.Sscan(tokens[0], &sec); err != nil {
				return tokens, err
			}
			arc += LOCArc(sec * float32(LOCArcSec))
			tokens = tokens[1:]
			if len(tokens) == 0 {
				return tokens, xerrors.Incomplete("LOC")
			}
			switch tokens[0] {
			case "N", "E":
				arc |= LOCArcHem
				tokens = tokens[1:]
			case "S", "W":
				tokens = tokens[1:]
			default:
				return tokens, xerrors.Invalid(tokens[0])
			}
		}
	}
	*p = arc
	return tokens, nil
}

func (v LOCArc) Hem() uint8 {
	return uint8(v>>LOCArcHemBit) & 1
}

func (v LOCArc) Deg() uint16 {
	v &= LOCArcHem - 1
	return uint16(v / LOCArcDeg)
}

func (v LOCArc) Min() uint8 {
	v &= LOCArcHem - 1
	v %= LOCArcDeg
	return uint8(v / LOCArcMin)
}

func (v LOCArc) Sec() float32 {
	v &= LOCArcHem - 1
	v %= LOCArcDeg
	v %= LOCArcMin
	sec := float32(v / LOCArcSec)
	if tsec := float32(v % LOCArcSec); tsec > 0 {
		sec += 1 / tsec
	}
	return sec
}

func (v LOCArc) String() string {
	return fmt.Sprintf("%d %d %.03f", v.Deg(), v.Min(), v.Sec())
}

func (p *LOCPrecision) Parse(tokens []string) ([]string, error) {
	var sp float32
	_, err := fmt.Sscan(strings.TrimSuffix(tokens[0], "m"), &sp)
	if err != nil {
		return tokens, err
	}
	tokens = tokens[1:]
	var exp LOCPrecision
	n := uint64(sp * 100)
	for ; n >= 10; exp++ {
		n /= 10
	}
	*p = (LOCPrecision(n) << 4) | exp
	return tokens, nil
}

func (v LOCPrecision) String() string {
	return fmt.Sprintf("%.2fm", (float64(v>>4)*math.Pow10(int(v&0xf)))/100)
}
