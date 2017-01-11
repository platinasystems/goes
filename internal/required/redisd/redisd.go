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
	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/group"
	"github.com/platinasystems/go/internal/parms"
	"github.com/platinasystems/go/internal/redis"
	"github.com/platinasystems/go/internal/redis/rpc/reg"
	"github.com/platinasystems/go/internal/sockfile"
	"github.com/platinasystems/go/internal/varrun"
	. "github.com/platinasystems/go/version"
)

const Name = "redisd"
const Log = varrun.Dir + "/log/redisd"

// Machines may restrict redisd listening to this list of net devices.
// If unset, the local admin may restrict this through /etc/default/goes ARGS.
// Otherwise, the default is all active net devices.
var Devs []string

// Machines may use this Hook to publish redis "key: value" strings before any
// other daemons are run.
var Hook = func(chan<- string) error { return nil }

// A non-empty Machine is published to redis as "machine: Machine"
var Machine string

// Admins may override the redis listening port through /etc/default/goes ARGS.
var Port = 6379

// Machines may override this list of published hashes.
var PublishedKeys = []string{redis.DefaultHash}

type cmd struct {
	redisd Redisd
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
func (*cmd) Usage() string   { return "redisd [-port PORT] [DEVICE]..." }

func (cmd *cmd) Main(args ...string) error {
	parm, args := parms.New(args, "-port")
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

	pub := make(chan string)
	go func(pub <-chan string) {
		for s := range pub {
			field := strings.Split(s, ": ")
			cmd.redisd.published[redis.DefaultHash][field[0]] =
				[]byte(field[1])
		}
	}(pub)
	pub <- fmt.Sprint("version: ", Version)
	if hostname, err := os.Hostname(); err == nil {
		pub <- fmt.Sprint("hostname: ", hostname)
	}
	if len(Machine) > 0 {
		pub <- fmt.Sprint("machine: ", Machine)
	}
	if keys, cl, err := cmdline.New(); err == nil {
		for _, k := range keys {
			pub <- fmt.Sprintf("cmdline.%s: %s", k, cl[k])
		}
	}
	err = Hook(pub)
	pub <- "redis.ready: true"
	close(pub)
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
	return err
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "a redis server",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	redisd - a redis server

SYNOPSIS
	redisd [-port PORT] [DEV]...

DESCRIPTION
	Run a redis server on the /run/goes/socks/redisd unix files socket and
	on all of the given network devices and the given or default port of
	6379.`,
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

func (redisd *Redisd) Del(key string, keys ...string) (int, error) {
	type t interface {
		Del(string, ...string) (int, error)
	}
	redisd.mutex.Lock()
	method, found := redisd.assignments.Find(key).(t)
	redisd.mutex.Unlock()
	if found {
		i, err := method.Del(key)
		if err == nil {
			redisd.mutex.Lock()
			redisd.assignments.Delete(key)
			redisd.mutex.Unlock()
		}
		return i, err
	}
	return 0, fmt.Errorf("can't del %s", key)
}

func (redisd *Redisd) Get(key string) ([]byte, error) {
	type t interface {
		Get(string) ([]byte, error)
	}
	f := func(key string) ([]byte, error) {
		return nil, fmt.Errorf("can't get %s", key)
	}
	redisd.mutex.Lock()
	if method, found := redisd.assignments.Find(key).(t); found {
		f = method.Get
	}
	redisd.mutex.Unlock()
	return f(key)
}

func (redisd *Redisd) Hdel(key, subkey string, subkeys ...string) (int, error) {
	type t interface {
		Hdel(string, string, ...string) (int, error)
	}
	f := func(key, subkey string, subkeys ...string) (int, error) {
		return 0, fmt.Errorf("can't hdel %s", key)
	}
	hashkey := fmt.Sprint(key, ":", subkey)
	redisd.mutex.Lock()
	if method, found := redisd.assignments.Find(hashkey).(t); found {
		f = method.Hdel
	} else if method, found := redisd.assignments.Find(key).(t); found {
		f = method.Hdel
	}
	redisd.mutex.Unlock()
	return f(key, subkey, subkeys...)
}

func (redisd *Redisd) Hexists(key, field string) (int, error) {
	type t interface {
		Hexists(string, string) (int, error)
	}
	f := func(key, field string) (int, error) {
		return 0, fmt.Errorf("can't hexists %s %s", key, field)
	}
	redisd.mutex.Lock()
	hashkey := fmt.Sprint(key, ":", field)
	if method, found := redisd.assignments.Find(hashkey).(t); found {
		f = method.Hexists
	} else if method, found := redisd.assignments.Find(key).(t); found {
		f = method.Hexists
	}
	redisd.mutex.Unlock()
	return f(key, field)
}

func (redisd *Redisd) Hget(key, subkey string) ([]byte, error) {
	type t interface {
		Hget(string, string) ([]byte, error)
	}
	f := func(key, subkey string) ([]byte, error) {
		return nil, fmt.Errorf("can't hget %s %s", key, subkey)
	}
	redisd.mutex.Lock()
	if hv, found := redisd.published[key]; found {
		defer redisd.mutex.Unlock()
		return hv[subkey], nil
	}
	hashkey := fmt.Sprint(key, ":", subkey)
	if method, found := redisd.assignments.Find(hashkey).(t); found {
		f = method.Hget
	} else if method, found := redisd.assignments.Find(key).(t); found {
		f = method.Hget
	}
	redisd.mutex.Unlock()
	return f(key, subkey)
}

func (redisd *Redisd) Hgetall(key string) ([][]byte, error) {
	type t interface {
		Hgetall(string) ([][]byte, error)
	}
	f := func(key string) ([][]byte, error) {
		return nil, fmt.Errorf("can't hgetall %s", key)
	}
	redisd.mutex.Lock()
	if hv, found := redisd.published[key]; found {
		subkeys := redisd.subkeys(key, hv)
		reply := make([][]byte, 0, len(hv)*2)
		for _, k := range subkeys {
			reply = append(reply, []byte(k), hv[k])
		}
		redisd.mutex.Unlock()
		return reply, nil
	}
	if method, found := redisd.assignments.Find(key).(t); found {
		f = method.Hgetall
	}
	redisd.mutex.Unlock()
	return f(key)
}

func (redisd *Redisd) Hkeys(key string) ([][]byte, error) {
	type t interface {
		Hkeys(string) ([][]byte, error)
	}
	f := func(key string) ([][]byte, error) {
		return nil, fmt.Errorf("can't hkeys %s", key)
	}
	redisd.mutex.Lock()
	if hv, found := redisd.published[key]; found {
		subkeys := redisd.subkeys(key, hv)
		reply := make([][]byte, 0, len(subkeys))
		for _, k := range subkeys {
			reply = append(reply, []byte(k))
		}
		redisd.mutex.Unlock()
		return reply, nil
	}
	if method, found := redisd.assignments.Find(key).(t); found {
		f = method.Hkeys
	}
	redisd.mutex.Unlock()
	return f(key)
}

func (redisd *Redisd) Hset(key, field string, value []byte) (int, error) {
	type t interface {
		Hset(string, string, []byte) (int, error)
	}
	f := func(key, field string, value []byte) (int, error) {
		return 0, fmt.Errorf("can't hset %s %s", key, field)
	}
	redisd.mutex.Lock()
	hashkey := fmt.Sprint(key, ":", field)
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

func (redisd *Redisd) Set(key string, value []byte) error {
	type t interface {
		Set(string, []byte) error
	}
	f := func(key string, value []byte) error {
		return fmt.Errorf("can't set %s", key)
	}
	redisd.mutex.Lock()
	if method, found := redisd.assignments.Find(key).(t); found {
		f = method.Set
	}
	redisd.mutex.Unlock()
	return f(key, value)
}

func (redisd *Redisd) Lrange(key string, start, stop int) ([][]byte, error) {
	type t interface {
		Lrange(string, int, int) ([][]byte, error)
	}
	f := func(key string, start, stop int) ([][]byte, error) {
		return nil, fmt.Errorf("can't lrange %s", key)
	}
	redisd.mutex.Lock()
	if method, found := redisd.assignments.Find(key).(t); found {
		f = method.Lrange
	}
	redisd.mutex.Unlock()
	return f(key, start, stop)
}

func (redisd *Redisd) Lindex(key string, index int) ([]byte, error) {
	type t interface {
		Lindex(string, int) ([]byte, error)
	}
	f := func(key string, index int) ([]byte, error) {
		return nil, fmt.Errorf("can't lindex %s", key)
	}
	redisd.mutex.Lock()
	if method, found := redisd.assignments.Find(key).(t); found {
		f = method.Lindex
	}
	redisd.mutex.Unlock()
	return f(key, index)
}

func (redisd *Redisd) Blpop(key string, keys ...string) ([][]byte, error) {
	type t interface {
		Blpop(string, ...string) ([][]byte, error)
	}
	f := func(key string, keys ...string) ([][]byte, error) {
		return nil, fmt.Errorf("can't blpop %s", key)
	}
	redisd.mutex.Lock()
	if method, found := redisd.assignments.Find(key).(t); found {
		f = method.Blpop
	}
	redisd.mutex.Unlock()
	return f(key, keys...)
}

func (redisd *Redisd) Brpop(key string, keys ...string) ([][]byte, error) {
	type t interface {
		Brpop(string, ...string) ([][]byte, error)
	}
	f := func(key string, keys ...string) ([][]byte, error) {
		return nil, fmt.Errorf("can't brpop %s", key)
	}
	redisd.mutex.Lock()
	if method, found := redisd.assignments.Find(key).(t); found {
		f = method.Brpop
	}
	redisd.mutex.Unlock()
	return f(key, keys...)
}

func (redisd *Redisd) Lpush(key string, value []byte, values ...[]byte) (int, error) {
	type t interface {
		Lpush(string, []byte, ...[]byte) (int, error)
	}
	f := func(key string, value []byte, values ...[]byte) (int, error) {
		return 0, fmt.Errorf("can't lpush %s", key)
	}
	redisd.mutex.Lock()
	if method, found := redisd.assignments.Find(key).(t); found {
		f = method.Lpush
	}
	redisd.mutex.Unlock()
	return f(key, value, values...)
}

func (redisd *Redisd) Rpush(key string, value []byte, values ...[]byte) (int, error) {
	type t interface {
		Rpush(string, []byte, ...[]byte) (int, error)
	}
	f := func(key string, value []byte, values ...[]byte) (int, error) {
		return 0, fmt.Errorf("can't rpush %s", key)
	}
	redisd.mutex.Lock()
	if method, found := redisd.assignments.Find(key).(t); found {
		f = method.Rpush
	}
	redisd.mutex.Unlock()
	return f(key, value, values...)
}

func (redisd *Redisd) Monitor() (*grs.MonitorReply, error) {
	// FIXME
	return &grs.MonitorReply{}, nil
}

func (redisd *Redisd) Ping() (*grs.StatusReply, error) {
	return grs.NewStatusReply("PONG"), nil
}

func (redisd *Redisd) Publish(key string, value []byte) (int, error) {
	redisd.mutex.Lock()
	if hv, found := redisd.published[key]; found {
		fields := bytes.Split(value, []byte(": "))
		if bytes.Compare(fields[0], []byte("delete")) == 0 {
			delete(hv, string(fields[1]))
		} else {
			hv[string(fields[0])] = fields[1]
		}
		redisd.flushSubkeyCache(key)
	}
	cws, found := redisd.sub[key]
	redisd.mutex.Unlock()
	if !found || len(cws) == 0 {
		return 0, nil
	}
	msg := []interface{}{
		"message",
		key,
		value,
	}
	i := 0
	for _, cw := range cws {
		select {
		case cw.Channel <- msg:
			i++
		default:
		}
	}
	return i, nil
}

func (redisd *Redisd) Select(key string) error {
	type t interface {
		Select(string) error
	}
	f := func(key string) error {
		return fmt.Errorf("can't select %s", key)
	}
	redisd.mutex.Lock()
	if method, found := redisd.assignments.Find(key).(t); found {
		f = method.Select
	}
	redisd.mutex.Unlock()
	return f(key)
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
