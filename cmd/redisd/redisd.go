// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
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
	"regexp"
	"sort"
	"strings"
	"sync"

	grs "github.com/platinasystems/go-redis-server"
	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/external/atsock"
	"github.com/platinasystems/goes/external/parms"
	"github.com/platinasystems/goes/external/redis"
	"github.com/platinasystems/goes/external/redis/publisher"
	"github.com/platinasystems/goes/external/redis/rpc/reg"
	"github.com/platinasystems/goes/internal/cmdline"
	"github.com/platinasystems/goes/internal/fields"
	"github.com/platinasystems/goes/lang"
)

type Command struct {
	// Machines may restrict redisd listening to this list of net devices.
	// If unset, the local admin may restrict this through
	// /etc/default/goes ARGS.  Otherwise, the default is all active net
	// devices.
	Devs []string

	// Machines may use this Hook to Print redis "[key: ]field: value"
	// strings before any other daemons are run.
	Hook func(*publisher.Publisher)

	// A non-empty Machine is published to redis as "machine: Machine"
	Machine string

	// default: 6379
	Port int

	// Machines may override this list of published hashes.
	// default: redis.DefaultHash
	PublishedKeys []string

	pubconn *net.UnixConn
	redisd  Redisd
}

func (*Command) String() string { return "redisd" }

func (*Command) Usage() string {
	return "redisd [-port PORT] [-set FIELD=VALUE]... [DEVICE]..."
}

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "a redis server",
	}
}

func (*Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
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

func (*Command) Kind() cmd.Kind { return cmd.Daemon }

func (c *Command) Main(args ...string) error {
	parm, args := parms.New(args, "-port", "-set")
	if s := parm.ByName["-port"]; len(s) > 0 {
		if _, err := fmt.Sscan(s, &c.Port); err != nil {
			return err
		}
	} else {
		c.Port = 6379
	}
	c.redisd.port = c.Port

	if len(args) == 0 {
		if len(c.Devs) == 0 {
			itfs, ierr := net.Interfaces()
			if ierr == nil {
				args = make([]string, len(itfs))
				for i, itf := range itfs {
					args[i] = itf.Name
				}
			}
		} else {
			args = c.Devs
		}
	}

	if false {
		grs.Debugf = grs.ActualDebugf
	} else {
		grs.Stderr = os.Stderr
	}

	c.redisd.devs = make(map[string][]*Server)
	c.redisd.sub = make(map[string]*grs.MultiChannelWriter)
	c.redisd.published = make(grs.HashHash)
	if len(c.PublishedKeys) == 0 {
		c.PublishedKeys = []string{redis.DefaultHash}
	}
	for _, k := range c.PublishedKeys {
		c.redisd.published[k] = make(grs.HashValue)
	}

	cfg := grs.DefaultConfig()
	cfg = cfg.Proto("unix")
	cfg = cfg.Host("@redisd")
	cfg = cfg.Handler(&c.redisd)

	srv, err := grs.NewServer(cfg)
	if err != nil {
		return err
	}

	c.redisd.devs["@redisd"] = []*Server{{server: srv}}

	c.redisd.reg, err = reg.New(c.redisd.assign, c.redisd.unassign)
	if err != nil {
		return err
	}

	c.pubconn, err = atsock.ListenUnixgram("redis.pub")
	if err != nil {
		return err
	}
	goes.WG.Add(1)
	go func() {
		defer goes.WG.Done()
		c.gopub()
	}()

	err = c.pubinit(fields.New(parm.ByName["-set"])...)
	if err != nil {
		return err
	}

	goes.WG.Add(1)
	go func() {
		defer goes.WG.Done()
		srv.Start()
	}()

	goes.WG.Add(1)
	go func(redisd *Redisd, args ...string) {
		defer goes.WG.Done()
		for _, name := range args {
			select {
			case <-goes.Stop:
				return
			default:
				redisd.listenOnInterface(name)
			}
		}
	}(&c.redisd, args...)

	<-goes.Stop

	if c.redisd.reg != nil {
		c.redisd.reg.Srvr.Close()
	}
	if c.pubconn != nil {
		c.pubconn.Close()
	}

	c.redisd.mutex.Lock()
	for k, srvs := range c.redisd.devs {
		for i, srv := range srvs {
			srv.server.Close()
			srvs[i] = nil
		}
		c.redisd.devs[k] = c.redisd.devs[k][:0]
		delete(c.redisd.devs, k)
	}
	c.redisd.mutex.Unlock()

	return nil
}

func (c *Command) gopub() {
	const sep = ": "
	var key, field string
	var fv, value []byte
	b := make([]byte, os.Getpagesize())
	for {
		n, err := c.pubconn.Read(b)
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
		c.redisd.mutex.Lock()
		hv, found := c.redisd.published[key]
		if !found {
			hv = make(grs.HashValue)
			c.redisd.published[key] = hv
		}
		if field == "delete" {
			for k := range hv {
				if strings.HasPrefix(k, string(value)) {
					delete(hv, k)
				}
			}
		} else {
			_, found := hv[field]
			if !found {
				hv[field] = make([]byte, 0, 256)
			} else {
				hv[field] = hv[field][:0]
			}
			hv[field] = append(hv[field], value...)
			if sub, found := c.redisd.sub[key]; found {
				mb := make([]byte, len(fv))
				copy(mb, fv)
				msg := make([]interface{}, 3)
				msg[0] = "message"
				msg[1] = key
				msg[2] = mb
				for i := 0; i < len(sub.Chans); {
					select {
					case sub.Chans[i].Channel <- msg:
						i++
					default:
						// cull this subscriber
						close(sub.Chans[i].Channel)
						n := len(sub.Chans) - 1
						if i != n {
							copy(sub.Chans[i:],
								sub.Chans[i+1:])
						}
						sub.Chans[n] = nil
						sub.Chans = sub.Chans[:n]
					}
				}
			}
		}
		c.redisd.flushSubkeyCache(key)
		c.redisd.mutex.Unlock()
	}
}

func (c *Command) pubinit(fieldEqValues ...string) error {
	pub, err := publisher.New()
	if err != nil {
		return err
	}
	defer pub.Close()

	if hostname, err := os.Hostname(); err == nil {
		pub.Print("hostname: ", hostname)
	}
	if len(c.Machine) > 0 {
		pub.Print("machine: ", c.Machine)
	}
	if keys, cl, err := cmdline.New(); err == nil {
		for _, k := range keys {
			pub.Printf("cmdline.%s: %s", k, cl[k])
		}
	}

	if c.Hook != nil {
		c.Hook(pub)
	}

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

	_, err = pub.Print("redis.ready: true")
	return err
}

type Server struct {
	addr   string
	server *grs.Server
}

type Redisd struct {
	mutex sync.Mutex
	devs  map[string][]*Server
	sub   map[string]*grs.MultiChannelWriter

	reg *reg.Reg

	assignments Assignments

	published grs.HashHash

	cachedKeys    []string
	cachedSubkeys map[string][]string

	port int
}

type Assignments []*assignment

type assignment struct {
	prefix string
	v      interface{}
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

func (redisd *Redisd) findServerOnInterface(name string, addr string) (found bool, index int) {
	for i, srv := range redisd.devs[name] {
		if srv.addr == addr {
			return true, i
		}
	}
	return false, 0
}

func (redisd *Redisd) listenOnInterface(name string) {
	redisd.mutex.Lock()
	defer redisd.mutex.Unlock()

	ok := true
	dev, err := net.InterfaceByName(name)
	if err != nil {
		fmt.Fprint(os.Stderr, name, ": ", err, "\n")
		ok = false
	}

	if ok && ((dev.Flags & net.FlagUp) != net.FlagUp) {
		fmt.Fprint(os.Stderr, name, ": down\n")
		ok = false
	}

	var addrs []net.Addr
	if ok {
		addrs, err = dev.Addrs()
		if err != nil {
			fmt.Fprint(os.Stderr, name, ": ", err, "\n")
			ok = false
		} else {
			if len(addrs) == 0 {
				fmt.Fprint(os.Stderr, name,
					": no address or isn't up\n")
				ok = false
			}
		}
	}

	srvs := make([]*Server, 0, len(redisd.devs[name])+2)

devloop:
	for i, srv := range redisd.devs[name] {
		for _, addr := range addrs {
			if srv.addr == addr.String() {
				srvs = append(srvs, srv)
				continue devloop
			}
		}
		if true {
			fmt.Fprint(os.Stderr, srv.addr, ": removed from ",
				name)
		}
		srv.server.Close()
		redisd.devs[name][i] = nil
	}

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
		if found, _ := redisd.findServerOnInterface(name, addr.String()); found {
			if false {
				fmt.Fprint(os.Stderr, addr, ": already up on: ",
					name, "\n")
			}
			continue
		}
		id := fmt.Sprint("[", ip, "%", name, "]:", redisd.port)
		cfg := grs.DefaultConfig()
		cfg = cfg.Handler(redisd)
		cfg = cfg.Port(redisd.port)
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
			srvs = append(srvs, &Server{server: srv,
				addr: addr.String()})
			goes.WG.Add(1)
			go func() {
				defer goes.WG.Done()
				srv.Start()
			}()
			if true {
				fmt.Println("listen:", id)
			}
		}
	}
	redisd.devs[name] = srvs
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
		for k := range redisd.published {
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
			Channel: make(chan []interface{}, 1024),
		}
		if sub := redisd.sub[string(key)]; sub == nil {
			redisd.sub[string(key)] = &grs.MultiChannelWriter{
				Chans: []*grs.ChannelWriter{cw},
			}
		} else {
			sub.Chans = append(sub.Chans, cw)
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
