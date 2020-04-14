// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package parms

import (
	"errors"
	"strings"
)

var ErrNotFound = errors.New("not found")

type Parms struct {
	ByName  ByName
	aliases Aliases
}

type ByName map[string]string
type Aliases map[string]string

// Define and parse {NAME{=|' '}VALUE} parameters from command arguments, e.g.
//
//	parm, args := parms.New([]string{ "-a", "A", "-b=B", "something"},
//		"-a", "-b", "-c")
//
// results in
//
//	parm.ByName["-a"] == "A"
//	parm.ByName["-b"] == "B"
//	parm.ByName["-c"] == ""
//	args == []string{"something"}
//
// Parmeters may be defined with strings or string slices that include aliases
// of the first entry, e.g.
//
//	parm, args := parms.New([]string{"-color", "blue},
//		"-a", "-b", []string{"-c", "-color", "-colour"})
//
// results in
//
//	parm.ByName["-a"] == ""
//	parm.ByName["-b"] == ""
//	parm.ByName["-c"] == "blue"
//	args == []string{}
func New(args []string, parms ...interface{}) (*Parms, []string) {
	p := &Parms{
		ByName:  make(ByName),
		aliases: make(Aliases),
	}
	if len(parms) > 0 {
		args = p.More(args, parms...)
	}
	return p, args
}

// Define and parse more parameters from command arguments.
func (p *Parms) More(args []string, parms ...interface{}) []string {
	for _, v := range parms {
		switch t := v.(type) {
		case string:
			p.ByName[t] = ""
		case []string:
			p.ByName[t[0]] = ""
			for _, aka := range t[1:] {
				p.aliases[aka] = t[0]
			}
		}
	}
	return p.Parse(args)
}

// Parse predefined parameters from command arguments.
func (p *Parms) Parse(args []string) []string {
	for i := 0; i < len(args); {
		if eq := strings.Index(args[i], "="); eq > 0 {
			if p.Set(args[i][:eq], args[i][eq+1:]) == nil {
				if i < len(args)-1 {
					copy(args[i:], args[i+1:])
				}
				args = args[:len(args)-1]
			} else {
				i++
			}
		} else if i < len(args)-1 {
			k, found := p.aliases[args[i]]
			if !found {
				k = args[i]
			}
			if p.Set(k, args[i+1]) == nil {
				copy(args[i:], args[i+2:])
				args = args[:len(args)-2]
			} else {
				i++
			}
		} else {
			i++
		}
	}
	return args
}

// Set will concatenate a non empty parmeter.
func (p *Parms) Set(name, value string) error {
	cur, found := p.ByName[name]
	if !found {
		return ErrNotFound
	}
	if len(cur) > 0 && len(value) > 0 {
		p.ByName[name] = cur + " " + value
	} else {
		p.ByName[name] = value
	}
	return nil
}

// Reset all parameters.
func (p *Parms) Reset() {
	for k := range p.ByName {
		p.ByName[k] = ""
	}
}
