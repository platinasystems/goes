// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func (p *parser) evalValueSpec(decl interface{}) (r uint64) {
	switch v := decl.(type) {
	case *ast.ValueSpec:
		r = p.eval(v.Values[0])
	default:
		panic(p.dump(v))
	}
	return
}

func (p *parser) eval(expr ast.Expr) (r uint64) {
	var err error
	switch v := expr.(type) {
	case *ast.BasicLit:
		r, err = strconv.ParseUint(v.Value, 0, 64)
		if err != nil {
			p.error(expr.Pos(), "bad integer")
			return
		}
	case *ast.Ident:
		if v.Obj != nil && v.Obj.Kind == ast.Con {
			r = p.evalValueSpec(v.Obj.Decl)
		} else {
			panic(p.dump(v))
			p.error(expr.Pos(), fmt.Sprintf("undefined %s", v.Name))
		}
	case *ast.ParenExpr:
		r = p.eval(v.X)
	case *ast.BinaryExpr:
		var x, y uint64
		x = p.eval(v.X)
		if p.errors != nil {
			return
		}
		y = p.eval(v.Y)
		switch v.Op {
		case token.ADD:
			r = x + y
		case token.SUB:
			r = x - y
		case token.MUL:
			r = x * y
		case token.QUO:
			r = x / y
		case token.REM:
			r = x % y
		case token.AND:
			r = x & y
		case token.AND_NOT:
			r = x &^ y
		case token.OR:
			r = x | y
		case token.XOR:
			r = x ^ y
		case token.SHL:
			r = x << y
		case token.SHR:
			r = x >> y
		default:
			p.error(expr.Pos(), fmt.Sprintf("unhandled binary op %s", v.Op))
		}
	default:
		p.error(expr.Pos(), fmt.Sprintf("unhandled expr %s", p.dump(v)))
	}
	return
}

func (m *Main) evalConst(d *ast.GenDecl) {
	for _, spec := range d.Specs {
		switch v := spec.(type) {
		case *ast.ValueSpec:
			if len(v.Names) != len(v.Values) {
				m.parser.error(v.Pos(), "# names != # values")
				continue
			}

			for i, n := range v.Names {
				r := m.eval(v.Values[i])
				if m.parser.errors == nil {
					m.consts[n.Name] = r
				}
			}
		}
	}
}

func (m *Main) evalConsts() {
	m.consts = make(Consts)
	for _, decl := range m.astFile.Decls {
		if d, ok := decl.(*ast.GenDecl); ok && d.Tok == token.CONST {
			m.evalConst(d)
		}
	}
}

type Node interface {
	dummy()
}

func (*Struct) dummy() {}
func (*Field) dummy()  {}

type common struct {
	name string
	typ  string
	// Offset and size in bits
	offset, size uint64
	Flags
}

type Struct struct {
	common
	// Length of array (default 1)
	len    uint64
	fields []Node
}

type Field struct {
	common
	Access
}

type Consts map[string]uint64

type Main struct {
	parser
	astFile     *ast.File
	regSize     uint64
	addrSize    uint64
	consts      Consts
	namedTypes  map[string]*Struct
	defaultType string
}

type Access int

const (
	ReadWrite Access = iota
	ReadOnly
)

type Flags int

const (
	IsBitfield Flags = 1 << iota
	IsFunc
)

//go:generate stringer -type=Flags

func (f Flags) toString() string {
	s := ""
	for f != 0 {
		x := f & -f
		s += x.String()
		f ^= x
		if f != 0 {
			s += ", "
		}
	}
	return s
}

func concatType(t0, t1 string) string {
	switch t1 {
	case "_":
		return t1
	default:
		return t0 + "_" + t1
	}
}

func (m *Main) foreachTypeSpec(f func(t *ast.TypeSpec, tok token.Token)) {
	for _, decl := range m.astFile.Decls {
		if d, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range d.Specs {
				if t, ok := spec.(*ast.TypeSpec); ok {
					f(t, d.Tok)
				}
			}
		}
	}
}

func (m *Main) findCycles(n string, fields []Node) {
	for _, f := range fields {
		switch v := f.(type) {
		case *Struct:
			m.findCycles(n, v.fields)
		case *Field:
			if v.typ == n {
				m.parser.error(m.astFile.Package, fmt.Sprintf("cycle in type reference %s", n))
				return
			}
			if len(v.typ) > 0 {
				if tp, ok := m.namedTypes[v.typ]; ok {
					m.findCycles(n, tp.fields)
				} else {
					panic(v)
				}
			}
		}
	}
}

func (m *Main) sizeStruct(s *Struct, t *ast.StructType) {
	offset := uint64(0)
	size := uint64(0)
	for _, f := range t.Fields.List {
		switch typ := f.Type.(type) {
		case *ast.StructType:
			for _, id := range f.Names {
				sf := &Struct{}
				sf.name = id.Name
				sf.typ = concatType(s.name, sf.name)
				isBitfield := typ.Incomplete
				if isBitfield {
					sf.Flags |= IsBitfield
				}
				sf.offset = offset
				m.sizeStruct(sf, typ)
				if isBitfield {
					if sf.size > m.regSize {
						m.parser.error(typ.Struct, fmt.Sprintf("size %d larger than reg size %d", sf.size, m.regSize))
						return
					}
					sf.size = m.regSize
				}
				size += sf.size
				offset += sf.size
				s.fields = append(s.fields, sf)
			}

		case *ast.ArrayType:
			for _, id := range f.Names {
				sf := &Struct{}
				sf.name = id.Name

				switch v := typ.Elt.(type) {
				case *ast.Ident:
					if v.Obj.Kind != ast.Typ {
						panic(m.parser.dump(v.Obj))
					}
					ts := v.Obj.Decl.(*ast.TypeSpec)
					sub := m.sizeType(ts, token.TYPE)
					sf.size = sub.size

				case nil:
					sf.size = m.regSize
					break
				default:
					panic(m.parser.dump(v))
				}

				expr := typ.Len.(ast.Expr)
				switch v := expr.(type) {
				case *ast.CompositeLit:
					if s.Flags&IsBitfield == 0 {
						m.parser.error(v.Elts[0].Pos(), "[hi:lo] not in bitfield")
						return
					}
					hi := m.parser.eval(v.Elts[0])
					lo := m.parser.eval(v.Elts[1])
					if hi < lo {
						m.parser.error(v.Elts[0].Pos(), "hi <= lo in bitfield")
						return
					}
					sf.size = 1 + (hi - lo)
					sf.offset = lo
					if hi > offset {
						offset = hi
						size = lo + sf.size
					}
				default:
					len := m.parser.eval(expr)
					if m.parser.errors != nil {
						return
					}

					if s.Flags&IsBitfield != 0 {
						sf.size = 1
						sf.offset = len
						if sf.offset > offset {
							offset = sf.offset
							size = offset + sf.size
						} else {
							size += sf.size
							offset += sf.size
						}
					} else {
						sf.len = len
						sf.size *= len
						sf.offset = offset
						size += sf.size
						offset += sf.size
					}
				}
				switch sf.size {
				case 0:
					sf.typ = "struct{}"
				case 1:
					sf.typ = "hwu1"
				default:
					sf.typ = concatType(s.name, sf.name)
				}
				s.fields = append(s.fields, sf)
			}

		case *ast.Ident:
			if typ.Obj.Kind != ast.Typ {
				panic(m.parser.dump(typ.Obj))
			}
			ts := typ.Obj.Decl.(*ast.TypeSpec)
			if ts.Type == nil {
				sz, _ := strconv.ParseUint(typ.Name, 0, 64)
				for _, id := range f.Names {
					sf := &Struct{}
					sf.name = id.Name
					sf.typ = concatType(s.name, sf.name)
					sf.size = sz
					sf.offset = offset
					s.fields = append(s.fields, sf)
					size += sz
					offset += sz
				}
			} else {
				sub := m.sizeType(ts, token.TYPE)

				for _, id := range f.Names {
					sf := &Struct{}
					sf.name = id.Name
					sf.typ = sub.name
					sf.size = sub.size
					sf.offset = offset
					s.fields = append(s.fields, sf)
					size += sf.size
					offset += sf.size
				}
			}

		case ast.Expr:
			// ...OFFSET
			offset = m.eval(typ)
			if m.parser.errors != nil {
				return
			}

			if f.Tag != nil {
				obj := m.parser.pkgScope.Lookup(f.Tag.Value)
				if obj.Kind != ast.Typ {
					panic(m.parser.dump(obj))
				}
				ts := obj.Decl.(*ast.TypeSpec)
				sub := m.sizeType(ts, token.TYPE)
				m.addrSize = sub.size
			}

			// Convert offset in bytes to bits.
			offset *= m.addrSize
			if offset > size {
				size = offset
			}

		default:
			panic(m.parser.dump(f))
		}
	}

	s.size = size
}

func (m *Main) sizeIdentType(s *Struct, tp *ast.Ident) (length uint64, ok bool) {
	if length, ok = m.isUintType(tp.Name); !ok {
		m.parser.error(tp.NamePos, "unknown type")
	}
	return
}

func (m *Main) sizeType(d *ast.TypeSpec, tok token.Token) *Struct {
	n := d.Name.Name
	if old, ok := m.namedTypes[n]; ok {
		return old
	}
	s := &Struct{}
	m.namedTypes[n] = s
	s.name = n
	s.typ = n
	s.len = 1
	if tok == token.FUNC {
		s.Flags |= IsFunc
	}
	switch tp := d.Type.(type) {
	case *ast.StructType:
		isBitfield := tp.Incomplete // kludge incomplete => bitfield
		if isBitfield {
			s.Flags |= IsBitfield
		}
		m.sizeStruct(s, tp)
	case *ast.ArrayType:
		var len uint64
		if tp.Len != nil {
			len = m.parser.eval(tp.Len)
			if m.parser.errors != nil {
				return nil
			}
		}
		switch elt := tp.Elt.(type) {
		case *ast.StructType:
			if tp.Len != nil {
				s.len = len
			}
			isBitfield := elt.Incomplete
			if isBitfield {
				s.Flags |= IsBitfield
			}
			m.sizeStruct(s, elt)
			if isBitfield {
				if tp.Len != nil {
					s.size = s.len
				}
				if s.size > m.regSize {
					m.parser.error(elt.Struct, fmt.Sprintf("size %d larger than reg size %d", s.size, m.regSize))
					return s
				}
			} else {
				s.size *= s.len
			}
		case nil:
			s.size = len
		default:
			panic(m.parser.dump(tp))
		}

	case *ast.Ident:
		var ok bool
		if s.size, ok = m.sizeIdentType(s, tp); !ok {
			return s
		}

	case *ast.StarExpr:
		x := tp.X.(*ast.Ident)
		var ok bool
		if s.size, ok = m.sizeIdentType(s, x); !ok {
			return s
		}
		s.len = 1
		m.defaultType = s.name
		m.regSize = s.size
		m.addrSize = s.size

	default:
		panic(m.parser.dump(d))
	}
	return s
}

func (m *Main) isUintType(name string) (len uint64, ok bool) {
	if n, _ := fmt.Sscanf(name, "uint%d", &len); n != 1 {
		return
	}
	ok = true
	return
}

func (m *Main) genBitFieldConst(w io.Writer, s *Struct) {
	if s.Flags&IsBitfield == 0 {
		return
	}
	for _, f := range s.fields {
		switch v := f.(type) {
		case *Struct:
			if v.name == "_" {
				continue
			}
			if s.size <= 64 {
				if v.size == 1 {
					fmt.Fprintf(w, "%s_%s %s = 1 << %d\n", s.name, v.name, m.regType(s.name), v.offset)
				} else {
					fmt.Fprintf(w, "%s_%s_shift = %d\n", s.name, v.name, v.offset)
					fmt.Fprintf(w, "%s_%s_mask = 0x%x\n", s.name, v.name, (1<<v.size)-1)
				}
			} else {
				fmt.Fprintf(w, "%s_%s_lo uint = %d\n", s.name, v.name, v.offset)
				fmt.Fprintf(w, "%s_%s_hi uint = %d\n", s.name, v.name, v.offset+v.size-1)
				fmt.Fprintf(w, "%s_%s_size uint = %d\n", s.name, v.name, v.size)
			}
		case *Field:
			panic(v)
		}
	}
}

func (m *Main) namedTypeForStruct(s *Struct) (t *Struct) {
	p, ok := m.namedTypes[s.typ]
	if ok && s.name != s.typ {
		t = p
	}
	return
}

func (m *Main) goTypeForBitField(s *Struct) (tp string) {
	if m.namedTypeForStruct(s) != nil {
		return s.typ
	}

	switch {
	case s.size == 1:
		tp = "bool"
	case s.size <= 8:
		tp = "uint8"
	case s.size <= 16:
		tp = "uint16"
	case s.size <= 32:
		tp = "uint32"
	case s.size <= 64:
		tp = "uint64"
	default:
		n32 := ((s.size + 31) &^ 31) / 32
		tp = fmt.Sprintf("[%d]uint32", n32)
	}
	return
}

func (m *Main) genStructFields(w io.Writer, s *Struct) {
	if s.Flags&IsBitfield == 0 {
		return
	}
	for _, f := range s.fields {
		switch v := f.(type) {
		case *Struct:
			if v.name == "_" {
				continue
			}
			fmt.Fprintf(w, "%s %s\n", v.name, m.goTypeForBitField(v))
		case *Field:
			panic(v)
		}
	}
}

func (m *Main) genStructFieldGet(w io.Writer, s *Struct, regVar, valVar string, offset uint64) {
	if s.Flags&IsBitfield == 0 {
		return
	}
	for _, f := range s.fields {
		switch v := f.(type) {
		case *Struct:
			if v.name == "_" {
				continue
			}
			t := m.namedTypeForStruct(v)
			o := v.offset + offset
			if t != nil {
				m.genStructFieldGet(w, t, regVar, fmt.Sprintf("%s.%s", valVar, v.name), o)
			} else {
				if s.size <= 64 && offset == 0 {
					if v.size == 1 {
						fmt.Fprintf(w, "if %s&(1<<%d) != 0 { %s.%s = true }\n", regVar, o, valVar, v.name)
					} else {
						fmt.Fprintf(w, "%s.%s = %s((%s >> %d) & ((1<<%d)-1))\n",
							valVar, v.name,
							m.goTypeForBitField(v),
							regVar, o, v.size)
					}
				} else {
					if v.size == 1 {
						fmt.Fprintf(w, "if get1(%s[:], %d) { %s.%s = true }\n", regVar, o, valVar, v.name)
					} else {
						fmt.Fprintf(w, "%s.%s = %s(getRange(%s[:], %d, %d))\n", valVar, v.name, m.goTypeForBitField(v), regVar,
							o, o+v.size-1)
					}
				}
			}
		case *Field:
			panic(v)
		}
	}
}

func (m *Main) genStructFieldSet(w io.Writer, s *Struct, regVar, valVar string, offset uint64) {
	if s.Flags&IsBitfield == 0 {
		return
	}
	for _, f := range s.fields {
		switch v := f.(type) {
		case *Struct:
			if v.name == "_" {
				continue
			}
			t := m.namedTypeForStruct(v)
			o := v.offset + offset
			if t != nil {
				m.genStructFieldSet(w, t, regVar, fmt.Sprintf("%s.%s", valVar, v.name), o)
			} else {
				if s.size <= 64 && offset == 0 {
					if v.size == 1 {
						fmt.Fprintf(w, "if %s.%s { %s |= 1 << %d }\n", valVar, v.name, regVar, o)
					} else {
						fmt.Fprintf(w, "%s |= %s(%s.%s & ((1 << %d) - 1)) << %d\n",
							regVar, m.regType(s.name), valVar, v.name, v.size, o)
					}
				} else {
					if v.size == 1 {
						fmt.Fprintf(w, "if %s.%s { set1(%s[:], %d, true) }\n", valVar, v.name, regVar, o)
					} else {
						fmt.Fprintf(w, "setRange(%s[:], %d, %d, uint64(%s.%s))\n", regVar, o, o+v.size-1, valVar, v.name)
					}
				}
			}
		case *Field:
			panic(v)
		}
	}
}

func (m *Main) regType(tp string) string {
	return fmt.Sprintf("%s_%s", tp, "reg")
}

func (m *Main) genBitfield(w io.Writer, s *Struct, level int) {
	regTp := m.regType(s.name)

	fmt.Fprintf(w, "type %s struct {\n", s.name)
	m.genStructFields(w, s)
	fmt.Fprintf(w, "}\n")

	tp := m.defaultType
	rTypeCast := fmt.Sprintf("(*%s)(r)", tp)
	wTypeCast := fmt.Sprintf("(%s)(reg)", tp)
	if len(tp) == 0 {
		tp = m.goTypeForBitField(s)
		rTypeCast = "r"
		wTypeCast = "reg"
	}

	fmt.Fprintf(w, "type %s %s\n", regTp, tp)
	fmt.Fprintln(w, "const (")
	m.genBitFieldConst(w, s)
	fmt.Fprintln(w, ")")

	// Only generate get/set methods for toplevel structs declared "func"
	// and for referenced bit fields.
	if level == 0 && s.Flags&IsFunc != 0 || level > 0 && s.Flags&IsBitfield != 0 {
		fmt.Fprintf(w, "func (r *%s) get() (v %s) {\n", regTp, s.name)
		fmt.Fprintf(w, "  reg := %s.read()\n", rTypeCast)
		m.genStructFieldGet(w, s, "reg", "v", 0 /* offset */)
		fmt.Fprintf(w, "  return\n")
		fmt.Fprintf(w, "}\n")

		fmt.Fprintf(w, "func (r *%s) set(v %s) {\n", regTp, s.name)
		fmt.Fprintf(w, "  var reg %s\n", regTp)
		m.genStructFieldSet(w, s, "reg", "v", 0 /* offset */)
		fmt.Fprintf(w, "  %s.write(%s)\n", rTypeCast, wTypeCast)
		fmt.Fprintf(w, "}\n")
	}
}

func (m *Main) genStruct(w io.Writer, s *Struct, level int) {
	for _, f := range s.fields {
		switch v := f.(type) {
		case *Struct:
			if v.Flags&IsBitfield != 0 {
				m.genBitfield(w, v, level+1)
			} else if len(v.fields) > 0 {
				m.genStruct(w, v, level+1)
			}
		case *Field:
			panic(v)
		}
	}

	offset := uint64(0)
	fmt.Fprintf(w, "type %s struct {\n", s.name)
	for _, f := range s.fields {
		switch v := f.(type) {
		case *Struct:
			if v.offset > offset {
				if offset != 0 {
					fmt.Fprintf(w, "_ [0x%x-0x%x]%s\n", v.offset/m.regSize, offset/m.regSize, m.defaultType)
				} else {
					fmt.Fprintf(w, "_ [0x%x]%s\n", v.offset/m.regSize, m.defaultType)
				}
				offset = v.offset
			}
			offset += v.size
			l := ""
			if v.len > 1 {
				l = fmt.Sprintf("[%d] ", v.len)
			}
			if len(v.fields) == 0 {
				fmt.Fprintf(w, "%s %s%s\n", v.name, l, m.defaultType)
			} else {
				fmt.Fprintf(w, "%s %s%s\n", v.name, l, m.regType(v.name))
			}
		case *Field:
			panic(v)
		}
	}
	fmt.Fprintf(w, "}\n")
}

func main() {
	var m Main
	var verbose, trace bool
	var inFile, outFile string

	flag.BoolVar(&verbose, "verbose", false, "Verbose output")
	flag.BoolVar(&trace, "trace", false, "Trace parser")
	flag.StringVar(&outFile, "o", "", "Output file (- for stdout)")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Printf("Usage: %s filename\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	inFile = flag.Arg(0)

	p := &m.parser

	var err error

	defer func() {
		if e := recover(); e != nil {
			// resume same panic if it's not a bailout
			if _, ok := e.(bailout); !ok {
				panic(e)
			}
		}
		p.errors.Sort()
		err = p.errors.Err()
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}
	}()

	err = p.init(inFile)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	p.trace = trace

	m.namedTypes = make(map[string]*Struct)

	af := p.parseFile()
	m.astFile = af

	if af != nil {
		for _, ident := range af.Unresolved {
			if len(ident.Name) == 0 {
				continue
			}
			if _, ok := m.isUintType(ident.Name); ok {
				continue
			}
			p.error(ident.Pos(), fmt.Sprintf("undefined %s", ident.Name))
		}
	}

	if p.errors.Len() == 0 {
		m.evalConsts()

		if _, ok := m.consts["address"]; !ok {
			m.consts["address"] = 8
		}
		m.addrSize = m.consts["address"]
	}

	if p.errors.Len() == 0 {
		m.foreachTypeSpec(func(t *ast.TypeSpec, tok token.Token) {
			m.sizeType(t, tok)
		})
	}

	if p.errors.Len() == 0 {
		// Look for circular types.
		for name, s := range m.namedTypes {
			m.findCycles(name, s.fields)
		}
	}

	if p.errors.Len() != 0 {
		return
	}

	if verbose {
		fmt.Printf("%+v\n", p.dump(af))
		fmt.Printf("%+v\n", &m)
	}

	w := new(bytes.Buffer)
	fmt.Fprintf(w, "// autogenerated: do not edit!\n")
	fmt.Fprintf(w, "// generated from hwmap\n")
	fmt.Fprintf(w, "package %s\n", af.Name.Name)

	for name, s := range m.namedTypes {
		switch name {
		case "_":
			continue

		case m.defaultType:
			fmt.Fprintf(w, "// Default type\ntype %s %s\n\n", m.defaultType, m.goTypeForBitField(s))

		default:
			if s.Flags&IsBitfield != 0 {
				m.genBitfield(w, s, 0)
			} else {
				m.genStruct(w, s, 0)
			}
		}
	}

	// gofmt result
	b := w.Bytes()
	b, err = format.Source(b)
	if err != nil {
		fmt.Printf("%s", w.Bytes())
		panic(err)
	}

	if outFile != "-" {
		if outFile == "" {
			ext := filepath.Ext(inFile)
			outFile = inFile[:len(inFile)-len(ext)] + "_hwgen.go"
		}
		err = ioutil.WriteFile(outFile, b, 0666)
		if err != nil {
			log.Fatalf("can't write output: %v\n", err)
		}
	} else {
		fmt.Printf("%s", b)
	}

	return
}

func (c *common) String() string {
	typ := c.typ
	if len(typ) == 0 {
		typ = "{}"
	}
	s := fmt.Sprintf("name %s, type: %s, offset: 0x%02x, size: %02d (0x%02x)",
		c.name, typ, c.offset, c.size, c.size)
	if c.Flags != 0 {
		s += fmt.Sprintf(", flags: %s", c.Flags.toString())
	}
	return s
}

func (t *Struct) subString(indent int) string {
	s := fmt.Sprintf("%*s", indent, " ")
	s += t.common.String()
	if t.len > 1 {
		s += fmt.Sprintf(", len %d", t.len)
	}
	for _, f := range t.fields {
		switch v := f.(type) {
		case *Struct:
			s += "\n" + v.subString(indent+2)
		case *Field:
			s += fmt.Sprintf("\n%*s%s", indent+2, " ", v.String())
		}
	}
	return s
}

func (t *Struct) String() string {
	return t.subString(0)
}

func (t *Field) String() string {
	s := t.common.String()
	return s
}

func (m *Main) String() string {
	s := "types:\n"
	for _, st := range m.namedTypes {
		s += fmt.Sprintf("%v\n", st.subString(2))
	}
	return s
}
