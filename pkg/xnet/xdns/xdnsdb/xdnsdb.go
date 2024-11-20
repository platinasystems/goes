// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsdb

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"maps"
	"math"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsmessage"
)

type Class = xdnsmessage.Class
type Resource = xdnsmessage.Resource
type TypedResource = xdnsmessage.TypedResource
type TypeAResource = xdnsmessage.TypeAResource
type TypeAAAAResource = xdnsmessage.TypeAAAAResource
type UniqueString = xdnsmessage.UniqueString

var MakeUniqueString = xdnsmessage.MakeUniqueString

type Entry struct {
	Deadline time.Time
	Class    Class
	TypedResource
}

func (entry Entry) Resource() Resource {
	return entry.TypedResource
}

var db = make(map[UniqueString][]Entry)
var appendReqCh = make(chan appendReq, 4)
var dumpReqCh = make(chan dumpReq, 1)
var getReqCh = make(chan getReq, 4)
var setReqCh = make(chan setReq, 4)

var TraceInclude = func(tokens []string) {
	// fmt.Println(tokens)
}

type done chan struct{}

type appendReq struct {
	name  UniqueString
	entry Entry
	done
}

func Append(name string, entry Entry) {
	req := appendReq{
		name:  MakeUniqueString(name),
		entry: entry,
		done:  make(done),
	}
	appendReqCh <- req
	<-req.done
}

type dumpReq struct {
	w io.Writer
	done
}

func Dump(w io.Writer) {
	req := dumpReq{
		w:    w,
		done: make(done),
	}
	dumpReqCh <- req
	<-req.done
}

type getReq struct {
	name    xdnsmessage.UniqueString
	entries chan<- []Entry
}

func Get(name UniqueString) []Entry {
	ch := make(chan []Entry, 1)
	getReqCh <- getReq{name, ch}
	return <-ch
}

func Include(ctx context.Context, origin, fn string) error {
	var defaultTTL, ttl time.Duration
	var name string
	var c xdnsmessage.Class
	var tokens []string
	var cont bool

	r := os.Stdin

	if fn == "-" {
		fn = filepath.Base(os.Stdin.Name())
	} else if f, err := os.Open(fn); err == nil {
		defer f.Close()
		r = f
	} else {
		return err
	}

	now := time.Now()
	sc := bufio.NewScanner(r)
	firstln := true
	for lno := 1; ctx.Err() == nil && sc.Scan(); lno++ {
		es := func(s string) string {
			return fmt.Sprintf("%s[%d]:%s", fn, lno, s)
		}
		ln := sc.Text()
		s := strings.TrimSpace(ln)
		if len(s) == 0 || s[0] == ';' {
			continue
		}
		if firstln {
			firstln = false
			if ln[0] != ' ' && ln[0] != '\t' {
				if len(name) > 0 {
					name = name[:0]
				}
				ttl = 0
			}
		}
		for _, s := range strings.Fields(ln) {
			if s[0] == ';' {
				break
			}
			if !cont {
				if strings.HasPrefix(s, "(") {
					s = strings.TrimPrefix(s, "(")
					cont = true
				}
				if len(s) > 0 {
					tokens = append(tokens, s)
				}
			} else {
				if strings.HasSuffix(s, ")") {
					s = strings.TrimSuffix(s, ")")
					cont = false
				}
				if len(s) > 0 {
					tokens = append(tokens, s)
				}
			}
		}
		if cont || len(tokens) == 0 {
			continue
		}
		firstln = true
		TraceInclude(tokens)
		if strings.HasPrefix(tokens[0], "$") {
			s := strings.ToUpper(strings.TrimPrefix(tokens[0], "$"))
			switch s {
			case "INCLUDE":
				if len(tokens) < 2 {
					return xerrors.Incomplete(es("INCLUDE"))
				}
				zone := origin
				if len(tokens) > 2 {
					zone = strings.ToLower(tokens[2])
				}
				err := Include(ctx, tokens[1], zone)
				if err != nil {
					return xerrors.Label(err, es("INCLUDE"))
				}
			case "ORIGIN":
				if len(tokens) == 1 || len(tokens[1]) == 0 {
					return xerrors.Incomplete(es("ORIGIN"))

				}
				origin = strings.ToLower(tokens[1])
			case "TTL":
				if len(tokens) == 1 {
					return xerrors.Incomplete(es("TTL"))
				}
				u, err := strconv.ParseUint(tokens[1], 10, 32)
				if err != nil {
					return xerrors.Label(err, es("TTL"))
				}
				defaultTTL = time.Second * time.Duration(u)
				ttl = defaultTTL
			}
			tokens = tokens[:0]
			continue
		}
		if len(name) == 0 {
			switch tokens[0] {
			case "@":
				name = origin
				tokens = tokens[1:]
			case ".":
				name = "."
				tokens = tokens[1:]
			default:
				s = strings.ToLower(tokens[0])
				name = fmt.Sprint(s, ".", origin)
				tokens = tokens[1:]
			}
		}
		for len(tokens) > 0 {
			if u, err := strconv.
				ParseUint(tokens[0], 10, 32); err == nil {
				ttl = time.Second * time.Duration(u)
				tokens = tokens[1:]
				continue
			}
			if tc, err := xdnsmessage.
				ClassNamed(tokens[0]); err == nil {
				c = tc
				tokens = tokens[1:]
				continue
			}
			break
		}
		switch len(tokens) {
		case 0:
			return xerrors.Incomplete(es("TYPE"))
		case 1:
			return xerrors.Incomplete(es("resource"))
		}
		t, err := xdnsmessage.TypeNamed(tokens[0])
		if err != nil {
			return xerrors.Label(err, es("TYPE"))
		}
		tokens = tokens[1:]
		if ttl == 0 {
			ttl = defaultTTL
		}
		v, err := t.Parse(tokens)
		if err != nil {
			return xerrors.Label(err, es(t.String()))
		}
		entry := Entry{now.Add(ttl), c, v}
		tokens = tokens[:0]
		Append(name, entry)
		if t == xdnsmessage.TypeA || t == xdnsmessage.TypeAAAA {
			var addr netip.Addr
			if t == xdnsmessage.TypeA {
				addr = entry.Resource().(TypeAResource).Addr
			} else if t == xdnsmessage.TypeAAAA {
				addr = entry.Resource().(TypeAAAAResource).Addr
			}
			rname := xdnsmessage.Reverse(addr)
			Set(rname, []Entry{{
				entry.Deadline,
				entry.Class,
				xdnsmessage.NewPTR(name),
			}})
		}
	}
	return nil
}

func Routine(ctx context.Context, wg *sync.WaitGroup, verbose interface {
	Print(...any)
}) {
	verbose.Print("start")
	defer verbose.Print("stopped")
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-appendReqCh:
			db[req.name] = append(db[req.name], req.entry)
			close(req.done)
		case req := <-dumpReqCh:
			dump(req.w)
			close(req.done)
		case req := <-getReqCh:
			req.entries <- db[req.name]
		case req := <-setReqCh:
			if req.entries == nil || len(req.entries) == 0 {
				delete(db, req.name)
			} else {
				db[req.name] = req.entries
			}
			close(req.done)
		}
	}
}

type setReq struct {
	name    xdnsmessage.UniqueString
	entries []Entry
	done
}

// Remove entry with nil or empty list of entries.
func Set(name string, entries []Entry) {
	req := setReq{
		name:    xdnsmessage.MakeUniqueString(name),
		entries: entries,
		done:    make(done),
	}
	setReqCh <- req
	<-req.done
}

func dump(w io.Writer) {
	var lastzone string
	now := time.Now()
	for _, key := range slices.SortedFunc(maps.Keys(db), namecmp) {
		var zone, name string
		fullname := key.String()
		period := strings.Index(fullname, ".")
		if period < 0 {
			fmt.Fprintf(w, "ERROR: invalid name: %q\n", fullname)
			return
		} else if period == 0 {
			zone = "."
			name = "@"
		} else {
			zone = fullname[period+1:]
			name = fullname[:period]
		}
		entries := db[key]
		if len(entries) == 0 {
			fmt.Fprintln(w, "; EMPTY", fullname)
			continue
		}
		if zone != lastzone {
			fmt.Fprintln(w, "$ORIGIN", zone)
			lastzone = zone
		}
		fmt.Printf("%-24s", name)
		for i, entry := range entries {
			if i > 0 {
				fmt.Fprintf(w, "%-24s", "")
			}
			var secs uint
			if entry.Deadline.After(now) {
				fsecs := entry.Deadline.Sub(now).Seconds()
				secs = uint(math.Round(fsecs))
			} else {
			}
			fmt.Fprintf(w, "%-8d", secs)
			fmt.Fprintf(w, "%-8s", entry.Class)
			fmt.Fprintf(w, "%-8s", entry.Type())
			xdnsmessage.LineWrap(w, entry.String(), 24+8+8+8)
		}
	}
}

// Compare domain name components in right to left order. e.g.
//
//	foo.com < bar.edu
func namecmp(us1, us2 xdnsmessage.UniqueString) int {
	s1 := us1.String()
	s2 := us2.String()
	dom1 := strings.Split(s1, ".")
	dom2 := strings.Split(s2, ".")
	i1 := len(dom1) - 1
	i2 := len(dom2) - 1
	if i1 >= 0 && len(dom1[i1]) == 0 {
		i1 -= 1
	}
	if i2 >= 0 && len(dom2[i2]) == 0 {
		i2 -= 1
	}
	for ; i1 >= 0 && i2 >= 0; i1, i2 = i1-1, i2-1 {
		if dom1[i1] != dom2[i2] {
			l1 := strings.ToLower(dom1[i1])
			l2 := strings.ToLower(dom2[i2])
			return strings.Compare(l1, l2)
		}
	}
	return len(dom1) - len(dom2)
}
