// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package internal

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"syscall"

	grs "github.com/platinasystems/go-redis-server"
	"github.com/platinasystems/go/emptych"
	"github.com/platinasystems/go/group"
	"github.com/platinasystems/go/sockfile"
)

const VarLogRedisd = "/var/log/redisd"
const debugAssignments = false

type Redisd struct {
	mutex sync.Mutex
	devs  map[string][]*grs.Server
	sig   chan os.Signal
	done  emptych.Out
	sub   grs.HashSub

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

var PublishedKeys = []string{"platina"}

func (p *Redisd) main() error {
	varlog := filepath.Dir(VarLogRedisd)
	if _, err := os.Stat(varlog); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(varlog, os.FileMode(0755))
		}
		if err != nil {
			return err
		}
	}

	if false {
		grs.Debugf = grs.ActualDebugf
	} else if w, err := os.Create(VarLogRedisd); err != nil {
		return err
	} else {
		grs.Stderr = w
	}

	p.devs = make(map[string][]*grs.Server)
	p.sig = make(chan os.Signal)
	p.sub = make(grs.HashSub)
	p.published = make(grs.HashHash)
	for _, k := range PublishedKeys {
		p.published[k] = make(grs.HashValue)
	}

	signal.Notify(p.sig, syscall.SIGTERM)

	cfg := grs.DefaultConfig()
	cfg = cfg.Proto("unix")
	cfg = cfg.Host(sockfile.Path("redisd"))
	cfg = cfg.Handler(p)

	srv, err := grs.NewServer(cfg)
	if err != nil {
		return err
	}
	sfn := sockfile.Path("redisd")
	p.devs[sfn] = []*grs.Server{srv}
	go srv.Start()

	if adm := group.Parse()["adm"].Gid(); adm > 0 {
		go func(sfn string, adm int) {
			for {
				_, err := os.Stat(sfn)
				if err == nil {
					break
				}
			}
			os.Chown(sfn, os.Geteuid(), adm)
		}(sfn, adm)
	}

	stop := emptych.Make()
	p.done = emptych.Out(stop)
	go func(stop emptych.In) {
		for sig := range p.sig {
			if sig == syscall.SIGTERM {
				break
			}
		}

		p.mutex.Lock()
		for k, srvs := range p.devs {
			for i, srv := range srvs {
				srv.Close()
				srvs[i] = nil
			}
			p.devs[k] = p.devs[k][:0]
			delete(p.devs, k)
		}
		p.mutex.Unlock()

		// FIXME close subscribers channels?

		stop.Close()
	}(emptych.In(stop))

	return nil
}

func (p *Redisd) Wait() {
	p.done.Wait()
}

func (p *Redisd) listen(dev, port string) (int, error) {
	if len(port) == 0 {
		port = "6379"
	}
	var iport int
	_, err := fmt.Sscan(port, &iport)
	if err != nil {
		return 0, err
	}

	p.mutex.Lock()
	_, found := p.devs[dev]
	if !found {
		// place holder
		p.devs[dev] = []*grs.Server{}
	}
	p.mutex.Unlock()
	if found {
		return 0, fmt.Errorf("%s: already running redisd", dev)
	}

	netdev, err := net.InterfaceByName(dev)
	if err != nil {
		return 0, err
	}
	addrs, err := netdev.Addrs()
	if err != nil {
		return 0, err
	}
	if len(addrs) == 0 {
		return 0, fmt.Errorf("%s: no address or isn't up", dev)
	}

	srvs := make([]*grs.Server, 0, 2)
	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			fmt.Fprintln(os.Stderr, "redisd: ", addr.String(),
				": CIDR: ", err)
			continue
		}
		if ip.IsMulticast() {
			continue
		}
		id := fmt.Sprint("[", ip, "%", dev, "]:", port)
		cfg := grs.DefaultConfig()
		cfg = cfg.Handler(p)
		cfg = cfg.Port(iport)
		if ip.To4() == nil {
			cfg = cfg.Proto("tcp6")
			cfg = cfg.Host(fmt.Sprint("[", ip, "%", dev, "]"))
		} else {
			cfg = cfg.Host(ip.String())
		}
		srv, err := grs.NewServer(cfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "redisd: ", id, ": ", err)
		} else {
			fmt.Println("redisd serve: ", id)
			srvs = append(srvs, srv)
			go srv.Start()
		}
	}

	p.mutex.Lock()
	p.devs[dev] = srvs
	p.mutex.Unlock()

	return 1, nil
}

func (p *Redisd) flushKeyCache() {
	p.cachedKeys = p.cachedKeys[:0]
}

func (p *Redisd) flushSubkeyCache(key string) {
	if p.cachedSubkeys == nil {
		return
	}
	a, found := p.cachedSubkeys[key]
	if found {
		p.cachedSubkeys[key] = a[:0]
	}
}

func (p *Redisd) Handler(key string) interface{} {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.assignments.Find(key)
}

func (p *Redisd) Del(key string, keys ...string) (int, error) {
	type t interface {
		Del(string, ...string) (int, error)
	}
	if method, found := p.Handler(key).(t); found {
		i, err := method.Del(key)
		if err == nil {
			p.mutex.Lock()
			p.assignments.Delete(key)
			p.mutex.Unlock()
		}
		return i, err
	}
	return 0, fmt.Errorf("can't del %s", key)
}

func (p *Redisd) Get(key string) ([]byte, error) {
	type t interface {
		Get(string) ([]byte, error)
	}
	if method, found := p.Handler(key).(t); found {
		return method.Get(key)
	}
	return nil, fmt.Errorf("can't get %s", key)
}

func (p *Redisd) Hdel(key, subkey string, subkeys ...string) (int, error) {
	type t interface {
		Hdel(string, string, ...string) (int, error)
	}
	hashkey := fmt.Sprint(key, ":", subkey)
	if method, found := p.Handler(hashkey).(t); found {
		return method.Hdel(key, subkey, subkeys...)
	}
	if method, found := p.Handler(key).(t); found {
		return method.Hdel(key, subkey, subkeys...)
	}
	if key != "redisd" {
		return 0, fmt.Errorf("can't hdel %s", key)
	}
	closed := 0
	if p.devs == nil {
		return closed, nil
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for _, dev := range append([]string{subkey}, subkeys...) {
		if lns, found := p.devs[dev]; found {
			for _, ln := range lns {
				ln.Close()
			}
			closed += 1
			lns = lns[:0]
			delete(p.devs, dev)
			fmt.Println("redisd: ", dev, ": closed")
		}
	}
	return closed, nil
}

func (p *Redisd) Hexists(key, field string) (int, error) {
	type t interface {
		Hexists(string, string) (int, error)
	}
	hashkey := fmt.Sprint(key, ":", field)
	if method, found := p.Handler(hashkey).(t); found {
		return method.Hexists(key, field)
	}
	if method, found := p.Handler(key).(t); found {
		return method.Hexists(key, field)
	}
	if key != "redisd" {
		return 0, fmt.Errorf("can't hexists %s", key)
	}
	if p.devs == nil {
		return 0, nil
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()
	i := 0
	if _, found := p.devs[field]; found {
		i = 1
	}
	return i, nil
}

func (p *Redisd) Hget(key, subkey string) ([]byte, error) {
	hv, found := p.hv(key)
	if found {
		p.mutex.Lock()
		defer p.mutex.Unlock()
		return hv[subkey], nil
	}
	type t interface {
		Hget(string, string) ([]byte, error)
	}
	hashkey := fmt.Sprint(key, ":", subkey)
	if method, found := p.Handler(hashkey).(t); found {
		return method.Hget(key, subkey)
	}
	if method, found := p.Handler(key).(t); found {
		return method.Hget(key, subkey)
	}
	return nil, fmt.Errorf("can't hget %s %s", key, subkey)
}

func (p *Redisd) Hgetall(key string) ([][]byte, error) {
	if debugAssignments && key == "platina" {
		for i, as := range p.assignments {
			hv := p.published["platina"]
			k := fmt.Sprintf("prefix.%03d", i)
			hv[k] = []byte(as.prefix)
		}
	}
	hv, found := p.hv(key)
	if found {
		subkeys := p.subkeys(key, hv)
		p.mutex.Lock()
		defer p.mutex.Unlock()
		reply := make([][]byte, 0, len(hv)*2)
		for _, k := range subkeys {
			reply = append(reply, []byte(k), hv[k])
		}
		return reply, nil
	}
	type t interface {
		Hgetall(string) ([][]byte, error)
	}
	if method, found := p.Handler(key).(t); found {
		return method.Hgetall(key)
	}
	return nil, fmt.Errorf("can't hgetall %s", key)
}

func (p *Redisd) Hkeys(key string) ([][]byte, error) {
	hv, found := p.hv(key)
	if found {
		subkeys := p.subkeys(key, hv)
		reply := make([][]byte, 0, len(subkeys))
		for _, k := range subkeys {
			reply = append(reply, []byte(k))
		}
		return reply, nil
	}
	type t interface {
		Hkeys(string) ([][]byte, error)
	}
	if method, found := p.Handler(key).(t); found {
		return method.Hkeys(key)
	}
	if key != "redisd" {
		return nil, fmt.Errorf("can't hkeys %s", key)
	}
	if p.devs == nil {
		return nil, nil
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()
	keys := make([][]byte, 0, len(p.devs))
	for k := range p.devs {
		keys = append(keys, []byte(k))
	}
	return keys, nil
}

func (p *Redisd) Hset(key, field string, value []byte) (int, error) {
	type t interface {
		Hset(string, string, []byte) (int, error)
	}
	hashkey := fmt.Sprint(key, ":", field)
	if method, found := p.Handler(hashkey).(t); found {
		return method.Hset(key, field, value)
	}
	if method, found := p.Handler(key).(t); found {
		return method.Hset(key, field, value)
	}
	if key != "redisd" {
		return 0, fmt.Errorf("can't hset %s %s", key, field)
	}
	return p.listen(field, string(value))
}

func (p *Redisd) hv(key string) (hv grs.HashValue, found bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	hv, found = p.published[key]
	return
}

func (p *Redisd) Keys(pattern string) ([][]byte, error) {
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
	keys := p.keys()
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

func (p *Redisd) keys() []string {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if len(p.cachedKeys) == 0 {
		for _, a := range p.assignments {
			k := a.prefix
			if i := strings.Index(k, ":"); i > 0 {
				k = k[:i]
			}
			p.cachedKeys = append(p.cachedKeys, k)
		}
		for _, k := range PublishedKeys {
			p.cachedKeys = append(p.cachedKeys, k)
		}
		sort.Strings(p.cachedKeys)
	}
	return p.cachedKeys
}

func (p *Redisd) Set(key string, value []byte) error {
	type t interface {
		Set(string, []byte) error
	}
	if method, found := p.Handler(key).(t); found {
		return method.Set(key, value)
	}
	return fmt.Errorf("can't set %s", key)
}

func (p *Redisd) Lrange(key string, start, stop int) ([][]byte, error) {
	type t interface {
		Lrange(string, int, int) ([][]byte, error)
	}
	if method, found := p.Handler(key).(t); found {
		return method.Lrange(key, start, stop)
	}
	return nil, fmt.Errorf("can't lrange %s", key)
}

func (p *Redisd) Lindex(key string, index int) ([]byte, error) {
	type t interface {
		Lindex(string, int) ([]byte, error)
	}
	if method, found := p.Handler(key).(t); found {
		return method.Lindex(key, index)
	}
	return nil, fmt.Errorf("can't lindex %s", key)
}

func (p *Redisd) Blpop(key string, keys ...string) ([][]byte, error) {
	type t interface {
		Blpop(string, ...string) ([][]byte, error)
	}
	if method, found := p.Handler(key).(t); found {
		return method.Blpop(key, keys...)
	}
	return nil, fmt.Errorf("can't blpop %s", key)
}

func (p *Redisd) Brpop(key string, keys ...string) ([][]byte, error) {
	type t interface {
		Brpop(string, ...string) ([][]byte, error)
	}
	if method, found := p.Handler(key).(t); found {
		return method.Brpop(key, keys...)
	}
	return nil, fmt.Errorf("can't brpop %s", key)
}

func (p *Redisd) Lpush(key string, value []byte, values ...[]byte) (int,
	error) {
	type t interface {
		Lpush(string, []byte, ...[]byte) (int, error)
	}
	if method, found := p.Handler(key).(t); found {
		return method.Lpush(key, value, values...)
	}
	return 0, fmt.Errorf("can't lpush %s", key)
}

func (p *Redisd) Rpush(key string, value []byte, values ...[]byte) (int,
	error) {
	type t interface {
		Rpush(string, []byte, ...[]byte) (int, error)
	}
	if method, found := p.Handler(key).(t); found {
		return method.Rpush(key, value, values...)
	}
	return 0, fmt.Errorf("can't rpush %s", key)
}

func (p *Redisd) Monitor() (*grs.MonitorReply, error) {
	// FIXME
	return &grs.MonitorReply{}, nil
}

func (p *Redisd) Ping() (*grs.StatusReply, error) {
	return grs.NewStatusReply("PONG"), nil
}

func (p *Redisd) Publish(key string, value []byte) (int, error) {
	p.mutex.Lock()
	if hv, found := p.published[key]; found {
		fields := bytes.Split(value, []byte(": "))
		if bytes.Compare(fields[0], []byte("delete")) == 0 {
			delete(hv, string(fields[1]))
		} else {
			hv[string(fields[0])] = fields[1]
		}
		p.flushSubkeyCache(key)
	}
	cws, found := p.sub[key]
	p.mutex.Unlock()
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

func (p *Redisd) Select(key string) error {
	type t interface {
		Select(string) error
	}
	if method, found := p.Handler(key).(t); found {
		return method.Select(key)
	}
	return fmt.Errorf("can't select %s", key)
}

func (p *Redisd) Subscribe(channels ...[]byte) (*grs.MultiChannelWriter,
	error) {
	mcw := &grs.MultiChannelWriter{
		Chans: make([]*grs.ChannelWriter, len(channels)),
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	for i, key := range channels {
		cw := &grs.ChannelWriter{
			FirstReply: []interface{}{
				"subscribe",
				key,
				1,
			},
			Channel: make(chan []interface{}, 128),
		}
		if p.sub[string(key)] == nil {
			p.sub[string(key)] = []*grs.ChannelWriter{cw}
		} else {
			p.sub[string(key)] = append(p.sub[string(key)], cw)
		}
		mcw.Chans[i] = cw
	}
	return mcw, nil
}

func (p *Redisd) subkeys(key string, hv grs.HashValue) []string {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.cachedSubkeys == nil {
		p.cachedSubkeys = make(map[string][]string)
	}
	subkeys, found := p.cachedSubkeys[key]
	if !found {
		subkeys = []string{}
	}
	if len(subkeys) != len(hv) {
		subkeys = subkeys[:0]
		for k := range hv {
			subkeys = append(subkeys, k)
		}
		sort.Strings(subkeys)
		p.cachedSubkeys[key] = subkeys
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
