// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vnet

import (
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/elib/dep"
	"github.com/platinasystems/go/elib/parse"

	"fmt"
	"os"
)

type Packager interface {
	GetPackage() *Package
	Configure(in *parse.Input)
	Init() (err error)
	Exit() (err error)
}

const (
	// Dependencies: packages this package depends on.
	forward = iota
	// Anti-dependencies: packages that are dependent on this package.
	anti
	nDepType
)

type Package struct {
	Vnet *Vnet
	name string

	depMap [nDepType]map[string]struct{}

	dep dep.Dep
}

func (p *Package) GetPackage() *Package { return p }
func (p *Package) Init() (err error)    { return } // likely overridden
func (p *Package) Exit() (err error)    { return } // likely overridden
func (p *Package) Configure(in *parse.Input) {
	panic(cli.ParseError)
}

type packageMain struct {
	packageByName parse.StringMap
	packages      []Packager
	deps          dep.Deps
}

func (p *Package) addDep(name string, typ int) {
	if len(name) == 0 {
		panic("empty dependency")
	}
	if p.depMap[typ] == nil {
		p.depMap[typ] = make(map[string]struct{})
	}
	p.depMap[typ][name] = struct{}{}
}
func (p *Package) DependsOn(names ...string) {
	for i := range names {
		p.addDep(names[i], forward)
	}
}
func (p *Package) DependedOnBy(names ...string) {
	for i := range names {
		p.addDep(names[i], anti)
	}
}

func (v *Vnet) AddPackage(name string, r Packager) (pi uint) {
	m := &v.packageMain
	// Package with index zero is always empty.
	// Protects against uninitialized package index variables.
	if len(m.packages) == 0 {
		m.packages = append(m.packages, &Package{name: "(empty)"})
	}

	// Already registered
	var ok bool
	if pi, ok = m.packageByName[name]; ok {
		return
	}

	pi = uint(len(m.packages))
	m.packageByName.Set(name, pi)
	m.packages = append(m.packages, r)
	p := r.GetPackage()
	p.name = name
	p.Vnet = v
	return
}

func (m *packageMain) GetPackage(i uint) Packager { return m.packages[i] }
func (m *packageMain) PackageByName(name string) (i uint, ok bool) {
	i, ok = m.packageByName[name]
	return
}

func (p *Package) configure(r Packager, in *parse.Input) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("configure %s: %s: %s", p.name, e, in)
		}
	}()
	r.Configure(in)
	return
}

func (v *Vnet) ConfigurePackages(in *parse.Input) (err error) {
	m := &v.packageMain
	// Parse package configuration.
	for !in.End() {
		var (
			i     uint
			subIn parse.Input
		)
		switch {
		case in.Parse("%v %v", m.packageByName, &i, &subIn):
			r := m.packages[i]
			p := r.GetPackage()
			if err = p.configure(r, &subIn); err != nil {
				return
			}
		case in.Parse("vnet %v", &subIn):
			if err = v.Configure(&subIn); err != nil {
				return
			}
		default:
			err = fmt.Errorf("%s: %s", parse.ErrInput, in)
			return
		}
	}
	return
}

func (v *Vnet) Configure(in *parse.Input) (err error) {
	for !in.End() {
		var logFile string
		switch {
		case in.Parse("log %v", &logFile):
			var f *os.File
			f, err = os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0644)
			if err != nil {
				return
			}
			v.loop.Config.LogWriter = f
		case in.Parse("quit %f", &v.loop.Config.QuitAfterDuration):
		case in.Parse("quit"):
			v.loop.Config.QuitAfterDuration = 1e-6 // must be positive to enable
		default:
			err = fmt.Errorf("%s: %s", parse.ErrInput, in)
			return
		}
	}
	return
}

func (m *packageMain) InitPackages() (err error) {
	// Resolve package dependencies.
	for i := range m.packages {
		p := m.packages[i].GetPackage()
		for typ := range p.depMap {
			for name := range p.depMap[typ] {
				if j, ok := m.packageByName[name]; ok {
					d := m.packages[j].GetPackage()
					if typ == forward {
						p.dep.Deps = append(p.dep.Deps, &d.dep)
					} else {
						p.dep.AntiDeps = append(p.dep.AntiDeps, &d.dep)
					}
				} else {
					panic(fmt.Errorf("%s: unknown dependent package `%s'", p.name, name))
				}
			}
		}
		m.deps.Add(&p.dep)
	}

	// Call package init functions.
	for i := range m.packages {
		p := m.packages[m.deps.Index(i)]
		err = p.Init()
		if err != nil {
			return
		}
	}
	return
}

func (m *packageMain) ExitPackages() (err error) {
	for i := range m.packages {
		p := m.packages[m.deps.IndexReverse(i)]
		err = p.Exit()
		if err != nil {
			return
		}
	}
	return
}
