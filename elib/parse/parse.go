package parse

import (
	"github.com/platinasystems/go/elib"

	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

type save struct {
	index int // Current buffer index
}

//go:generate gentemplate -d Package=parse -id save -d VecType=saveVec -d Type=save github.com/platinasystems/go/elib/vec.tmpl

type Input struct {
	r      io.Reader
	sawEnd bool // read EOF or reader is nil.
	buf    elib.ByteVec
	save
	saves               saveVec
	err                 error
	strictSpaceMatching bool
}

func (in *Input) Init(r io.Reader) {
	in.r = r
	in.index = 0
	in.sawEnd = false
	if in.buf != nil {
		in.buf = in.buf[:0]
	}
}

func (in *Input) Add(args ...string) {
	s := string(in.buf)
	for i := range args {
		if len(s) > 0 {
			s += " "
		}
		s += strings.TrimSpace(args[i])
	}
	in.buf = []byte(s)
}

func (in *Input) AddBuffer(b []byte) {
	if len(in.buf) == 0 {
		in.buf = b
	} else {
		in.buf = append(in.buf, b...)
	}
}

func (in *Input) Save() uint {
	i := in.saves.Len()
	in.saves.Validate(i)
	in.saves[i] = in.save
	return i
}

func (in *Input) String() (s string) {
	s = strings.TrimSpace(string(in.buf[in.index:]))
	s = strings.Replace(s, "\n", "\\n", -1)
	const max = 32
	if len(s) > max {
		s = s[:max] + "..."
	}
	return
}

func (in *Input) restore(advance bool) {
	i := in.saves.Len() - 1
	if !advance {
		in.save = in.saves[i]
	}
	in.saves = in.saves[:i]
}

func (in *Input) Advance() { in.restore(true) }
func (in *Input) Restore() { in.restore(false) }

func (in *Input) truncate() {
	i, l := in.index, len(in.buf)
	copy(in.buf[0:], in.buf[i:])
	in.index = 0
	in.buf = in.buf[:l-i]
}

var UnexpectedEOF = errors.New("unexpected end of input")

func (in *Input) read() {
	if in.r == nil {
		in.sawEnd = true
		return
	}

	if len(in.saves) == 0 {
		in.truncate()
	}

	l := len(in.buf)
	in.buf.Resize(4096)
	n, err := in.r.Read(in.buf[l:])
	in.sawEnd = err == io.EOF
	in.buf = in.buf[:l+n]
}

func (in *Input) end(skipSpace bool) (end bool) {
	for {
		for skipSpace && in.index < len(in.buf) {
			r, size := in.ReadRune()
			if !isSpace(r) {
				in.Unread(size)
				break
			}
		}
		end = in.index >= len(in.buf)
		if !end || in.sawEnd {
			return
		}
		in.read()
	}
}

func (in *Input) EndNoSkip() (end bool) { return in.end(false) }
func (in *Input) End() (end bool)       { return in.end(true) }

func (in *Input) Skip() { in.index = len(in.buf) }

func (in *Input) doRead(n int, must bool) (r []byte, ok bool) {
	for in.index+n > len(in.buf) {
		if in.sawEnd {
			if must {
				panic(UnexpectedEOF)
			} else {
				return
			}
		}
		in.read()
	}
	i0 := in.index
	i1 := i0 + n
	r = in.buf[i0:i1]
	in.index = i1
	ok = true
	return
}

func (in *Input) mustRead(n int) (r []byte)         { r, _ = in.doRead(n, true); return }
func (in *Input) tryRead(n int) (r []byte, ok bool) { return in.doRead(n, false) }

func (in *Input) mustReadByte() byte {
	x := in.mustRead(1)
	return x[0]
}

func (in *Input) fastByte() (b byte, ok bool) {
	if ok = in.index < len(in.buf); ok {
		b = in.buf[in.index]
		in.index++
	}
	return
}

func (in *Input) ReadByte() byte {
	if b, ok := in.fastByte(); ok {
		return b
	}
	return in.mustReadByte()
}

var RuneError = errors.New("input rune error")

func (in *Input) ReadRune() (r rune, size int) {
	if b, ok := in.fastByte(); ok && b < utf8.RuneSelf {
		r, size = rune(b), 1
	} else {
		var buf [utf8.UTFMax]byte
		nByte := 0
		if ok {
			nByte = 1
			buf[0] = b
		}
		for {
			buf[nByte] = in.mustReadByte()
			nByte++
			if x := buf[:nByte]; utf8.FullRune(x) {
				r, size = utf8.DecodeRune(x)
				if r == utf8.RuneError {
					panic(RuneError)
				}
				break
			}
		}
	}
	return
}

func (in *Input) Unread(size int) {
	if elib.Debug && in.index < size {
		panic("Unread")
	}
	in.index -= size
}

func (in *Input) TokenF(f func(rune) bool) (s string) {
	in.Save()
	in.skipSpace()
	i0 := in.index
	for !in.EndNoSkip() {
		r, size := in.ReadRune()
		if f(r) {
			in.Unread(size)
			break
		}
	}
	i1 := in.index
	ok := i1 > i0
	if ok {
		s = string(in.buf[i0:i1])
	}
	in.restore(ok)
	return
}

func (in *Input) Token() (s string) { return in.TokenF(unicode.IsSpace) }

func isSpace(r rune) bool { return unicode.IsSpace(r) }

func (in *Input) skipSpace() (nSpace int) {
	for !in.EndNoSkip() {
		r, size := in.ReadRune()
		if !isSpace(r) {
			in.Unread(size)
			break
		}
		nSpace++
	}
	return
}

func (in *Input) InputSpaceMustMatchFormat(new bool) (old bool) {
	old = in.strictSpaceMatching
	in.strictSpaceMatching = new
	return
}

func (in *Input) AtRune(r rune) (ok bool) {
	rʹ, size := in.ReadRune()
	if ok = rʹ == r; !ok {
		in.Unread(size)
	}
	return
}

func (in *Input) At(s string) (ok bool) {
	l := len(s)
	var r []byte
	if r, ok = in.tryRead(l); ok {
		if ok = string(r) == s; !ok {
			in.Unread(l)
		}
	}
	return
}

func (in *Input) AtOneof(s string) (i int) {
	rʹ, size := in.ReadRune()
	l := len(s)
	for i = 0; i < l; i++ {
		if rʹ == rune(s[i]) {
			return
		}
	}
	in.Unread(size)
	return
}

var IntegerOverflow = errors.New("integer overflow")

func (in *Input) parseInt(base, bitSize int, signed bool) (x uint64, ok bool) {
	in.Save()
	nDigits := 0
	negate := false
	if signed {
		negate = in.AtOneof("+-") == 1
	}
	for !in.EndNoSkip() {
		r, size := in.ReadRune()
		d := base
		switch {
		case r >= '0' && r <= '9':
			d = int(r - '0')
		case r >= 'a' && r <= 'f':
			d = 10 + int(r-'a')
		case r >= 'A' && r <= 'F':
			d = 10 + int(r-'A')
		}
		if d >= base {
			in.Unread(size)
			break
		}
		xʹ := uint64(base)*x + uint64(d)
		if xʹ < x {
			panic(IntegerOverflow)
		}
		x = xʹ
		nDigits++
	}
	ok = nDigits > 0
	in.restore(ok)
	if negate {
		x = -x
	}
	if signed {
		max := uint64(1 << uint(bitSize-1))
		if !negate && x >= max || negate && x > max {
			panic(IntegerOverflow)
		}
	} else if bitSize < 64 {
		max := uint64(1 << uint(bitSize))
		if x >= max {
			panic(IntegerOverflow)
		}
	}
	return
}

func (in *Input) ParseInt(base, bitSize int) (x int64, ok bool) {
	var v uint64
	v, ok = in.parseInt(base, bitSize, true)
	x = int64(v)
	return
}

func (in *Input) ParseUint(base, bitSize int) (x uint64, ok bool) {
	x, ok = in.parseInt(base, bitSize, false)
	return
}

var FloatOverflow = errors.New("floating point overflow")

func (in *Input) ParseFloat() (x float64, ok bool) {
	var (
		n                       [2]uint64
		nDigits                 [2]int
		negate                  [2]bool
		sawSign                 [2]bool
		sawPoint                bool
		i, nDigitsBeforeDecimal int
	)
	in.Save()
	for !in.EndNoSkip() {
		r, size := in.ReadRune()
		done := false
		switch r {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			xʹ := 10*n[i] + uint64(r-'0')
			if xʹ < n[i] {
				panic(FloatOverflow)
			}
			n[i] = xʹ
			nDigits[i]++
		case '.':
			if sawPoint || i != 0 { // second . or . after expon
				done = true
			} else if i == 0 {
				sawPoint = true
				nDigitsBeforeDecimal = nDigits[0]
			}
		case 'e', 'E':
			if i == 0 {
				i++
			} else {
				done = true // second eE
			}
		case '+', '-':
			if !sawSign[i] {
				sawSign[i] = true
				negate[i] = r == '-'
			} else {
				done = true // second sign
			}
		default:
			done = true // unknown rune
		}
		if done {
			in.Unread(size)
			break
		}
	}
	ok = nDigits[0] > 0
	in.restore(ok)
	if ok {
		// Apply sign
		for i := range n {
			if negate[i] {
				n[i] = -n[i]
			}
		}
		expon := int64(n[1])
		frac := int64(n[0])
		if sawPoint {
			expon -= int64(nDigits[0] - nDigitsBeforeDecimal)
		}
		x = timesExpon(float64(frac), expon)
	}
	return
}

// Returns y = x 10^n
func timesExpon(x float64, n int64) (y float64) {
	var (
		posPow10 = [8]float64{1e+0, 1e+1, 1e+2, 1e+3, 1e+4, 1e+5, 1e+6, 1e+7}
		negPow10 = [8]float64{1e-0, 1e-1, 1e-2, 1e-3, 1e-4, 1e-5, 1e-6, 1e-7}
	)
	y = x
	if n >= 0 {
		for n >= 8 {
			y *= 1e+8
			n -= 8
		}
		y *= posPow10[n]
	} else {
		for n <= -8 {
			y *= 1e-8
			n += 8
		}
		y *= negPow10[-n]
	}
	return
}

var ErrPercentEnd = errors.New("missing verb: % at end of format string")
var ErrMatch = errors.New("input does not match format")

func (in *Input) Error() error { return in.err }

func (in *Input) Parse(format string, args ...interface{}) (ok bool) {
	ok = true
	l := len(format)
	if l == 0 {
		return
	}

	in.Save()
	in.skipSpace()
	in.err = nil
	defer func() {
		if e := recover(); e != nil {
			ok = false
			// Save away error and input context.
			in.err = fmt.Errorf("%s: %s", e, in)
		}
		in.restore(ok)
	}()

	as := Args(args)
	matchOptional := false
	skippedSpace := false
	for i := 0; i < l; {
		fmtc, w := utf8.DecodeRuneInString(format[i:])
		matchFormat := true
		if fmtc == '%' {
			if i+w >= l {
				panic(ErrPercentEnd)
			}
			i += w
			var verb rune
			verb, w = utf8.DecodeRuneInString(format[i:])
			switch verb {
			case '%':
				// %% -> match % in input
			case '*':
				// %* -> letter characters in format up to next non-letter or end are optional.
				matchOptional = true
				matchFormat = false
			default:
				in.doPercent(verb, &as)
				matchFormat = false
			}
		}

		fmtcSpace := isSpace(fmtc)
		if matchFormat {
			// Any non-letter in format string ends optional match.
			if matchOptional {
				matchOptional = unicode.IsLetter(fmtc)
			}

			if fmtcSpace {
				if !skippedSpace {
					minSpace := 0
					if in.strictSpaceMatching {
						minSpace = 1
					}
					if ok = in.skipSpace() >= minSpace; !ok {
						break
					}
				}
			} else if matchOptional && in.EndNoSkip() {
				// Advance past optional format characters with no input.
			} else if r, size := in.ReadRune(); r != fmtc {
				if matchOptional && !unicode.IsLetter(r) {
					in.Unread(size)
				} else {
					ok = false
					break
				}
			}
		}
		skippedSpace = fmtcSpace
		i += w
	}

	// For optional match make sure input terminates with non-letter.
	// This handles the case of format "f%*oo" which should not match input "food" but
	// should match "foo!".
	if ok && matchOptional && !in.EndNoSkip() {
		r, size := in.ReadRune()
		if ok = !unicode.IsLetter(r); ok {
			in.Unread(size)
		}
	}

	return
}

func (in *Input) ParseLoose(format string, args ...interface{}) (ok bool) {
	save := in.InputSpaceMustMatchFormat(false)
	ok = in.Parse(format, args...)
	in.InputSpaceMustMatchFormat(save)
	return
}

type Args []interface{}

func (as *Args) Get() (a interface{}) {
	a = (*as)[0]
	*as = (*as)[1:]
	return
}

func (args *Args) SetNextInt(v uint64) {
	arg := args.Get()
	switch a := arg.(type) {
	case *int:
		*a = int(v)
	case *uint:
		*a = uint(v)
	case *int8:
		*a = int8(v)
	case *int16:
		*a = int16(v)
	case *int32:
		*a = int32(v)
	case *int64:
		*a = int64(v)
	case *uint8:
		*a = uint8(v)
	case *uint16:
		*a = uint16(v)
	case *uint32:
		*a = uint32(v)
	case *uint64:
		*a = uint64(v)
	default:
		val := reflect.ValueOf(a)
		ptr := val
		if ptr.Kind() != reflect.Ptr {
			panic(fmt.Errorf("type not a pointer: " + val.Type().String()))
			return
		}
		switch e := ptr.Elem(); e.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			e.SetInt(int64(v))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			e.SetUint(v)
		default:
			panic(fmt.Errorf("can't parse type: " + val.Type().String()))
		}
	}
}

var (
	intBits     = reflect.TypeOf(0).Bits()
	uintptrBits = reflect.TypeOf(uintptr(0)).Bits()
)

type Parser interface {
	Parse(in *Input)
}
type ParserWithArgs interface {
	ParseWithArgs(in *Input, args *Args)
}

func (in *Input) doParser(verb rune, p Parser, pa ParserWithArgs, args *Args) {
	in.Save()
	defer func() {
		e := recover()
		ok := e == nil
		in.restore(ok)
		if !ok {
			panic(e)
		}
	}()
	if p != nil {
		p.Parse(in)
	} else {
		pa.ParseWithArgs(in, args)
	}
}

func (in *Input) doPercent(verb rune, args *Args) {
	arg := args.Get()

	if p, ok := arg.(Parser); ok {
		in.doParser(verb, p, nil, args)
		return
	} else if pa, ok := arg.(ParserWithArgs); ok {
		in.doParser(verb, nil, pa, args)
		return
	}

	switch v := arg.(type) {
	case *bool:
		*v = in.doBool(verb)
	case *int:
		*v = int(in.doInt(verb, intBits, true))
	case *uint:
		*v = uint(in.doInt(verb, intBits, false))
	case *int8:
		*v = int8(in.doInt(verb, 8, true))
	case *uint8:
		*v = uint8(in.doInt(verb, 8, false))
	case *int16:
		*v = int16(in.doInt(verb, 16, true))
	case *uint16:
		*v = uint16(in.doInt(verb, 16, false))
	case *int32:
		*v = int32(in.doInt(verb, 32, true))
	case *uint32:
		*v = uint32(in.doInt(verb, 32, false))
	case *int64:
		*v = int64(in.doInt(verb, 64, true))
	case *uint64:
		*v = uint64(in.doInt(verb, 64, false))
	case *float64:
		*v = float64(in.doFloat(verb))
	case *float32:
		*v = float32(in.doFloat(verb))
	case *string:
		*v = string(in.doString(verb))
	case *Input:
		v.Add(string(in.doString(verb)))
	default:
		val := reflect.ValueOf(v)
		ptr := val
		if ptr.Kind() != reflect.Ptr {
			panic(fmt.Errorf("type not a pointer: " + val.Type().String()))
			return
		}
		switch v := ptr.Elem(); v.Kind() {
		case reflect.Bool:
			v.SetBool(in.doBool(verb))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v.SetInt(int64(in.doInt(verb, v.Type().Bits(), true)))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			v.SetUint(in.doInt(verb, v.Type().Bits(), false))
		case reflect.String:
			v.SetString(in.doString(verb))
		default:
			panic(fmt.Errorf("can't parse type: " + val.Type().String()))
		}
	}
}

var (
	ErrVerb            = errors.New("illegal verb after %")
	ErrInt             = errors.New("expected integer")
	ErrFloat           = errors.New("expected float")
	ErrInput           = errors.New("invalid input")
	ErrUnmatchedBraces = errors.New("unmatched braces")
)

func (in *Input) doBool(verb rune) (v bool) {
	in.skipSpace()
	if i := in.AtOneof("01ftFT"); i < 6 {
		v = i%2 != 0
	} else {
		panic(ErrInput)
	}
	return
}

func (in *Input) doInt(verb rune, bitSize int, signed bool) uint64 {
	base := 10
	switch verb {
	case 'v':
		if in.At("0") {
			base = 8 // 0DDD => octal
			if in.At("x") {
				base = 16 // 0xDDD => hex
			}
		}
	case 'd':
		base = 10
	case 'b':
		base = 2
	case 'o':
		base = 8
	case 'x', 'X':
		base = 16
	default:
		panic(ErrVerb)
	}
	if x, ok := in.parseInt(base, bitSize, signed); !ok {
		panic(ErrInt)
	} else {
		return x
	}
}

func (in *Input) doFloat(verb rune) float64 {
	switch verb {
	case 'f':
	default:
		panic(ErrVerb)
	}
	if x, ok := in.ParseFloat(); !ok {
		panic(ErrFloat)
	} else {
		return x
	}
}

// parseString parses delimited string.  If string starts with { then string
// is delimited by balenced parenthesis.  Otherwise, string is delimited by white space.
func (in *Input) parseString(delimiter rune) (s string) {
	switch delimiter {
	case '%', ' ', '\t':
		delimiter = 0
	}
	in.skipSpace()
	in.Save()
	backslash := false
	is_paren_delimited := false
	is_line_delimited := delimiter == '\n'
	paren := 0
loop:
	for !in.EndNoSkip() {
		r, size := in.ReadRune()
		add := true
		if backslash {
			backslash = false
		} else if isSpace(r) {
			if !(is_paren_delimited || is_line_delimited) || (is_line_delimited && r == '\n') {
				in.Unread(size)
				break loop
			}
		} else {
			switch r {
			case '\\':
				backslash = true
				add = false
			case '{':
				if paren == 0 && (len(s) == 0 || is_line_delimited) {
					is_paren_delimited = true
					is_line_delimited = false // no longer line delimited
					add = false
				}
				paren++
			case '}':
				paren--
				if is_paren_delimited && paren == 0 {
					break loop
				}
			default:
				if !is_paren_delimited && r == delimiter {
					in.Unread(size)
					break loop
				}
			}
		}
		if add {
			var x [utf8.MaxRune]byte
			utf8.EncodeRune(x[:], r)
			s += string(x[:size])
		}
	}
	ok := paren == 0
	in.restore(ok)
	if !ok {
		panic(ErrUnmatchedBraces)
	}
	return
}

func (in *Input) doString(verb rune) (s string) {
	switch verb {
	case 's':
		s = in.Token()
	case 'v':
		s = in.parseString(' ')
	case 'l':
		s = in.parseString('\n')
	default:
		panic(ErrVerb)
	}
	return
}
