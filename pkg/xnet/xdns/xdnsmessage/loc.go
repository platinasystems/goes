// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"golang.org/x/net/dns/dnsmessage"
)

const (
	sealevel = 10000000
	ceiling  = 4284967295
)

// Centimeters + 100,000m
type Altitude uint32

func (p *Altitude) Parse(tokens []string) ([]string, error) {
	var alt float32
	if len(tokens) == 0 {
		return tokens, xerrors.Incomplete("LOC", "altitude")
	}
	_, err := fmt.Sscan(strings.TrimSuffix(tokens[0], "m"), &alt)
	if err != nil {
		return tokens, xerrors.Label(err, "LOC", "altitude")
	}
	cm := int64(alt * 100)
	if cm < -sealevel || cm > ceiling {
		return tokens, xerrors.Range("LOC", "altitude")
	}
	*p = Altitude(sealevel + cm)
	return tokens[1:], nil
}

func (v Altitude) Format(w fmt.State, verb rune) {
	if v < sealevel {
		fmt.Fprint(w, "-")
		v = sealevel - v
	} else {
		v -= sealevel
	}
	fmt.Fprintf(w, "%.2fm", (float64(v) / 100))
}

// Thousandths arc seconds
type Arc uint32

const Arc0 = 0x80000000

func (v Arc) Format(w fmt.State, verb rune) {
	if v > Arc0 {
		v -= Arc0
	} else {
		v = Arc0 - v
	}
	t := uint(v) % 1000
	v /= 1000
	s := uint(v) % 60
	v /= 60
	m := uint(v) % 60
	v /= 60
	d := uint(v)
	fmt.Fprintf(w, "%d %d %d.%03d", d, m, s, t)
}

func (p *Arc) Parse(tokens []string) ([]string, error) {
	var (
		deglimit,
		deg uint16
		min uint8
		sec float32
		ne  bool
	)
	isCompassPoint := func() bool {
		switch tokens[0] {
		case "N":
			ne = true
			fallthrough
		case "S":
			deglimit = 90
			return true
		case "E":
			ne = true
			fallthrough
		case "W":
			deglimit = 180
			return true
		}
		return false
	}
	if len(tokens) < 2 {
		return tokens, xerrors.Incomplete("LOC")
	} else if _, err := fmt.Sscan(tokens[0], &deg); err != nil {
		return tokens, xerrors.Label(err, "arc", "deg")
	} else if tokens = tokens[1:]; isCompassPoint() {
		tokens = tokens[1:]
	} else if len(tokens) < 2 {
		return tokens, xerrors.Incomplete("LOC")
	} else if _, err = fmt.Sscan(tokens[0], &min); err != nil {
		return tokens, xerrors.Label(err, "arc", "min")
	} else if tokens = tokens[1:]; isCompassPoint() {
		tokens = tokens[1:]
	} else if len(tokens) < 2 {
		return tokens, xerrors.Incomplete("LOC")
	} else if _, err = fmt.Sscan(tokens[0], &sec); err != nil {
		return tokens, xerrors.Label(err, "arc", "sec")
	} else if tokens = tokens[1:]; isCompassPoint() {
		tokens = tokens[1:]
	} else {
		return tokens, xerrors.Invalid("compass_point", tokens[0])
	}
	if deg >= deglimit {
		return tokens, xerrors.Range("arc", "degrees", deg)
	} else if min >= 60 {
		return tokens, xerrors.Range("arc", "minutes", min)
	} else if sec >= 60.0 {
		return tokens, xerrors.Range("arc", "seconds", sec)
	} else if ne {
		*p = Arc0 + (Arc(deg)*3600+Arc(min)*60)*1000 +
			Arc(sec*1000)
	} else {
		*p = Arc0 - (Arc(deg)*3600+Arc(min)*60)*1000 -
			Arc(sec*1000)
	}
	return tokens, nil
}

type Latitude struct{ Arc }

func (v Latitude) Format(w fmt.State, verb rune) {
	v.Arc.Format(w, verb)
	if v.Arc >= Arc0 {
		fmt.Fprint(w, " N")
	} else {
		fmt.Fprint(w, " S")
	}
}

type Longitude struct{ Arc }

func (v Longitude) Format(w fmt.State, verb rune) {
	v.Arc.Format(w, verb)
	if v.Arc >= Arc0 {
		fmt.Fprint(w, " E")
	} else {
		fmt.Fprint(w, " W")
	}
}

// Centimeters in nibbles of (v>>4)e(v&15)
type Precision uint8

func (p *Precision) Parse(tokens []string) ([]string, error) {
	var sp float32
	_, err := fmt.Sscan(strings.TrimSuffix(tokens[0], "m"), &sp)
	if err != nil {
		return tokens, err
	}
	tokens = tokens[1:]
	var exp Precision
	n := uint64(sp * 100)
	for ; n >= 10; exp++ {
		n /= 10
	}
	*p = (Precision(n) << 4) | exp
	return tokens, nil
}

func (v Precision) String() string {
	return fmt.Sprintf("%.2fm", (float64(v>>4)*math.Pow10(int(v&0xf)))/100)
}

// RFC-1876
type TypeLOCResource struct {
	Version uint8 // Must be 0
	Size,
	Horz,
	Vert Precision
	Latitude
	Longitude
	Altitude
}

func ParseLOC(tokens []string) (TypeLOCResource, error) {
	var (
		err error
		loc TypeLOCResource
	)
	if tokens, err = loc.Latitude.Parse(tokens); err != nil {
		return loc, fmt.Errorf("latitude: %w", err)
	}
	if tokens, err = loc.Longitude.Parse(tokens); err != nil {
		return loc, fmt.Errorf("longitude: %w", err)
	}
	if tokens, err = loc.Altitude.Parse(tokens); err != nil {
		return loc, fmt.Errorf("altitude: %w", err)
	}
	if len(tokens) == 0 {
		return loc, nil
	}
	if tokens, err = loc.Size.Parse(tokens); err != nil {
		return loc, fmt.Errorf("size: %w", err)
	}
	if len(tokens) == 0 {
		return loc, nil
	}
	if tokens, err = loc.Horz.Parse(tokens); err != nil {
		return loc, fmt.Errorf("horizontal: %w", err)
	}
	if len(tokens) == 0 {
		return loc, nil
	}
	tokens, err = loc.Vert.Parse(tokens)
	return loc, err
}

func (v TypeLOCResource) AppendTo(data []byte) ([]byte, error) {
	if data == nil {
		data = make([]byte, 0, 16)
	}
	data = append(data, 0, uint8(v.Size), uint8(v.Horz), uint8(v.Vert))
	data = binary.BigEndian.AppendUint32(data, uint32(v.Latitude.Arc))
	data = binary.BigEndian.AppendUint32(data, uint32(v.Longitude.Arc))
	data = binary.BigEndian.AppendUint32(data, uint32(v.Altitude))
	return data, nil
}

func (v TypeLOCResource) construct(
	mb *dnsmessage.Builder, h dnsmessage.ResourceHeader,
) error {
	data, _ := v.AppendTo(nil)
	return mb.UnknownResource(h, dnsmessage.UnknownResource{
		Type: dnsmessage.Type(TypeLOC),
		Data: data,
	})
}

func (v TypeLOCResource) String() string {
	w := new(strings.Builder)
	fmt.Fprintln(w, v.Latitude)
	fmt.Fprintln(w, v.Longitude)
	fmt.Fprint(w, v.Altitude)
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

func (loc *TypeLOCResource) UnmarshalBinary(data []byte) error {
	if n := len(data); n < 16 {
		return ErrUnderrun
	} else if n > 16 {
		return ErrOverrun
	}
	loc.Version = uint8(data[0])
	loc.Size = Precision(data[1])
	loc.Horz = Precision(data[2])
	loc.Vert = Precision(data[3])
	loc.Latitude.Arc = Arc(binary.BigEndian.Uint32(data[4:]))
	loc.Longitude.Arc = Arc(binary.BigEndian.Uint32(data[8:]))
	loc.Altitude = Altitude(binary.BigEndian.Uint32(data[12:]))
	return nil
}
