// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package reply provides types and methods for the redis RPC replies.
package reply

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type BB [][]byte
type Values []reflect.Value

type Del int
type Get []byte
type Hdel int
type Hexists int
type Hget []byte
type Hgetall struct{ BB }
type Hkeys struct{ BB }
type Hset int
type Lrange struct{ BB }
type Lindex []byte
type Blpop struct{ BB }
type Brpop struct{ BB }
type Lpush int
type Rpush int

func (p *BB) Sappend(s string) {
	*p = append(*p, []byte(s))
}

func (p *BB) Sprint(args ...interface{}) {
	*p = append(*p, []byte(fmt.Sprint(args...)))
}

func (bb BB) Redis() [][]byte { return [][]byte(bb) }
func (bb BB) Len() int        { return len(bb.Redis()) }

func (bb BB) Less(i, j int) bool {
	return bytes.Compare(bb.Redis()[i], bb.Redis()[j]) < 0
}

func (bb BB) Swap(i, j int) {
	bb.Redis()[i], bb.Redis()[j] = bb.Redis()[j], bb.Redis()[i]
}

func (bb BB) String() string {
	buf := &bytes.Buffer{}
	for i, b := range bb {
		if i > 0 {
			buf.Write([]byte{' '})
		}
		buf.Write(b)
	}
	return buf.String()
}

func (r Del) Redis() int { return int(r) }

func (r Get) Redis() []byte { return []byte(r) }

func (r Hdel) Redis() int { return int(r) }

func (r Hexists) Redis() int { return int(r) }

func (r Hget) Redis() []byte { return []byte(r) }

// Reflection recursively descends the given struct to retrieve the value of
// the named member.
func (p *Hget) Reflection(fields []string, v interface{}) error {
	val := reflect.Indirect(reflect.ValueOf(v))
	if val.Kind() == reflect.Map {
		mv := val.MapIndex(reflect.ValueOf(fields[0]))
		if !mv.IsValid() {
			return fmt.Errorf("%s not found", fields[0])
		}
		mv = reflect.Indirect(mv)
		mvkind := mv.Kind()
		if mvkind == reflect.Struct || mvkind == reflect.Map {
			return p.Reflection(fields[1:], mv.Interface())
		}
		*p = append(*p, []byte(fmt.Sprint(mv.Interface()))...)
		return nil
	}
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		tfield := val.Type().Field(i)
		name := strings.ToLower(tfield.Name)
		if name != fields[0] {
			continue
		}
		if len(fields) > 1 {
			fkind := field.Kind()
			if fkind == reflect.Struct || fkind == reflect.Map {
				return p.Reflection(fields[1:],
					field.Interface())
			} else {
				break
			}
		}
		*p = append(*p, []byte(fmt.Sprint(field.Interface()))...)
		return nil
	}
	return fmt.Errorf("%s not found", fields[0])
}

func (r Hgetall) Redis() [][]byte { return r.BB.Redis() }

// Reflection recursively descends the given struct to retrieve the name, value
// of each member.
func (p *Hgetall) Reflection(prefix string, v interface{}) {
	reflection(&p.BB, true, prefix, v)
}

func (r Hkeys) Redis() [][]byte { return r.BB.Redis() }

// Reflection recursively loads the receiver with the lowercased name of each
// member from the given struct.
func (p *Hkeys) Reflection(prefix string, v interface{}) {
	reflection(&p.BB, false, prefix, v)
}

func (r Hset) Redis() int        { return int(r) }
func (r Lrange) Redis() [][]byte { return r.BB.Redis() }
func (r Lindex) Redis() []byte   { return []byte(r) }
func (r Blpop) Redis() [][]byte  { return r.BB.Redis() }
func (r Brpop) Redis() [][]byte  { return r.BB.Redis() }
func (r Lpush) Redis() int       { return int(r) }
func (r Rpush) Redis() int       { return int(r) }

func MakeBB(size int) BB { return BB(make([][]byte, 0, size)) }

func MakeHget(size int) Hget       { return Hget(make([]byte, 0, size)) }
func MakeHgetall(size int) Hgetall { return Hgetall{MakeBB(size)} }
func MakeHkeys(size int) Hkeys     { return Hkeys{MakeBB(size)} }
func MakeLrange(size int) Lrange   { return Lrange{MakeBB(size)} }
func MakeBlpop(size int) Blpop     { return Blpop{MakeBB(size)} }
func MakeBrpop(size int) Brpop     { return Brpop{MakeBB(size)} }

func Key(v interface{}) string {
	s := fmt.Sprint(v)
	if strings.Contains(s, ".") {
		s = fmt.Sprint("[", s, "]")
	}
	return s
}

func reflection(pbb *BB, withValues bool, prefix string, v interface{}) {
	val := reflect.Indirect(reflect.ValueOf(v))
	if val.Kind() == reflect.Map {
		keys := Values(val.MapKeys())
		sort.Sort(keys)
		for _, mk := range keys {
			mv := reflect.Indirect(val.MapIndex(mk))
			mvkind := mv.Kind()
			if mvkind == reflect.Struct || mvkind == reflect.Map {
				reflection(pbb, withValues,
					fmt.Sprint(prefix, Key(mk.Interface()),
						"."),
					mv.Interface())
			} else if mv.Interface() != nil {
				if withValues {
					if s := vfilter(mv); len(s) > 0 {
						pbb.Sprint(prefix,
							Key(mk.Interface()))
						pbb.Sappend(s)
					}
				} else {
					pbb.Sprint(prefix, Key(mk.Interface()))
				}
			}
		}
		return
	}
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		tfield := val.Type().Field(i)
		name := strings.ToLower(tfield.Name)
		fkind := field.Kind()
		if fkind == reflect.Struct || fkind == reflect.Map {
			reflection(pbb, withValues,
				fmt.Sprint(prefix, name, "."),
				field.Interface())
		} else if field.Interface() != nil {
			if withValues {
				if s := vfilter(field); len(s) > 0 {
					pbb.Sprint(prefix, name)
					pbb.Sappend(s)
				}
			} else {
				pbb.Sprint(prefix, name)
			}
		}
	}
}

func vfilter(v reflect.Value) string {
	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			return fmt.Sprint(v.Interface())
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64:
		if v.Int() != 0 {
			return fmt.Sprint(v.Interface())
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64:
		if v.Uint() != 0 {
			return fmt.Sprint(v.Interface())
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() != 0.0 {
			return fmt.Sprint(v.Interface())
		}
	case reflect.String:
		if s := v.String(); len(s) > 0 {
			return s
		}
	default:
		return fmt.Sprint(v.Interface())
	}
	return ""
}

func (values Values) Len() int { return len(values) }

func (values Values) Less(i, j int) bool {
	switch values[i].Kind() {
	case reflect.String:
		return values[i].String() < values[j].String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64:
		return values[i].Int() < values[j].Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64:
		return values[i].Uint() < values[j].Uint()
	case reflect.Float32, reflect.Float64:
		return values[i].Float() < values[j].Float()
	default:
		return fmt.Sprint(values[i].Interface()) <
			fmt.Sprint(values[j].Interface())
	}
}

func (values Values) Swap(i, j int) {
	values[i], values[j] = values[j], values[i]
}
