// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// This package contains the *Args and *Reply types for redis RPC along with
// standard commands for the local server.
package redis

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/platinasystems/go/sch"
	"github.com/platinasystems/go/sockfile"
)

const Timeout = 500 * time.Millisecond

type AssignArgs struct {
	Key  string
	File string
	Name string
}

type UnassignArgs struct {
	Key string
}

type BB [][]byte

type DelArgs struct {
	Key  string
	Keys []string
}

type DelReply int
type GetArgs struct{ Key string }
type GetReply []byte

type SetArgs struct {
	Key   string
	Value []byte
}

type HdelArgs struct {
	Key    string
	Field  string
	Fields []string
}

type HdelReply int
type HexistsArgs struct{ Key, Field string }
type HexistsReply int
type HgetArgs struct{ Key, Field string }
type HgetReply []byte
type HgetallArgs struct{ Key string }
type HgetallReply struct{ BB }
type HkeysArgs struct{ Key string }
type HkeysReply struct{ BB }

type HsetArgs struct {
	Key, Field string
	Value      []byte
}

type HsetReply int

type LrangeArgs struct {
	Key   string
	Start int
	Stop  int
}

type LrangeReply struct{ BB }

type LindexArgs struct {
	Key   string
	Index int
}

type LindexReply []byte

type BlpopArgs struct {
	Key  string
	Keys []string
}

type BlpopReply struct{ BB }

type BrpopArgs struct {
	Key  string
	Keys []string
}

type BrpopReply struct{ BB }

type LpushArgs struct {
	Key    string
	Value  []byte
	Values [][]byte
}

type LpushReply int

type RpushArgs struct {
	Key    string
	Value  []byte
	Values [][]byte
}

type RpushReply int

type Values []reflect.Value

var keyRe *regexp.Regexp
var empty = struct{}{}

func Key(v interface{}) string {
	s := fmt.Sprint(v)
	if strings.Contains(s, ".") {
		s = fmt.Sprint("[", s, "]")
	}
	return s
}

// Split a structured redis key.
//
//	"eth0" -> ["eth0"]
//	"eth0.mtu" -> ["eth0", "mtu"]
//	"eth0.addr.[10.0.2.15]" -> ["eth0", "addr", "10.0.2.15"]
//	"eth0.addr.[10.0.2.15].broadcast" ->
//		["eth0", "addr", "10.0.2.15", "broadcast"]
func Split(key string) []string {
	if keyRe == nil {
		keyRe = regexp.MustCompile("[\\[][^\\]]+[\\]]|[^.]+")
	}
	fields := keyRe.FindAllString(key, -1)
	for i, field := range fields {
		if field[0] == '{' {
			if field[len(field)-1] == '}' {
				fields[i] = field[1 : len(field)-1]
			}
		} else if field[0] == '[' {
			if field[len(field)-1] == ']' {
				fields[i] = field[1 : len(field)-1]
			}
		}
	}
	return fields
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

func (r DelReply) Redis() int { return int(r) }

func (r GetReply) Redis() []byte { return []byte(r) }

func (r HdelReply) Redis() int { return int(r) }

func (r HexistsReply) Redis() int { return int(r) }

func (r HgetReply) Redis() []byte { return []byte(r) }

// Reflection recursively descends the given struct to retrieve the value of
// the named member.
func (p *HgetReply) Reflection(fields []string, v interface{}) error {
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

func (r HgetallReply) Redis() [][]byte { return r.BB.Redis() }

// Reflection recursively descends the given struct to retrieve the name, value
// of each member.
func (p *HgetallReply) Reflection(prefix string, v interface{}) {
	reflection(&p.BB, true, prefix, v)
}

func (r HkeysReply) Redis() [][]byte { return r.BB.Redis() }

// Reflection recursively loads the receiver with the lowercased name of each
// member from the given struct.
func (p *HkeysReply) Reflection(prefix string, v interface{}) {
	reflection(&p.BB, false, prefix, v)
}

func (r HsetReply) Redis() int { return int(r) }

func (r LrangeReply) Redis() [][]byte { return r.BB.Redis() }

func (r LindexReply) Redis() []byte { return []byte(r) }

func (r BlpopReply) Redis() [][]byte { return r.BB.Redis() }

func (r BrpopReply) Redis() [][]byte { return r.BB.Redis() }

func (r LpushReply) Redis() int { return int(r) }

func (r RpushReply) Redis() int { return int(r) }

func MakeBB(size int) BB { return BB(make([][]byte, 0, size)) }

func MakeHgetReply(size int) HgetReply {
	return HgetReply(make([]byte, 0, size))
}

func MakeHgetallReply(size int) HgetallReply {
	return HgetallReply{MakeBB(size)}
}

func MakeHkeysReply(size int) HkeysReply   { return HkeysReply{MakeBB(size)} }
func MakeLrangeReply(size int) LrangeReply { return LrangeReply{MakeBB(size)} }
func MakeBlpopReply(size int) BlpopReply   { return BlpopReply{MakeBB(size)} }
func MakeBrpopReply(size int) BrpopReply   { return BrpopReply{MakeBB(size)} }

// Assign an RPC handler for the given key.
func Assign(key, file, name string) error {
	cl, err := sockfile.NewRpcClient("redis-reg")
	if err != nil {
		return err
	}
	defer cl.Close()
	return cl.Call("RedisReg.Assign", AssignArgs{key, file, name}, &empty)
}

// Unassign an RPC handler for the given key.
func Unassign(key string) error {
	cl, err := sockfile.NewRpcClient("redis-reg")
	if err != nil {
		return err
	}
	defer cl.Close()
	return cl.Call("RedisReg.Unassign", UnassignArgs{key}, &empty)
}

// Connect to the redis file socket.
func Connect() (redis.Conn, error) {
	conn, err := sockfile.Dial("redisd")
	if err != nil {
		return nil, err
	}
	return redis.NewConn(conn, Timeout, Timeout), nil
}

func Get(key string) (s string, err error) {
	conn, err := Connect()
	if err != nil {
		return
	}
	defer conn.Close()
	v, err := conn.Do("GET", key)
	if v != nil && err == nil {
		s = vstring(v)
	}
	return
}

func Hdel(key, field string, fields ...string) (i int, err error) {
	conn, err := Connect()
	if err != nil {
		return
	}
	defer conn.Close()
	ret, err := conn.Do("HDEL", key, field, fields)
	if err == nil {
		i = int(ret.(int64))
	}
	return
}

func Hexists(key, field string) (i int, err error) {
	conn, err := Connect()
	if err != nil {
		return
	}
	defer conn.Close()
	ret, err := conn.Do("HEXISTs", key, field)
	if err == nil {
		i = int(ret.(int64))
	}
	return
}

func Hget(key, field string) (s string, err error) {
	conn, err := Connect()
	if err != nil {
		return
	}
	defer conn.Close()
	v, err := conn.Do("HGET", key, field)
	if v != nil && err == nil {
		s = vstring(v)
	}
	return
}

func Hkeys(key string) (keys []string, err error) {
	conn, err := Connect()
	if err != nil {
		return
	}
	defer conn.Close()
	ret, err := conn.Do("HKEYS", key)
	if ret != nil && err == nil {
		vs := ret.([]interface{})
		keys = make([]string, 0, len(vs))
		for _, v := range vs {
			keys = append(keys, vstring(v))
		}
	}
	return
}

func Hset(key, field string, v interface{}) (i int, err error) {
	conn, err := Connect()
	if err != nil {
		return
	}
	defer conn.Close()
	ret, err := conn.Do("HSET", key, field, v)
	if ret != nil && err == nil {
		i = int(ret.(int64))
	}
	return
}

func Keys(pattern string) (keys []string, err error) {
	conn, err := Connect()
	if err != nil {
		return
	}
	defer conn.Close()
	ret, err := conn.Do("KEYS", pattern)
	if ret != nil && err == nil {
		vs := ret.([]interface{})
		keys = make([]string, 0, len(vs))
		for _, v := range vs {
			keys = append(keys, vstring(v))
		}
	}
	return
}

func Lrange(key string, start, stop int) (keys []string, err error) {
	conn, err := Connect()
	if err != nil {
		return
	}
	defer conn.Close()
	ret, err := conn.Do("LRANGE", key, start, stop)
	if ret != nil && err == nil {
		vs := ret.([]interface{})
		keys = make([]string, 0, len(vs))
		for _, v := range vs {
			keys = append(keys, vstring(v))
		}
	}
	return
}

// Publish messages to the named redis channel.  Messages sent through the
// returned channel are forwarded to the redis server until the channel is
// closed.
//
//	pub, err := redis.Publish(NAME)
//	if err {
//		panic()
//	}
//	defer close(pub)
//	...
//	pub.Print("hello world")
func Publish(name string) (sch.In, error) {
	conn, err := Connect()
	if err != nil {
		return nil, err
	}
	in, out := sch.New(16)
	go func(name string, out sch.Out, conn redis.Conn) {
		defer conn.Close()

		for {
			// block until next message
			msg, opened := <-out
			if !opened {
				return
			}
			conn.Send("PUBLISH", name, msg)

		drain: // drain buffer of up to 64 total messages
			for n := 1; n < 64; n++ {
				select {
				case msg, opened = <-out:
					if !opened {
						conn.Do("")
						return
					}
					conn.Send("PUBLISH", name, msg)
				default:
					break drain
				}
			}
			conn.Do("")
		}

	}(name, out, conn)
	return in, nil
}

func Set(key string, value interface{}) (s string, err error) {
	conn, err := Connect()
	if err != nil {
		return
	}
	defer conn.Close()
	v, err := conn.Do("SET", key, value)
	if v != nil && err == nil {
		s = vstring(v)
	}
	return
}

func Subscribe(channel string) (out sch.Out) {
	var err error
	in, out := sch.New(4)
	defer func() {
		if err != nil {
			in <- err.Error()
			close(in)
		}
	}()
	conn, err := sockfile.Dial("redisd")
	if err != nil {
		return out
	}
	psc := redis.PubSubConn{redis.NewConn(conn, 0, Timeout)}
	if err := psc.Subscribe(channel); err != nil {
		return
	}
	go func(psc redis.PubSubConn, in sch.In) {
		for {
			v := psc.Receive()
			switch t := v.(type) {
			case redis.Message:
				in <- string(t.Data)
			case error:
				in <- t.Error()
				close(in)
				return
			}
		}
	}(psc, in)
	return
}

func vstring(v interface{}) (s string) {
	type stringer interface {
		String() string
	}
	switch t := v.(type) {
	case []byte:
		s = string(t)
	case string:
		s = t
	default:
		if method, found := t.(stringer); found {
			s = method.String()
		} else {
			s = fmt.Sprint(t)
		}
	}
	return
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
