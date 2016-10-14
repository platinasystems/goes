// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package machined

import (
	"fmt"
	"log"
	"net/rpc"
	"sort"
	"strings"

	"github.com/platinasystems/go/recovered"
	"github.com/platinasystems/go/redis"
	"github.com/platinasystems/go/redis/rpc/args"
	"github.com/platinasystems/go/redis/rpc/reply"
	"github.com/platinasystems/go/sch"
	"github.com/platinasystems/go/sockfile"
	"github.com/platinasystems/goes"
)

const debugRegistry = false

// Machines may modify this info provider list or overwrite any of its
// references with a target specific implementation.
var InfoProviders = []InfoProvider{
	Hostname,
	NetLink,
	Uptime,
	Version,
	Test,
}

var pub sch.In

var Hook = func() {}

func Publish(key, value interface{}) {
	pub <- fmt.Sprint(key, ": ", value)
}

type InfoProvider interface {
	// Provider should return a list of longest match keys supported by
	// info provider
	Prefixes(...string) []string

	// Main should detect and report machine changes like this until Close.
	//
	//	Publish(KEY, VALUE)
	//
	// or, if device or attribute is removed,
	//
	//	Publish("delete", KEY)
	//
	Main(...string) error

	// Close should stop all info go-routines and release all resources.
	Close() error

	// Del should remove the attribute then publish with,
	//
	//	Publish("delete", KEY)
	Del(string) error

	// Set should assign the given machine attribute then publish the new
	// value with,
	//
	//	Publish(KEY, VALUE)
	//
	Set(string, string) error

	// String should return the provider name
	String() string
}

type Attrs map[string]interface{}
type Registry []*entry

type entry struct {
	prefix string
	info   InfoProvider
}

type machined struct {
	registry Registry
}

func New() *machined { return &machined{} }

func (*machined) String() string { return "machined" }
func (*machined) Usage() string  { return "machined" }
func (*machined) Daemon() int    { return 1 }

func (p *machined) Main(args ...string) error {
	var i, n int

	goes.Standby("machined")
	Hook()
	for _, info := range InfoProviders {
		n += len(info.Prefixes())
	}
	p.registry = make(Registry, n)
	for _, info := range InfoProviders {
		for _, prefix := range info.Prefixes() {
			p.registry[i] = &entry{prefix, info}
			i++
		}
	}
	sort.Sort(p)
	rpc.Register(p.registry)
	sock, err := sockfile.NewRpcServer("machined")
	if err != nil {
		return err
	}
	pub, err = redis.Publish("platina")
	if err != nil {
		return err
	}
	defer close(pub)
	for i, entry := range p.registry {
		key := fmt.Sprint("platina:", entry.prefix)
		redis.Assign(key, "machined", "Registry")
		if debugRegistry {
			Publish(fmt.Sprintf("machined.%03d", i),
				entry.prefix)
		}
	}
	for _, info := range InfoProviders {
		if info.Main != nil {
			go func(info InfoProvider) {
				func() {
					err := recovered.New(info).Main()
					if err != nil {
						log.Print("daemon", "err", err)
					}
				}()
			}(info)
		}
	}
	sock.Wait()
	for _, info := range InfoProviders {
		if info.Close != nil {
			info.Close()
		}
	}
	return nil
}

// Interface to sort machine info registry
func (p *machined) Len() int {
	return len(p.registry)
}

func (p *machined) Less(i, j int) (t bool) {
	ni, nj := len(p.registry[i].prefix), len(p.registry[j].prefix)
	switch {
	case ni < nj:
		t = true
	case ni > nj:
		t = false
	case ni == nj:
		t = p.registry[i].prefix < p.registry[j].prefix
	default:
		panic("oops")
	}
	return t
}

func (p *machined) Swap(i, j int) {
	p.registry[i], p.registry[j] = p.registry[j], p.registry[i]
}

func (r Registry) Hdel(args args.Hdel, reply *reply.Hdel) error {
	for _, entry := range r {
		if strings.HasPrefix(args.Field, entry.prefix) {
			if entry.info.Del == nil {
				break
			}
			err := entry.info.Del(args.Field)
			if err == nil {
				*reply = 1
			}
			return err
		}
	}
	return fmt.Errorf("can't delete %s", args.Field)
}

func (r Registry) Hset(args args.Hset, reply *reply.Hset) error {
	value := string(args.Value)
	for _, entry := range r {
		if strings.HasPrefix(args.Field, entry.prefix) {
			if entry.info.Set == nil {
				break
			}
			err := entry.info.Set(args.Field, value)
			if err == nil {
				*reply = 1
			}
			return err
		}
	}
	return fmt.Errorf("can't set %s", args.Field)
}

func CantDel(key string) error {
	return fmt.Errorf("can't delete %s", key)
}

func CantSet(key string) error {
	return fmt.Errorf("can't set %s", key)
}
