// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vnet

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/elib/parse"

	"fmt"
	"sort"
	"time"
)

func (hi *Hi) ParseWithArgs(in *parse.Input, args *parse.Args) {
	v := args.Get().(*Vnet)
	if !in.Parse("%v", v.hwIfIndexByName, hi) {
		in.ParseError()
	}
}

func (si *Si) ParseWithArgs(in *parse.Input, args *parse.Args) {
	v := args.Get().(*Vnet)
	var hi Hi
	if !in.Parse("%v", v.hwIfIndexByName, &hi) {
		in.ParseError()
	}
	// Initially get software interface from hardware interface.
	h := v.HwIfer(hi)
	hw := h.GetHwIf()
	*si = hw.si
	var (
		id IfId
		ok bool
	)
	if h.ParseId(&id, in) {
		if *si, ok = hw.subSiById[id]; !ok {
			panic(fmt.Errorf("unkown sub interface id: %d", id))
		}
	}
}

type ifChooser struct {
	v         *Vnet
	isHw      bool
	finalized bool
	re        parse.Regexp
	siMap     map[Si]struct{}
	hiMap     map[Hi]struct{}
	sw        swIfIndices
	hw        hwIfIndices
}

func (c *ifChooser) parse(in *parse.Input) {
	var empty struct{}
	var (
		si Si
		hi Hi
	)
	switch {
	case !c.isHw && in.Parse("%v", &si, c.v):
		c.siMap[si] = empty
	case c.isHw && in.Parse("%v", &hi, c.v):
		c.hiMap[hi] = empty
	case in.Parse("m%*atching %v", &c.re):
	default:
		in.ParseError()
	}
}

func (c *ifChooser) finalize() {
	if c.finalized {
		return
	}
	c.finalized = true
	var empty struct{}
	if c.isHw {
		if len(c.hiMap) == 0 || c.re.Valid() {
			c.v.hwIferPool.Foreach(func(r HwInterfacer) {
				h := r.GetHwIf()
				if h.unprovisioned {
					return
				}
				if c.re.Valid() && !c.re.MatchString(h.name) {
					return
				}
				c.hiMap[h.hi] = empty
			})
		}
		c.hw.Vnet = c.v
		for hi := range c.hiMap {
			c.hw.ifs = append(c.hw.ifs, hi)
		}
		sort.Sort(&c.hw)
	} else {
		if len(c.siMap) == 0 || c.re.Valid() {
			c.v.swInterfaces.ForeachIndex(func(i uint) {
				si := Si(i)
				if c.re.Valid() && !c.re.MatchString(si.Name(c.v)) {
					return
				}
				c.siMap[si] = empty
			})
		}
		c.sw.Vnet = c.v
		for si := range c.siMap {
			c.sw.ifs = append(c.sw.ifs, si)
		}
		sort.Sort(&c.sw)
	}
}

type HwIfChooser ifChooser
type SwIfChooser ifChooser

func (c *HwIfChooser) Init(v *Vnet) {
	c.v = v
	c.isHw = true
	c.hiMap = make(map[Hi]struct{})
}
func (c *SwIfChooser) Init(v *Vnet) {
	c.v = v
	c.isHw = false
	c.siMap = make(map[Si]struct{})
}

func (c *HwIfChooser) Parse(in *parse.Input) { (*ifChooser)(c).parse(in) }
func (c *SwIfChooser) Parse(in *parse.Input) { (*ifChooser)(c).parse(in) }

func (c *HwIfChooser) Foreach(f func(v *Vnet, h HwInterfacer)) {
	(*ifChooser)(c).finalize()
	for _, hi := range c.hw.ifs {
		f(c.v, c.v.HwIfer(hi))
	}
}
func (c *SwIfChooser) Foreach(f func(v *Vnet, si Si)) {
	(*ifChooser)(c).finalize()
	for _, si := range c.sw.ifs {
		f(c.v, si)
	}
}

type showIfConfig struct {
	detail  bool
	summary bool
	re      parse.Regexp
	colMap  map[string]bool
	siMap   map[Si]bool
	hiMap   map[Hi]bool
}

func (c *showIfConfig) parse(v *Vnet, in *cli.Input, isHw bool) {
	c.detail = false
	c.colMap = map[string]bool{
		"Rate":   false,
		"Driver": false,
	}
	if isHw {
		c.hiMap = make(map[Hi]bool)
	} else {
		c.siMap = make(map[Si]bool)
	}
	for !in.End() {
		var (
			si Si
			hi Hi
		)
		switch {
		case !isHw && in.Parse("%v", &si, v):
			c.siMap[si] = true
		case isHw && in.Parse("%v", &hi, v):
			c.hiMap[hi] = true
		case in.Parse("m%*atching %v", &c.re):
		case in.Parse("d%*etail"):
			c.detail = true
			c.colMap["Driver"] = true
		case in.Parse("s%*ummary"):
			c.summary = true
		case in.Parse("r%*ate"):
			c.colMap["Rate"] = true
		default:
			in.ParseError()
		}
	}
	if c.summary {
		c.colMap["Counter"] = false
		c.colMap["Count"] = false
		if isHw {
			c.colMap["Driver"] = true
		}
	}
}

type swIfIndices struct {
	*Vnet
	ifs []Si
}

func (h *swIfIndices) Less(i, j int) bool { return h.SwLessThan(h.SwIf(h.ifs[i]), h.SwIf(h.ifs[j])) }
func (h *swIfIndices) Swap(i, j int)      { h.ifs[i], h.ifs[j] = h.ifs[j], h.ifs[i] }
func (h *swIfIndices) Len() int           { return len(h.ifs) }

type showSwIf struct {
	Name    string `format:"%-30s" align:"left"`
	State   string `format:"%-12s" align:"left"`
	Counter string `format:"%-30s" align:"left"`
	Count   string `format:"%16s" align:"right"`
	Rate    string `format:"%16s" align:"right"`
}
type showSwIfs []showSwIf

func (v *Vnet) showSwIfs(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {

	cf := &showIfConfig{}
	cf.parse(v, in, false)

	swIfs := &swIfIndices{Vnet: v}
	if len(cf.siMap) == 0 {
		v.swInterfaces.ForeachIndex(func(i uint) {
			si := Si(i)
			// Skip unprovisioned interfaces.
			sw := v.SwIf(si)
			if hw := v.SupHwIf(sw); hw != nil && hw.unprovisioned {
				return
			}
			// Skip interfaces which don't match regexps.
			if cf.re.Valid() && !cf.re.MatchString(si.Name(v)) {
				return
			}
			swIfs.ifs = append(swIfs.ifs, si)
		})
	} else {
		for si, _ := range cf.siMap {
			swIfs.ifs = append(swIfs.ifs, si)
		}
	}

	if cf.re.Valid() && len(swIfs.ifs) == 0 {
		fmt.Fprintf(w, "No interfaces match expression: `%s'\n", cf.re)
		return
	}

	sort.Sort(swIfs)

	v.syncSwIfCounters()

	sifs := showSwIfs{}
	dt := time.Since(v.timeLastClear).Seconds()
	alwaysReport := len(cf.siMap) > 0 || cf.re.Valid()
	for i := range swIfs.ifs {
		si := swIfs.ifs[i]
		sw := v.SwIf(si)
		first := true
		firstIf := showSwIf{
			Name:  si.Name(v),
			State: sw.flags.String(),
		}
		if cf.summary {
			sifs = append(sifs, firstIf)
			continue
		}
		v.foreachSwIfCounter(cf.detail, si, func(counter string, count uint64) {
			s := showSwIf{
				Counter: counter,
				Count:   fmt.Sprintf("%d", count),
				Rate:    fmt.Sprintf("%.2e", float64(count)/dt),
			}
			if first {
				first = false
				s.Name = firstIf.Name
				s.State = firstIf.State
			}
			sifs = append(sifs, s)
		})
		// Always at least report name and state for specified interfaces.
		if first && alwaysReport {
			sifs = append(sifs, firstIf)
		}
	}
	if len(sifs) > 0 {
		elib.Tabulate(sifs).WriteCols(w, cf.colMap)
	} else {
		fmt.Fprintln(w, "All counters are zero")
	}
	return
}

func (v *Vnet) clearSwIfs(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	v.clearIfCounters()
	return
}

type hwIfIndices struct {
	*Vnet
	ifs []Hi
}

func (h *hwIfIndices) Less(i, j int) bool { return h.HwLessThan(h.HwIf(h.ifs[i]), h.HwIf(h.ifs[j])) }
func (h *hwIfIndices) Swap(i, j int)      { h.ifs[i], h.ifs[j] = h.ifs[j], h.ifs[i] }
func (h *hwIfIndices) Len() int           { return len(h.ifs) }

type showHwIf struct {
	Name    string `format:"%-30s"`
	Driver  string `width:"20" format:"%-12s" align:"center"`
	Address string `format:"%-12s" align:"center"`
	Link    string `width:"12"`
	Counter string `format:"%-30s" align:"left"`
	Count   string `format:"%16s" align:"right"`
	Rate    string `format:"%16s" align:"right"`
}
type showHwIfs []showHwIf

func (ns showHwIfs) Less(i, j int) bool { return ns[i].Name < ns[j].Name }
func (ns showHwIfs) Swap(i, j int)      { ns[i], ns[j] = ns[j], ns[i] }
func (ns showHwIfs) Len() int           { return len(ns) }

func (v *Vnet) showHwIfs(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	cf := showIfConfig{}
	cf.parse(v, in, true)

	hwIfs := &hwIfIndices{Vnet: v}

	if len(cf.hiMap) == 0 {
		v.hwIferPool.Foreach(func(r HwInterfacer) {
			h := r.GetHwIf()
			if h.unprovisioned {
				return
			}
			if cf.re.Valid() && !cf.re.MatchString(h.name) {
				return
			}
			hwIfs.ifs = append(hwIfs.ifs, h.hi)
		})
	} else {
		for hi, _ := range cf.hiMap {
			hwIfs.ifs = append(hwIfs.ifs, hi)
		}
	}

	if cf.re.Valid() && len(hwIfs.ifs) == 0 {
		fmt.Fprintf(w, "No interfaces match expression: `%s'\n", cf.re)
		return
	}

	sort.Sort(hwIfs)

	ifs := showHwIfs{}
	dt := time.Since(v.timeLastClear).Seconds()
	alwaysReport := len(cf.siMap) > 0 || cf.re.Valid()
	for i := range hwIfs.ifs {
		hi := v.HwIfer(hwIfs.ifs[i])
		h := hi.GetHwIf()
		first := true
		firstIf := showHwIf{
			Name:    h.name,
			Driver:  hi.DriverName(),
			Address: hi.FormatAddress(),
			Link:    h.LinkString(),
		}
		if cf.summary {
			ifs = append(ifs, firstIf)
			continue
		}
		v.foreachHwIfCounter(cf.detail, h.hi, func(counter string, count uint64) {
			s := showHwIf{
				Counter: counter,
				Count:   fmt.Sprintf("%d", count),
				Rate:    fmt.Sprintf("%.2e", float64(count)/dt),
			}
			if first {
				first = false
				s.Name = firstIf.Name
				s.Driver = firstIf.Driver
				s.Address = firstIf.Address
				s.Link = firstIf.Link
			}
			ifs = append(ifs, s)
		})
		// Always at least report name and state for specified interfaces.
		if first && alwaysReport {
			ifs = append(ifs, firstIf)
		}
	}
	if len(ifs) > 0 {
		elib.Tabulate(ifs).WriteCols(w, cf.colMap)
	} else {
		fmt.Fprintln(w, "All counters are zero")
	}
	return
}

func (v *Vnet) setSwIf(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	var (
		isUp parse.UpDown
		si   Si
	)
	switch {
	case in.Parse("state %v %v", &si, v, &isUp):
		s := v.SwIf(si)
		err = s.SetAdminUp(v, bool(isUp))
	default:
		err = cli.ParseError
	}
	return
}

func (v *Vnet) setHwIf(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	var hi Hi

	var (
		mtu      uint
		bw       Bandwidth
		enable   parse.Enable
		up       parse.UpDown
		loopback IfLoopbackType
	)

	if !in.Parse("%v", &hi, v) {
		err = fmt.Errorf("no such hardware interface")
		return
	}

	h := v.HwIfer(hi)
	hwif := v.HwIf(hi)

	switch {
	case in.Parse("lo%*opback %v", &loopback):
		err = h.SetLoopback(loopback)
	case in.Parse("li%*nk %v", &up):
		if elib.Debug {
			err = hwif.SetLinkUp(bool(up))
		} else {
			err = fmt.Errorf("not supported")
		}
	case in.Parse("mtu %d", &mtu):
		err = hwif.SetMaxPacketSize(mtu)
	case in.Parse("p%*rovision %v", &enable):
		err = hwif.SetProvisioned(bool(enable))
	case in.Parse("s%*peed %v", &bw):
		err = hwif.SetSpeed(bw)
	default:
		var ok bool
		if ok, err = h.ConfigureHwIf(in); !ok {
			err = cli.ParseError
		}
	}
	return
}

func init() {
	AddInit(func(v *Vnet) {
		cmds := [...]cli.Command{
			cli.Command{
				Name:      "show interfaces",
				ShortHelp: "show interface statistics",
				Action:    v.showSwIfs,
			},
			cli.Command{
				Name:      "clear interfaces",
				ShortHelp: "clear interface statistics",
				Action:    v.clearSwIfs,
			},
			cli.Command{
				Name:      "show hardware-interfaces",
				ShortHelp: "show hardware interface statistics",
				Action:    v.showHwIfs,
			},
			cli.Command{
				Name:      "set interface",
				ShortHelp: "set interface commands",
				Action:    v.setSwIf,
			},
			cli.Command{
				Name:      "set hardware-interface",
				ShortHelp: "set hardware interface commands",
				Action:    v.setHwIf,
			},
			cli.Command{
				Name:      "show buffers",
				ShortHelp: "show dma buffer usage",
				Action:    v.showBufferUsage,
			},
		}
		for i := range cmds {
			v.CliAdd(&cmds[i])
		}
	})
}
