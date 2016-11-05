// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vnet

import (
	"github.com/platinasystems/go/elib/loop"
	"github.com/platinasystems/go/elib/parse"

	"errors"
	"fmt"
	"time"
)

type HwIf struct {
	vnet *Vnet

	name string

	hi Hi
	si Si

	// Hardware link state: up or down
	linkUp bool

	// Hardware is unprovisioned.
	// Interfaces with 4 SERDES lanes will be represented as 4 interfaces.
	// Lanes may all be a single interface (1 provisioned 4 lane interface +
	// 3 unprovisioned 0 lane interfaces).
	unprovisioned bool

	speed Bandwidth

	// Mask of SERDES lanes for this interface.
	laneMask LaneMask

	// Max size of packet in bytes (MTU)
	maxPacketSize uint

	defaultId IfIndex
	subSiById map[IfIndex]Si
}

//go:generate gentemplate -d Package=vnet -id HwIf -d PoolType=hwIferPool -d Type=HwInterfacer -d Data=elts github.com/platinasystems/go/elib/pool.tmpl

type IfIndex uint32
type LaneMask uint32

type HwInterfacer interface {
	Devicer
	HwIfClasser
	GetHwIf() *HwIf
}

func (h *HwIf) GetHwIf() *HwIf { return h }
func (h *HwIf) Name() string   { return h.name }
func (h *HwIf) Si() Si         { return h.si }
func (h *HwIf) Hi() Hi         { return h.hi }
func (h *HwIf) IsUnix() bool   { return false }

func (h *HwIf) SetName(v *Vnet, name string) {
	h.name = name
	v.hwIfIndexByName.Set(name, uint(h.hi))
}
func (v *Vnet) HwIfByName(name string) (Hi, bool) {
	hi, ok := v.hwIfIndexByName[name]
	return Hi(hi), ok
}

func (h *HwIf) SetSubInterface(id IfIndex, si Si) {
	if h.subSiById == nil {
		h.subSiById = make(map[IfIndex]Si)
	}
	h.subSiById[id] = si
}

func (h *HwIf) LinkString() (s string) {
	s = "down"
	if h.linkUp {
		s = "up"
	}
	return
}

// Software and hardware interface index.
// Alias for commonly used types.
type Si IfIndex
type Hi IfIndex

const (
	SiNil Si = ^Si(0)
	HiNil Hi = ^Hi(0)
)

type swIfType uint16

const (
	swIfTypeHardware swIfType = iota + 1
	swIfTypeSubInterface
)

type swIfFlag uint16

const (
	swIfAdminUpIndex, swIfAdminUp swIfFlag = iota, 1 << iota
	swIfPuntIndex, swIfPunt
)

func (f swIfFlag) String() (s string) {
	s = "down"
	if f&swIfAdminUp != 0 {
		s = "up"
	}
	extra := ""
	if f&swIfPunt != 0 {
		if extra != "" {
			extra += ", "
		}
		extra += "punt"
	}
	if extra != "" {
		s += "(" + extra + ")"
	}
	return
}

type swIf struct {
	typ   swIfType
	flags swIfFlag

	// Pool index for this interface.
	si Si

	// Software interface index of super-interface.
	// Equal to index if this interface is not a sub-interface.
	supSi Si

	// For hardware interface: HwIfIndex
	// For sub interface: sub interface id (e.g. vlan/vc number).
	id IfIndex
}

//go:generate gentemplate -d Package=vnet -id swIf -d PoolType=swIfPool -d Type=swIf -d Data=elts github.com/platinasystems/go/elib/pool.tmpl

func (m *Vnet) NewSwIf(typ swIfType, id IfIndex) (si Si) {
	si = Si(m.swInterfaces.GetIndex())
	s := m.SwIf(si)
	s.typ = typ
	s.si = si
	s.supSi = si
	s.id = id
	m.counterValidateSw(si)

	isDel := false
	for i := range m.swIfAddDelHooks.hooks {
		err := m.swIfAddDelHooks.Get(i)(m, s.si, isDel)
		if err != nil {
			panic(err) // how to recover?
		}
	}
	return
}

func (m *interfaceMain) SwIf(i Si) *swIf { return &m.swInterfaces.elts[i] }
func (m *interfaceMain) SupSi(i Si) Si   { return m.SwIf(i).supSi }
func (m *interfaceMain) SupSwIf(s *swIf) (sup *swIf) {
	sup = s
	if s.supSi != s.si {
		sup = m.SwIf(s.supSi)
	}
	return
}
func (m *interfaceMain) HwIfer(i Hi) HwInterfacer { return m.hwIferPool.elts[i] }
func (m *interfaceMain) HwIf(i Hi) *HwIf          { return m.HwIfer(i).GetHwIf() }
func (m *interfaceMain) SupHwIf(s *swIf) *HwIf {
	sup := m.SupSwIf(s)
	return m.HwIf(Hi(sup.id))
}
func (m *interfaceMain) SupHi(si Si) Hi {
	sw := m.SwIf(si)
	hw := m.SupHwIf(sw)
	return hw.hi
}

func (m *interfaceMain) HwIferForSi(i Si) (h HwInterfacer, ok bool) {
	sw := m.SwIf(i)
	if ok = sw.typ == swIfTypeHardware; ok {
		h = m.HwIfer(Hi(sw.id))
	}
	return
}

func (s *swIf) IfName(vn *Vnet) (v string) {
	v = vn.SupHwIf(s).name
	if s.typ != swIfTypeHardware {
		v += fmt.Sprintf(".%d", s.id)
	}
	return
}
func (i Si) Name(v *Vnet) string { return v.SwIf(i).IfName(v) }
func (i Hi) Name(v *Vnet) string { return v.HwIf(i).name }

func (i *swIf) Id(v *Vnet) (id IfIndex) {
	id = i.id
	if i.typ == swIfTypeHardware {
		h := v.HwIf(Hi(id))
		id = h.defaultId
	}
	return
}

func (i *swIf) IsAdminUp() bool      { return i.flags&swIfAdminUp != 0 }
func (si Si) IsAdminUp(v *Vnet) bool { return v.SwIf(si).IsAdminUp() }

func (sw *swIf) SetAdminUp(v *Vnet, wantUp bool) (err error) {
	isUp := sw.flags&swIfAdminUp != 0
	if isUp == wantUp {
		return
	}
	sw.flags ^= swIfAdminUp
	for i := range v.swIfAdminUpDownHooks.hooks {
		err = v.swIfAdminUpDownHooks.Get(i)(v, sw.si, wantUp)
		if err != nil {
			return
		}
	}
	return
}

func (si Si) SetAdminUp(v *Vnet, isUp bool) (err error) {
	s := v.SwIf(si)
	return s.SetAdminUp(v, isUp)
}

func (h *HwIf) SetAdminUp(isUp bool) (err error) {
	if h.unprovisioned {
		err = errors.New("hardware interface is unprovisioned")
		return
	}

	s := h.vnet.SwIf(h.si)
	err = s.SetAdminUp(h.vnet, isUp)
	return
}

func (hi Hi) SetAdminUp(v *Vnet, isUp bool) (err error) {
	h := v.HwIf(hi)
	return h.SetAdminUp(isUp)
}

func (h *HwIf) IsProvisioned() bool { return !h.unprovisioned }

func (h *HwIf) SetProvisioned(v bool) (err error) {
	if !h.unprovisioned == v {
		return
	}
	vn := h.vnet
	for i := range vn.hwIfProvisionHooks.hooks {
		err = vn.hwIfProvisionHooks.Get(i)(vn, h.hi, v)
		if err != nil {
			break
		}
	}
	// Toggle provisioning hooks show no error.
	if err == nil {
		h.unprovisioned = !v
	}
	return
}

func (h *HwIf) IsLinkUp() bool      { return h.linkUp }
func (hi Hi) IsLinkUp(v *Vnet) bool { return v.HwIf(hi).IsLinkUp() }

func (h *HwIf) SetLinkUp(v bool) (err error) {
	if h.linkUp == v {
		return
	}
	h.linkUp = v
	vn := h.vnet
	for i := range vn.hwIfLinkUpDownHooks.hooks {
		err = vn.hwIfLinkUpDownHooks.Get(i)(vn, h.hi, v)
		if err != nil {
			return
		}
	}
	return
}

type LinkStateEvent struct {
	Event
	Hi   Hi
	IsUp bool
}

func (e *LinkStateEvent) EventAction() {
	h := e.Vnet().HwIf(e.Hi)
	if err := h.SetLinkUp(e.IsUp); err != nil {
		panic(err)
	}
}

func (e *LinkStateEvent) String() string {
	return fmt.Sprintf("link-state %s %v", e.Hi.Name(e.Vnet()), e.IsUp)
}

func (h *HwIf) MaxPacketSize() uint { return h.maxPacketSize }

func (h *HwIf) SetMaxPacketSize(v uint) (err error) {
	h.maxPacketSize = v
	// fixme call hooks
	return
}

func (h *HwIf) Speed() Bandwidth   { return h.speed }
func (h *HwIf) LaneMask() LaneMask { return h.laneMask }

func (hw *HwIf) SetSpeed(v Bandwidth) (err error) {
	vn := hw.vnet
	h := vn.HwIfer(hw.hi)
	err = h.ValidateSpeed(v)
	if err == nil {
		hw.speed = v
	}
	return
}
func (hi Hi) SetSpeed(v *Vnet, s Bandwidth) error { return v.HwIf(hi).SetSpeed(s) }

var ErrNotSupported = errors.New("not supported")

// Default versions.
func (h *HwIf) ValidateSpeed(v Bandwidth) (err error) { return }
func (h *HwIf) SetLoopback(v IfLoopbackType) (err error) {
	switch v {
	case IfLoopbackNone:
	default:
		err = ErrNotSupported
	}
	return
}
func (h *HwIf) GetSwInterfaceCounterNames() (nm InterfaceCounterNames) { return }
func (h *HwIf) DefaultId() IfIndex                                     { return 0 }
func (a *HwIf) LessThan(b HwInterfacer) bool                           { return a.hi < b.GetHwIf().hi }

type interfaceMain struct {
	hwIferPool      hwIferPool
	hwIfIndexByName parse.StringMap
	swInterfaces    swIfPool
	ifThreads       ifThreadVec

	// Counters
	swIfCounterNames     InterfaceCounterNames
	swIfCounterSyncHooks SwIfCounterSyncHookVec

	swIfAddDelHooks      SwIfAddDelHookVec
	swIfAdminUpDownHooks SwIfAdminUpDownHookVec
	hwIfAddDelHooks      HwIfAddDelHookVec
	hwIfLinkUpDownHooks  HwIfLinkUpDownHookVec
	hwIfProvisionHooks   HwIfProvisionHookVec

	timeLastClear time.Time
}

func (m *interfaceMain) init() {
	// Give clear counters time an initial value.
	m.timeLastClear = time.Now()
}

//go:generate gentemplate -d Package=vnet -id ifThread -d VecType=ifThreadVec -d Type=*InterfaceThread github.com/platinasystems/go/elib/vec.tmpl

func (v *Vnet) RegisterAndProvisionHwInterface(h HwInterfacer, provision bool, format string, args ...interface{}) (err error) {
	hi := Hi(v.hwIferPool.GetIndex())
	v.hwIferPool.elts[hi] = h
	hw := h.GetHwIf()
	hw.hi = hi
	hw.SetName(v, fmt.Sprintf(format, args...))
	hw.vnet = v
	hw.defaultId = h.DefaultId()
	hw.unprovisioned = !provision
	hw.si = v.NewSwIf(swIfTypeHardware, IfIndex(hw.hi))

	isDel := false
	m := &v.interfaceMain
	for i := range m.hwIfAddDelHooks.hooks {
		err := m.hwIfAddDelHooks.Get(i)(v, hi, isDel)
		if err != nil {
			panic(err) // how to recover?
		}
	}
	return
}

func (v *Vnet) RegisterHwInterface(h HwInterfacer, format string, args ...interface{}) (err error) {
	return v.RegisterAndProvisionHwInterface(h, true, format, args...)
}

func (m *interfaceMain) newInterfaceThread() (t *InterfaceThread) {
	t = &InterfaceThread{}
	m.counterInit(t)
	return
}

func (m *interfaceMain) GetIfThread(id uint) (t *InterfaceThread) {
	m.ifThreads.Validate(id)
	if t = m.ifThreads[id]; t == nil {
		t = m.newInterfaceThread()
		m.ifThreads[id] = t
	}
	return
}
func (n *Node) GetIfThread() *InterfaceThread { return n.Vnet.GetIfThread(n.ThreadId()) }

func (v *Vnet) ForeachHwIf(unixOnly bool, f func(hi Hi)) {
	for i := range v.hwIferPool.elts {
		if v.hwIferPool.IsFree(uint(i)) {
			continue
		}
		hwifer := v.hwIferPool.elts[i]
		if unixOnly && !hwifer.IsUnix() {
			continue
		}
		h := hwifer.GetHwIf()
		if h.unprovisioned {
			continue
		}
		f(h.hi)
	}
}

// Interface ordering for output.
func (v *Vnet) HwLessThan(a, b *HwIf) bool {
	ha, hb := v.HwIfer(a.hi), v.HwIfer(b.hi)
	da, db := ha.DriverName(), hb.DriverName()
	if da != db {
		return da < db
	}
	return ha.LessThan(hb)
}

func (v *Vnet) SwLessThan(a, b *swIf) bool {
	ha, hb := v.SupHwIf(a), v.SupHwIf(b)
	if ha != hb {
		return v.HwLessThan(ha, hb)
	}
	return a.id < b.id
}

// Interface can loopback at MAC or PHY.
type IfLoopbackType int

const (
	IfLoopbackNone IfLoopbackType = iota
	IfLoopbackMac
	IfLoopbackPhy
)

func (x *IfLoopbackType) Parse(in *parse.Input) {
	switch text := in.Token(); text {
	case "none":
		*x = IfLoopbackNone
	case "mac":
		*x = IfLoopbackMac
	case "phy":
		*x = IfLoopbackPhy
	default:
		panic(parse.ErrInput)
	}
	return
}

// To clarify units: 1e9 * vnet.Bps
const (
	Bps    = 1e0
	Kbps   = 1e3
	Mbps   = 1e6
	Gbps   = 1e9
	Tbps   = 1e12
	Bytes  = 1
	Kbytes = 1 << 10
	Mbytes = 1 << 20
	Gbytes = 1 << 30
)

type Bandwidth float64

func (b Bandwidth) String() string {
	if b == 0 {
		return "autoneg"
	}
	unit := Bandwidth(1)
	prefix := ""
	switch {
	case b < Kbps:
		break
	case b <= Mbps:
		unit = Kbps
		prefix = "k"
	case b <= Gbps:
		unit = Mbps
		prefix = "m"
	case b <= Tbps:
		unit = Gbps
		prefix = "g"
	default:
		unit = Tbps
		prefix = "t"
	}
	b /= unit
	return fmt.Sprintf("%g%s", b, prefix)
}

func (b *Bandwidth) Parse(in *parse.Input) {
	var f float64

	// Special speed code "autoneg" means auto-negotiate speed.
	if in.Parse("au%*toneg") {
		*b = 0
		return
	}

	in.Parse("%f", &f)
	unit := Bps
	switch {
	case in.AtOneof("Kk") < 2:
		unit = Kbps
	case in.AtOneof("Mm") < 2:
		unit = Mbps
	case in.AtOneof("Gg") < 2:
		unit = Gbps
	case in.AtOneof("Tt") < 2:
		unit = Tbps
	}
	*b = Bandwidth(float64(f) * unit)
}

// Class of hardware interfaces, for example, ethernet, sonet, srp, docsis, etc.
type HwIfClasser interface {
	DefaultId() IfIndex
	FormatAddress() string
	SetRewrite(v *Vnet, r *Rewrite, t PacketType, dstAddr []byte)
	FormatRewrite(r *Rewrite) string
	ParseRewrite(r *Rewrite, in *parse.Input)
}

type Devicer interface {
	Noder
	loop.OutputLooper
	DriverName() string // name of device driver
	LessThan(b HwInterfacer) bool
	IsUnix() bool
	ValidateSpeed(speed Bandwidth) error
	SetLoopback(v IfLoopbackType) error
	GetHwInterfaceCounterNames() InterfaceCounterNames
	GetSwInterfaceCounterNames() InterfaceCounterNames
	GetHwInterfaceCounterValues(t *InterfaceThread)
}

type SwIfAddDelHook func(v *Vnet, si Si, isDel bool) error
type SwIfAdminUpDownHook func(v *Vnet, si Si, isUp bool) error
type SwIfCounterSyncHook func(v *Vnet)
type HwIfAddDelHook func(v *Vnet, hi Hi, isDel bool) error
type HwIfLinkUpDownHook func(v *Vnet, hi Hi, isUp bool) error
type HwIfProvisionHook func(v *Vnet, hi Hi, isProvisioned bool) error

//go:generate gentemplate -id SwIfAddDelHook -d Package=vnet -d DepsType=SwIfAddDelHookVec -d Type=SwIfAddDelHook -d Data=hooks github.com/platinasystems/go/elib/dep/dep.tmpl
//go:generate gentemplate -id SwIfAdminUpDownHook -d Package=vnet -d DepsType=SwIfAdminUpDownHookVec -d Type=SwIfAdminUpDownHook -d Data=hooks github.com/platinasystems/go/elib/dep/dep.tmpl
//go:generate gentemplate -id HwIfAddDelHook -d Package=vnet -d DepsType=HwIfAddDelHookVec -d Type=HwIfAddDelHook -d Data=hooks github.com/platinasystems/go/elib/dep/dep.tmpl
//go:generate gentemplate -id HwIfLinkUpDownHook -d Package=vnet -d DepsType=HwIfLinkUpDownHookVec -d Type=HwIfLinkUpDownHook -d Data=hooks github.com/platinasystems/go/elib/dep/dep.tmpl
//go:generate gentemplate -id HwIfProvisionHook -d Package=vnet -d DepsType=HwIfProvisionHookVec -d Type=HwIfProvisionHook -d Data=hooks github.com/platinasystems/go/elib/dep/dep.tmpl
//go:generate gentemplate -id SwIfCounterSyncHookVec -d Package=vnet -d DepsType=SwIfCounterSyncHookVec -d Type=SwIfCounterSyncHook -d Data=hooks github.com/platinasystems/go/elib/dep/dep.tmpl

func (m *interfaceMain) RegisterSwIfAddDelHook(h SwIfAddDelHook) {
	m.swIfAddDelHooks.Add(h)
}
func (m *interfaceMain) RegisterSwIfAdminUpDownHook(h SwIfAdminUpDownHook) {
	m.swIfAdminUpDownHooks.Add(h)
}
func (m *interfaceMain) RegisterSwIfCounterSyncHook(h SwIfCounterSyncHook) {
	m.swIfCounterSyncHooks.Add(h)
}
func (m *interfaceMain) RegisterHwIfAddDelHook(h HwIfAddDelHook) {
	m.hwIfAddDelHooks.Add(h)
}
func (m *interfaceMain) RegisterHwIfLinkUpDownHook(h HwIfLinkUpDownHook) {
	m.hwIfLinkUpDownHooks.Add(h)
}
func (m *interfaceMain) RegisterHwIfProvisionHook(h HwIfProvisionHook) {
	m.hwIfProvisionHooks.Add(h)
}
