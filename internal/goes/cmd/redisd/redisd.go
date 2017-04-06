// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package redisd provides a redis server daemon that is started by /sbin/init
// or /usr/sbin/goesd *before* all other daemons.
package redisd

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	grs "github.com/platinasystems/go-redis-server"
	"github.com/platinasystems/go/internal/cmdline"
	"github.com/platinasystems/go/internal/fields"
	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/group"
	"github.com/platinasystems/go/internal/parms"
	"github.com/platinasystems/go/internal/redis"
	"github.com/platinasystems/go/internal/redis/publisher"
	"github.com/platinasystems/go/internal/redis/rpc/reg"
	"github.com/platinasystems/go/internal/sockfile"
	"github.com/platinasystems/go/internal/varrun"
	. "github.com/platinasystems/go/version"
)

const Name = "redisd"
const Log = varrun.Dir + "/log/redisd"

// Machines may use Init to set redisd parameters before exec.
var Init = func() {}
var once sync.Once

// Machines may restrict redisd listening to this list of net devices.
// If unset, the local admin may restrict this through /etc/default/goes ARGS.
// Otherwise, the default is all active net devices.
var Devs []string

// Machines may use this Hook to Print redis "[key: ]field: value" strings
// before any other daemons are run.
var Hook = func(*publisher.Publisher) {}

// A non-empty Machine is published to redis as "machine: Machine"
var Machine string

// Admins may override the redis listening port through /etc/default/goes ARGS.
var Port = 6379

// Machines may override this list of published hashes.
var PublishedKeys = []string{redis.DefaultHash}

type cmd struct {
	pubconn *net.UnixConn
	redisd  Redisd
}

type Redisd struct {
	mutex sync.Mutex
	devs  map[string][]*grs.Server
	sub   grs.HashSub

	reg *reg.Reg

	assignments Assignments

	published grs.HashHash

	cachedKeys    []string
	cachedSubkeys map[string][]string
}

type Assignments []*assignment

type assignment struct {
	prefix string
	v      interface{}
}

func New() *cmd { return &cmd{} }

func (*cmd) Kind() goes.Kind { return goes.Daemon }
func (*cmd) String() string  { return Name }
func (*cmd) Usage() string {
	return "redisd [-port PORT] [-set FIELD=VALUE]... [DEVICE]..."
}

func (cmd *cmd) Main(args ...string) error {
	once.Do(Init)

	parm, args := parms.New(args, "-port", "-set")
	if s := parm["-port"]; len(s) > 0 {
		_, err := fmt.Sscan(s, &Port)
		if err != nil {
			return err
		}
	}

	if len(args) == 0 {
		if len(Devs) == 0 {
			itfs, err := net.Interfaces()
			if err == nil {
				args = make([]string, len(itfs))
				for i, itf := range itfs {
					args[i] = itf.Name
				}
			}
		} else {
			args = Devs
		}
	}

	err := varrun.New(sockfile.Dir)
	if err != nil {
		return err
	}

	err = varrun.New(filepath.Dir(Log))
	if err != nil {
		return err
	}

	logf, err := varrun.Create(Log)
	if err != nil {
		return err
	}
	defer os.Remove(Log)
	defer logf.Close()

	if false {
		grs.Debugf = grs.ActualDebugf
	} else {
		grs.Stderr = logf
	}

	cmd.redisd.devs = make(map[string][]*grs.Server)
	cmd.redisd.sub = make(grs.HashSub)
	cmd.redisd.published = make(grs.HashHash)
	for _, k := range PublishedKeys {
		cmd.redisd.published[k] = make(grs.HashValue)
	}

	sfn := sockfile.Path(Name)
	cfg := grs.DefaultConfig()
	cfg = cfg.Proto("unix")
	cfg = cfg.Host(sfn)
	cfg = cfg.Handler(&cmd.redisd)

	srv, err := grs.NewServer(cfg)
	if err != nil {
		return err
	}
	err = sockfile.Chgroup(sfn, "adm")
	if err != nil {
		return err
	}

	cmd.redisd.reg, err =
		reg.New("redis-reg", cmd.redisd.assign, cmd.redisd.unassign)
	if err != nil {
		return err
	}

	cmd.redisd.devs[sfn] = []*grs.Server{srv}

	go func(redisd *Redisd, fn string, args ...string) {
		adm := group.Parse()["adm"].Gid()
		for i := 0; i < 30; i++ {
			if _, err := os.Stat(fn); err == nil {
				if adm > 0 {
					err = os.Chown(fn, os.Geteuid(), adm)
				}
				if err != nil {
					fmt.Fprint(os.Stderr, fn, ": chown: ",
						err, "\n")
				} else {
					fmt.Println("listen:", fn)
				}
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		redisd.listen(args...)
	}(&cmd.redisd, sfn, args...)

	cmd.pubconn, err = sockfile.ListenUnixgram(publisher.FileName)
	if err != nil {
		return err
	}
	go cmd.gopub()

	err = cmd.pubinit(fields.New(parm["-set"])...)
	if err != nil {
		return err
	}

	return srv.Start()
}

func (cmd *cmd) Close() error {
	var err error
	cmd.redisd.mutex.Lock()
	defer cmd.redisd.mutex.Unlock()
	for k, srvs := range cmd.redisd.devs {
		for i, srv := range srvs {
			xerr := srv.Close()
			if err == nil {
				err = xerr
			}
			srvs[i] = nil
		}
		cmd.redisd.devs[k] = cmd.redisd.devs[k][:0]
		delete(cmd.redisd.devs, k)
	}
	if cmd.redisd.reg != nil {
		cmd.redisd.reg.Srvr.Close()
	}
	cmd.pubconn.Close()
	return err
}

func (cmd *cmd) gopub() {
	const sep = ": "
	var key, field string
	var fv, value []byte
	b := make([]byte, 4096)
	for {
		n, err := cmd.pubconn.Read(b)
		if err != nil {
			break
		}
		t := bytes.TrimSpace(b[:n])
		x := bytes.Split(t, []byte(sep))
		switch len(x) {
		case 2:
			key = redis.DefaultHash
			field = string(x[0])
			value = x[1]
			fv = t
		case 3:
			key = string(x[0])
			field = string(x[1])
			value = x[2]
			fv = t[bytes.Index(t, []byte(sep))+2:]
		default:
			continue
		}
		cmd.redisd.mutex.Lock()
		cws := cmd.redisd.sub[key]
		hv, found := cmd.redisd.published[key]
		if !found {
			hv = make(grs.HashValue)
			cmd.redisd.published[key] = hv
		}
		if field == "delete" {
			delete(hv, string(value))
		} else {
			_, found := hv[field]
			if !found {
				hv[field] = make([]byte, 0, 256)
			} else {
				hv[field] = hv[field][:0]
			}
			hv[field] = append(hv[field], value...)
		}
		cmd.redisd.flushSubkeyCache(key)
		cmd.redisd.mutex.Unlock()
		if len(cws) > 0 {
			cmd.pubdist(cws, key, fv)
		}
	}
}

func (*cmd) pubdist(cws []*grs.ChannelWriter, key string, fv []byte) {
	mb := make([]byte, len(fv))
	copy(mb, fv)
	msg := []interface{}{
		"message",
		key,
		mb,
	}
	for _, cw := range cws {
		select {
		case cw.Channel <- msg:
		default:
		}
	}
}

func (cmd *cmd) pubinit(fieldEqValues ...string) error {
	pub, err := publisher.New()
	if err != nil {
		return err
	}
	defer pub.Close()

	pub.Print("version: ", Version)
	if hostname, err := os.Hostname(); err == nil {
		pub.Print("hostname: ", hostname)
	}
	if len(Machine) > 0 {
		pub.Print("machine: ", Machine)
	}
	if keys, cl, err := cmdline.New(); err == nil {
		for _, k := range keys {
			pub.Printf("cmdline.%s: %s", k, cl[k])
		}
	}

	Hook(pub)

	for _, feqv := range fieldEqValues {
		var field, value string
		eq := strings.Index(feqv, "=")
		if eq == 0 {
			continue
		}
		if eq < 0 {
			field = feqv
		} else {
			field = feqv[:eq]
		}
		if eq < len(feqv)-1 {
			value = feqv[eq+1:]
		}
		pub.Print(field, ": ", value)
	}

	pub.Print("redis.ready: true")
	return pub.Error()
}

func (*cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "a redis server",
	}
}

func (*cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	redisd - a redis server

SYNOPSIS
	redisd [-port PORT] [-set FIELD=VALUE]... [DEV]...

DESCRIPTION
	Run a redis server on the /run/goes/socks/redisd unix socket file.

OPTIONS
	DEV...	list of listening network devices
	-port PORT
		network port, default: 6379
	-set FIELD=VALUE
		initialize the default hash with the given field values`,
	}
}

func (redisd *Redisd) assign(key string, v interface{}) error {
	redisd.mutex.Lock()
	defer redisd.mutex.Unlock()
	redisd.assignments = redisd.assignments.Insert(key, v)
	redisd.flushKeyCache()
	return nil
}

func (redisd *Redisd) unassign(key string) error {
	redisd.mutex.Lock()
	defer redisd.mutex.Unlock()
	if redisd.assignments.Find(key) == nil {
		return fmt.Errorf("%s: not found", key)
	}
	redisd.assignments = redisd.assignments.Delete(key)
	redisd.flushKeyCache()
	return nil
}

func (redisd *Redisd) listen(names ...string) {
	for _, name := range names {
		dev, err := net.InterfaceByName(name)
		if err != nil {
			fmt.Fprint(os.Stderr, name, ": ", err, "\n")
			continue
		}
		if (dev.Flags & net.FlagUp) != net.FlagUp {
			fmt.Fprint(os.Stderr, name, ": down\n")
			continue
		}
		addrs, err := dev.Addrs()
		if err != nil {
			fmt.Fprint(os.Stderr, name, ": ", err, "\n")
			continue
		}
		if len(addrs) == 0 {
			fmt.Fprint(os.Stderr, name,
				": no address or isn't up\n")
			continue
		}

		srvs := make([]*grs.Server, 0, 2)

		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				fmt.Fprint(os.Stderr, addr, ": CIDR: ",
					err, "\n")
				continue
			}
			if ip.IsMulticast() {
				continue
			}
			id := fmt.Sprint("[", ip, "%", name, "]:", Port)
			cfg := grs.DefaultConfig()
			cfg = cfg.Handler(redisd)
			cfg = cfg.Port(Port)
			if ip.To4() == nil {
				cfg = cfg.Proto("tcp6")
				host := fmt.Sprint("[", ip, "%", name, "]")
				cfg = cfg.Host(host)
			} else {
				cfg = cfg.Host(ip.String())
			}
			srv, err := grs.NewServer(cfg)
			if err != nil {
				fmt.Fprint(os.Stderr, id, ": ", err, "\n")
			} else {
				srvs = append(srvs, srv)
				go srv.Start()
				fmt.Println("listen:", id)
			}
		}
		redisd.devs[name] = srvs
	}
}

func (redisd *Redisd) flushKeyCache() {
	redisd.cachedKeys = redisd.cachedKeys[:0]
}

func (redisd *Redisd) flushSubkeyCache(key string) {
	if redisd.cachedSubkeys == nil {
		return
	}
	a, found := redisd.cachedSubkeys[key]
	if found {
		redisd.cachedSubkeys[key] = a[:0]
	}
}

func (redisd *Redisd) Hexists(key, field string) (int, error) {
	redisd.mutex.Lock()
	defer redisd.mutex.Unlock()
	hv, found := redisd.published[key]
	if !found {
		return 0, fmt.Errorf("%s: not found", key)
	}
	_, found = hv[field]
	if !found {
		return 0, fmt.Errorf("%s: not found in %s", field, key)
	}
	return 1, nil
}

func (redisd *Redisd) Hget(key, field string) ([]byte, error) {
	var keys []string

	redisd.mutex.Lock()
	defer redisd.mutex.Unlock()

	hv, found := redisd.published[key]
	if !found {
		return nil, fmt.Errorf("%s: not found", key)
	}
	if len(field) == 0 {
		keys = make([]string, 0, len(hv))
		for k := range hv {
			keys = append(keys, k)
		}
	} else if b, found := hv[field]; found {
		return b, nil
	}
	if len(keys) == 0 {
		re, err := regexp.Compile(field)
		if err != nil {
			return nil, err
		}
		keys = make([]string, 0, len(hv))
		for k := range hv {
			if re.MatchString(k) {
				keys = append(keys, k)
			}
		}
		if len(keys) == 0 {
			return nil, fmt.Errorf("%s: not found in %s",
				field, key)
		}
	}
	sort.Strings(keys)
	b := make([]byte, 0, 4096)
	for i, k := range keys {
		if i > 0 {
			b = append(b, '\n')
		}
		b = append(b, []byte(k)...)
		b = append(b, []byte(": ")...)
		b = append(b, hv[k]...)
	}
	return b, nil
}

func (redisd *Redisd) Hgetall(key string) ([][]byte, error) {
	var bs [][]byte
	redisd.mutex.Lock()
	defer redisd.mutex.Unlock()
	hv, found := redisd.published[key]
	if !found {
		return bs, fmt.Errorf("%s: not found", key)
	}
	subkeys := redisd.subkeys(key, hv)
	bs = make([][]byte, 0, len(hv)*2)
	for _, k := range subkeys {
		bs = append(bs, []byte(k), hv[k])
	}
	return bs, nil
}

func (redisd *Redisd) Hkeys(key string) ([][]byte, error) {
	var bs [][]byte
	redisd.mutex.Lock()
	defer redisd.mutex.Unlock()
	hv, found := redisd.published[key]
	if !found {
		return bs, fmt.Errorf("%s: not found", key)
	}
	subkeys := redisd.subkeys(key, hv)
	bs = make([][]byte, len(subkeys))
	for i, k := range subkeys {
		bs[i] = []byte(k)
	}
	return bs, nil
}

func (redisd *Redisd) Hset(key, field string, value []byte) (int, error) {
	type t interface {
		Hset(string, string, []byte) (int, error)
	}
	f := func(key, field string, value []byte) (int, error) {
		return 0, fmt.Errorf("can't hset %s %s", key, field)
	}
	hashkey := fmt.Sprint(key, ":", field)
	redisd.mutex.Lock()
	if method, found := redisd.assignments.Find(hashkey).(t); found {
		f = method.Hset
	} else if method, found := redisd.assignments.Find(key).(t); found {
		f = method.Hset
	}
	redisd.mutex.Unlock()
	return f(key, field, value)
}

func (redisd *Redisd) Keys(pattern string) ([][]byte, error) {
	var re *regexp.Regexp
	var err error
	isMatch := func(k string) bool { return true }
	if len(pattern) > 0 && pattern != "*" {
		if strings.ContainsAny(pattern, "?*\\") {
			re, err = regexp.Compile(pattern)
			if err != nil {
				return nil, err
			}
			isMatch = func(k string) bool {
				return re.MatchString(k)
			}
		} else {
			isMatch = func(k string) bool {
				return k == pattern
			}
		}
	}
	keys := redisd.keys()
	reply := make([][]byte, 0, len(keys))
	rmap := make(map[string]struct{})
	defer func() {
		for k := range rmap {
			delete(rmap, k)
		}
		rmap = nil
	}()
	for _, k := range keys {
		if isMatch(k) {
			if _, found := rmap[k]; !found {
				reply = append(reply, []byte(k))
				rmap[k] = struct{}{}
			}
		}
	}
	return reply, nil
}

func (redisd *Redisd) keys() []string {
	redisd.mutex.Lock()
	defer redisd.mutex.Unlock()
	if len(redisd.cachedKeys) == 0 {
		for _, a := range redisd.assignments {
			k := a.prefix
			if i := strings.Index(k, ":"); i > 0 {
				k = k[:i]
			}
			redisd.cachedKeys = append(redisd.cachedKeys, k)
		}
		for _, k := range PublishedKeys {
			redisd.cachedKeys = append(redisd.cachedKeys, k)
		}
		sort.Strings(redisd.cachedKeys)
	}
	return redisd.cachedKeys
}

func (redisd *Redisd) Monitor() (*grs.MonitorReply, error) {
	// FIXME
	return &grs.MonitorReply{}, nil
}

func (redisd *Redisd) Ping() (*grs.StatusReply, error) {
	return grs.NewStatusReply("PONG"), nil
}

func (redisd *Redisd) Subscribe(channels ...[]byte) (*grs.MultiChannelWriter,
	error) {
	mcw := &grs.MultiChannelWriter{
		Chans: make([]*grs.ChannelWriter, len(channels)),
	}

	redisd.mutex.Lock()
	defer redisd.mutex.Unlock()

	for i, key := range channels {
		cw := &grs.ChannelWriter{
			FirstReply: []interface{}{
				"subscribe",
				key,
				1,
			},
			Channel: make(chan []interface{}, 128),
		}
		if redisd.sub[string(key)] == nil {
			redisd.sub[string(key)] = []*grs.ChannelWriter{cw}
		} else {
			redisd.sub[string(key)] =
				append(redisd.sub[string(key)], cw)
		}
		mcw.Chans[i] = cw
	}
	return mcw, nil
}

func (redisd *Redisd) subkeys(key string, hv grs.HashValue) []string {
	if redisd.cachedSubkeys == nil {
		redisd.cachedSubkeys = make(map[string][]string)
	}
	subkeys, found := redisd.cachedSubkeys[key]
	if !found {
		subkeys = []string{}
	}
	if len(subkeys) != len(hv) {
		subkeys = subkeys[:0]
		for k := range hv {
			subkeys = append(subkeys, k)
		}
		sort.Strings(subkeys)
		redisd.cachedSubkeys[key] = subkeys
	}
	return subkeys
}

func (as Assignments) Delete(key string) Assignments {
	for i := len(as) - 1; i >= 0; i-- {
		if strings.HasPrefix(key, as[i].prefix) {
			if i == 0 {
				as = as[1:]
			} else if i == len(as)-1 {
				as = as[:i]
			} else {
				as = append(as[:i], as[i+1:]...)
			}
			break
		}
	}
	return as
}

func (as Assignments) Find(key string) interface{} {
	for i := len(as) - 1; i >= 0; i-- {
		if strings.HasPrefix(key, as[i].prefix) {
			return as[i].v
		}
	}
	return struct{}{}
}

func (as Assignments) Insert(prefix string, v interface{}) Assignments {
	p := &assignment{prefix, v}
	if len(as) == 0 {
		return append(as, p)
	}
	for i, a := range as {
		ni := len(a.prefix)
		np := len(p.prefix)
		switch {
		case np > ni:
			return as.insertat(i, p)
		case np == ni:
			if p.prefix < a.prefix {
				return as.insertat(i, p)
			}
		}
	}
	return append(as, p)
}

func (as Assignments) insertat(i int, p *assignment) Assignments {
	return append(as[:i], append(Assignments{p}, as[i:]...)...)
}

func (as Assignments) Len() int { return len(as) }

// for reverse order, longest match sort
func (as Assignments) Less(i, j int) (t bool) {
	ni, nj := len(as[i].prefix), len(as[j].prefix)
	switch {
	case ni < nj:
		t = true
	case ni > nj:
		t = false
	case ni == nj:
		t = as[i].prefix < as[j].prefix
	default:
		panic("oops")
	}
	return t
}

func (as Assignments) Swap(i, j int) {
	as[i], as[j] = as[j], as[i]
}
