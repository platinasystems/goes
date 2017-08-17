// go build -o ~/y -tags "elog gtk_3_16" -gcflags "-N -l" github.com/platinasystems/go/wip/elogview

package main

import (
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/elog"

	"fmt"
	"math"
	"sort"
)

// (x,y) coordinate as complex number for easy arithmetic.
type X2 complex128

func (x X2) X() float64    { return real(x) }
func (x X2) Y() float64    { return imag(x) }
func XY(x, y float64) X2   { return X2(complex(x, y)) }
func XX(x float64) X2      { return XY(x, x) }
func (x X2) Conj() X2      { return XY(real(x), -imag(x)) }
func (x X2) Abs2() float64 { return real(x)*real(x) + imag(x)*imag(x) }
func (x X2) Abs() float64  { return math.Hypot(x.X(), x.Y()) }

type ctx cairo.Context

func (c *ctx) cr() *cairo.Context { return (*cairo.Context)(c) }
func (c *ctx) translate(x X2)     { c.cr().Translate(x.X(), x.Y()) }
func (c *ctx) scale(x X2)         { c.cr().Scale(x.X(), x.Y()) }

const (
	text_align_center = iota
	text_align_left
	text_align_right
)

type text_line struct {
	s string
	e cairo.TextExtents
}

func (c *ctx) text_box(lines []text_line) (bounding_box X2) {
	cr := c.cr()
	fe := cr.FontExtents()
	max_width := float64(0)
	for i := range lines {
		te := cr.TextExtents(lines[i].s)
		lines[i].e = te
		if te.Width > max_width {
			max_width = te.Width
		}
	}
	bounding_box = XY(max_width, fe.Height*float64(len(lines)))
	return
}

func (c *ctx) text(x X2, text_align int, lines ...text_line) X2 {
	var a float64
	switch text_align {
	case text_align_center:
		a = .5
	case text_align_left:
		a = 0
	case text_align_right:
		a = 1
	}
	cr := c.cr()
	fe := cr.FontExtents()
	bounding_box := c.text_box(lines)
	y0 := fe.Ascent - .5*fe.Height*float64(len(lines))
	for i := range lines {
		x1 := x + XY(-bounding_box.X()*a+lines[i].e.XBearing, y0+float64(i)*fe.Height)
		cr.MoveTo(x1.X(), x1.Y())
		cr.ShowText(lines[i].s)
	}
	return bounding_box
}
func (c *ctx) textf(x X2, text_align int, format string, args ...interface{}) X2 {
	var l [1]text_line
	l[0].s = fmt.Sprintf(format, args...)
	return c.text(x, text_align, l[:]...)
}

func (c *ctx) moveTo(x X2) { c.cr().MoveTo(x.X(), x.Y()) }
func (c *ctx) lineTo(x X2) { c.cr().LineTo(x.X(), x.Y()) }
func (c *ctx) rect(x0, dx X2) {
	c.cr().Rectangle(x0.X(), x0.Y(), dx.X(), dx.Y())
}

func (x *ctx) roundedRect(x0, dx X2, r float64) {
	const p2 = math.Pi / 2
	x1 := x0 + dx
	c := (*cairo.Context)(x)
	c.Arc(x0.X()+r, x0.Y()+r, r, 2*p2, 3*p2)
	c.Arc(x1.X()-r, x0.Y()+r, r, 3*p2, 4*p2)
	c.Arc(x1.X()-r, x1.Y()-r, r, 0*p2, 1*p2)
	c.Arc(x0.X()+r, x1.Y()-r, r, 1*p2, 2*p2)
	c.ClosePath()
}

type viewer struct {
	win        *gtk.Window
	screen_dpi float64
	win_dx     X2
	da         *gtk.DrawingArea
	eb         *gtk.EventBox
	eb_x       X2
	ev         *elog.View
	des        []draw_event
	ps         map[uint]*popup
	ps_slice   []*popup
	m          map[uintptr]*decoration
}

func elog_viewer(ev *elog.View, width, height int) {
	v := &viewer{ev: ev}

	gtk.Init(nil)
	v.win, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	v.win.SetTitle("Event log")
	v.win_dx = XY(float64(width), float64(height))
	v.win.SetDefaultSize(width, height)
	v.win.Connect("destroy", gtk.MainQuit)
	scr, _ := v.win.GetScreen()
	v.screen_dpi = scr.GetResolution()

	v.eb, _ = gtk.EventBoxNew()
	v.eb.SetCanFocus(true)
	v.eb.SetEvents(int(gdk.KEY_PRESS_MASK | gdk.ENTER_NOTIFY_MASK | gdk.LEAVE_NOTIFY_MASK | gdk.BUTTON_PRESS_MASK))
	v.eb.SetBorderWidth(20)
	v.eb_x = XY(20, 20)
	v.win.Add(v.eb)

	v.da, _ = gtk.DrawingAreaNew()
	v.eb.Add(v.da)

	v.da.Connect("draw", v.draw_events)
	v.eb.Connect("button_press_event", v.button_press)
	v.eb.Connect("key_press_event", v.key_press)

	v.win.ShowAll()
	gtk.Main()
}

func (v *viewer) key_press(eb *gtk.EventBox, ev *gdk.Event) {
	ke := &gdk.EventKey{ev}
	key, state := ke.KeyVal(), gdk.ModifierType(ke.State())
	t := &v.ev.Times

	subview := false
	t_min, t_max := t.Min, t.Max
	switch key {
	case KEY_q:
		gtk.MainQuit()
	case KEY_Escape:
		break
	case KEY_r:
		v.ev.Reset()
	case KEY_plus:
		dt := t.Dt * .05 // 10% smaller
		t_min += dt
		t_max -= dt
		subview = true
	case KEY_minus:
		dt := t.Dt * .05 // 10% bigger
		t_min -= dt
		t_max += dt
		subview = true
	case KEY_leftarrow, KEY_Left:
		dt := t.Unit
		if state&gdk.GDK_SHIFT_MASK != 0 {
			dt *= 10
		}
		t_min -= dt
		t_max -= dt
		subview = true
	case KEY_rightarrow, KEY_Right:
		dt := t.Unit
		if state&gdk.GDK_SHIFT_MASK != 0 {
			dt *= 10
		}
		t_min += dt
		t_max += dt
		subview = true
	default:
		return // ignore unknown key
	}
	if subview {
		v.ev.SubView(t_min, t_max)
	}
	v.delete_all_popups()
	v.eb.QueueDraw()
}

func (v *viewer) button_press(eb *gtk.EventBox, e *gdk.Event) { v.do_button_press(e, 0) }
func (p *popup) button_press(win *gtk.Window, e *gdk.Event)   { p.v.do_button_press(e, p.v.eb_x) }
func (v *viewer) do_button_press(e *gdk.Event, dx X2) {
	be := &gdk.EventButton{e}
	bv := be.ButtonVal()
	bx, by := be.MotionVal()
	if bv != 1 {
		return
	}
	mouse_x := XY(bx, by) - dx
	min_i, min_ds := uint(0), 1e10
	for i := range v.ev.Events {
		de := &v.des[i]
		if !de.visible {
			continue
		}
		if ds := de.min_distance(mouse_x); ds < min_ds {
			min_ds = ds
			min_i = uint(i)
		}
	}
	// Only accept when screen distance is less than 1/2 an inch of center.
	if min_ds < v.screen_dpi/4 {
		// Event already displayed in popup?
		if p, is_del := v.ps[min_i]; is_del {
			p.del()
		} else {
			v.new_popup(min_i)
		}
	}
}

func (v *viewer) order_popups() {
	if v.ps_slice != nil {
		v.ps_slice = v.ps_slice[:0]
	}
	ps := v.ps_slice
	for _, q := range v.ps {
		ps = append(ps, q)
	}
	sort.Slice(ps, func(i, j int) bool {
		pi, pj := ps[i], ps[j]
		return pi.ei < pj.ei
	})
	for i := range ps {
		p := ps[i]
		var prev, next *popup
		if i > 0 {
			prev = ps[i-1]
		}
		if i+1 < len(ps) {
			next = ps[i+1]
		}
		if p.prev != prev {
			p.prev = prev
			if prev != nil {
				prev.win.QueueDraw()
			}
		}
		if p.next != next {
			p.next = next
			if next != nil {
				next.win.QueueDraw()
			}
		}
	}
	v.ps_slice = ps
}

const dt_prev_invalid = 1e10

type popup struct {
	v          *viewer
	win        *gtk.Window
	da         *gtk.DrawingArea
	ei         uint
	l          []text_line
	x, dx      X2
	prev, next *popup
}

func (p *popup) get_event() *elog.Event      { return &p.v.ev.Events[p.ei] }
func (p *popup) get_draw_event() *draw_event { return &p.v.des[p.ei] }

func (v *viewer) new_popup(event_index uint) (p *popup) {
	p = &popup{v: v, ei: event_index}
	w, _ := gtk.WindowNew(gtk.WINDOW_POPUP)
	p.win = w
	w.SetResizable(false)
	w.SetDecorated(false)
	p.dx = v.win_dx
	p.x = XY(0, 0)
	w.SetSizeRequest(int(p.dx.X()), int(p.dx.Y())) // max size of drawing area
	w.SetTransientFor(v.win)
	w.SetPosition(gtk.WIN_POS_CENTER_ON_PARENT)
	scr, _ := w.GetScreen()
	vis, _ := scr.GetRGBAVisual()
	w.SetVisual(vis)

	da, _ := gtk.DrawingAreaNew()
	p.da = da
	w.Add(da)

	p.win.Connect("draw", v.draw_window_transparent_background)
	p.da.Connect("draw", p.draw_popup)
	p.win.Connect("button_press_event", p.button_press)
	if v.ps == nil {
		v.ps = make(map[uint]*popup)
	}
	v.ps[event_index] = p
	v.order_popups()
	p.win.ShowAll()
	return
}

func (p *popup) del() {
	if p.win == nil {
		return
	}
	p.win.Hide()
	p.win.Destroy()
	delete(p.v.ps, p.ei)
	p.v.order_popups()
}
func (v *viewer) delete_all_popups() {
	for _, p := range v.ps {
		p.del()
	}
}

func (p0 *popup) dt(p1 *popup) (dt float64) {
	e0, e1 := p0.get_event(), p1.get_event()
	return p1.v.ev.ElapsedTime(e1) - p0.v.ev.ElapsedTime(e0)
}

func (p *popup) draw_popup(da *gtk.DrawingArea, cr *cairo.Context) {
	e, de := p.get_event(), p.get_draw_event()

	lines := e.Strings()
	_, pc := p.v.ev.EventPath(e)
	d, _ := p.v.decorationForPc(pc)

	tb := &p.v.ev.Times
	et := p.v.ev.ElapsedTime(e)
	lines = append(lines,
		fmt.Sprintf("%.2f", et/tb.Unit))
	if p.next != nil {
		lines = append(lines, fmt.Sprintf("%.4f", p.dt(p.next)/tb.Unit))
	}

	if p.l != nil {
		p.l = p.l[:0]
	}
	for i := range lines {
		p.l = append(p.l, text_line{s: lines[i]})
	}

	c := (*ctx)(cr)
	cr.SetFontSize(10)
	bbox := c.text_box(p.l)

	center := de.rr_center + p.v.eb_x
	const radius = 4
	rr_dx := bbox + XY(2*radius, radius)
	c.roundedRect(center-rr_dx/2, rr_dx, radius)
	dc := d.color.lighten(.375)
	cr.SetSourceRGBA(dc.r, dc.g, dc.b, 1)
	cr.Fill()

	cr.SetSourceRGBA(0, 0, 0, 1)
	c.text(center-XY(bbox.X()/2, 0), text_align_left, p.l...)
	cr.Stroke()
}

type rgba struct{ r, g, b, a float64 }

func (x rgba) darken(f float64) (y rgba) {
	y.r = f * x.r
	y.g = f * x.g
	y.b = f * x.b
	return
}

func (x rgba) lighten(f float64) (y rgba) {
	y.r = 1 - f*(1-x.r)
	y.g = 1 - f*(1-x.g)
	y.b = 1 - f*(1-x.b)
	return
}

var standardColors = [...]rgba{
	rgba{r: 1},              // red
	rgba{g: 1, a: .3},       // green
	rgba{b: 1, a: .3},       // blue
	rgba{r: 1, b: 1, a: .3}, // yellow
	rgba{g: 1, b: 1, a: .3},
	rgba{r: 1, b: 1},
}

type decoration struct {
	color rgba
}

func (v *viewer) decorationForPc(pc uintptr) (d *decoration, ok bool) {
	if v.m == nil {
		v.m = make(map[uintptr]*decoration)
	}
	if d, ok = v.m[pc]; !ok {
		i := len(v.m)
		i0 := i % len(standardColors)
		d = &decoration{color: standardColors[i0]}
		if d.color.a == 0 {
			d.color.a = .2
		}
		v.m[pc] = d
	}
	return
}

type draw_event struct {
	visible bool
	// Center/size of rounded rect.
	rr_center, rr_dx X2
}

func (e *draw_event) min_distance(x X2) (ds float64) {
	// Distance to center.
	ds = (e.rr_center - x).Abs()
	// Distance to 4 corners.
	for i := 0; i < 4; i++ {
		dx2 := e.rr_dx / 2
		x, y := dx2.X(), dx2.Y()
		if i&1 != 0 {
			x = -x
		}
		if i&2 != 0 {
			y = -y
		}
		dx2 = XY(x, y)
		if l := (e.rr_center - dx2).Abs(); l < ds {
			ds = l
		}
	}
	return
}

func (v *viewer) draw_events(da *gtk.DrawingArea, cr *cairo.Context) {
	ev, tb := v.ev, &v.ev.Times

	dw := XX(40)
	w := XY(float64(da.GetAllocatedWidth()), float64(da.GetAllocatedHeight()))
	w -= dw

	c := (*ctx)(cr)
	cr.SetAntialias(cairo.ANTIALIAS_SUBPIXEL)

	// Title
	{
		cr.SetSourceRGB(0, 0, 0)
		x := XY(w.X()/2, dw.Y()/2)
		cr.SetFontSize(18)
		c.textf(x, text_align_center, "%d events, %s", len(ev.Events), tb.Start.Format("2006-01-02 15:04:05"))
		cr.Fill()
	}

	t_min, t_max, t_unit := tb.Min, tb.Max, tb.Unit

	// Left & right time axis markers.
	{
		cr.SetSourceRGBA(0, 0, 0, .25)
		cr.SetLineWidth(1)
		c.moveTo(dw)
		c.lineTo(dw + XY(0, w.Y()))
		cr.Stroke()

		tmp := XY(dw.X()+w.X(), dw.Y())
		cr.MoveTo(tmp.X(), tmp.Y())
		cr.LineTo(tmp.X(), tmp.Y()+w.Y())
		cr.Stroke()

		cr.SetSourceRGBA(0, 0, 0, 1)
		cr.SetFontSize(10)
		c.textf(dw-10i, text_align_left, "%.0f%s", t_min/t_unit, tb.UnitName)
		cr.Stroke()
		c.textf(tmp-10i, text_align_right, "%.0f%s", t_max/t_unit, tb.UnitName)
		cr.Stroke()
	}

	var x_last elib.Float64Vec

	if v.des == nil {
		v.des = make([]draw_event, len(ev.Events))
	}

	// Draw events.
	for i := range ev.Events {
		e := &ev.Events[i]
		de := &v.des[i]
		t := ev.ElapsedTime(e)

		x := XY((t-t_min)/(t_max-t_min)*w.X(), w.Y()/2)
		if de.visible = x.X() >= 0 && x.X() < w.X(); !de.visible {
			continue
		}

		lines := e.Strings()
		_, pc := ev.EventPath(e)
		d, _ := v.decorationForPc(pc)

		center := dw + x
		radius := 4.
		cr.SetSourceRGBA(d.color.r, d.color.g, d.color.b, d.color.a)
		cr.SetFontSize(9)
		fe, te := cr.FontExtents(), cr.TextExtents(lines[0])

		// Choose integer Y such that text will not overlap with other text.
		iy := uint(0)
		const iy_max = 64
		for iy < iy_max {
			x_last.Validate(uint(iy))
			if x_left := center.X() - te.Width/2; x_last[iy]+2*radius < x_left {
				x_last[iy] = x_left + te.Width
				break
			}
			iy++
		}
		if iy > iy_max {
			de.visible = false
			continue
		}

		// Avoid level 0 which will be used to plot event points.  See (A) below.
		iy++

		// Odd goes above baseline; even goes below baseline.
		var idy int
		if iy%2 != 0 {
			idy = int((iy + 1) / 2)
		} else {
			idy = -int(iy / 2)
		}

		// Rounded rect with event text inside.
		rr_center := center - XY(0, (fe.Height+2*radius)*float64(idy))

		if de.visible = rr_center.Y() >= dw.Y() && rr_center.Y()+dw.Y() < w.Y(); !de.visible {
			continue
		}

		cr.SetSourceRGB(0, 0, 0)
		tw := c.textf(rr_center, text_align_center, lines[0])
		rr_dx := tw + XY(2*radius, radius)
		cr.Stroke()
		cr.SetSourceRGBA(d.color.r, d.color.g, d.color.b, d.color.a)
		c.roundedRect(rr_center-rr_dx/2, rr_dx, radius)
		cr.Fill()

		// Save away
		de.rr_center = rr_center
		de.rr_dx = rr_dx

		// (A) Plot event points on event line at iy == 0 and lines between text & event line.
		max_idy := 4
		if idy >= -max_idy && idy <= +max_idy {
			dot_radius := .9 * radius
			if idy > 0 {
				c.moveTo(rr_center + XY(0, (rr_dx/2).Y()))
				c.lineTo(center - XY(0, dot_radius))
			} else {
				c.moveTo(rr_center - XY(0, (rr_dx/2).Y()))
				c.lineTo(center + XY(0, dot_radius))
			}
			cr.SetOperator(cairo.OPERATOR_ATOP)
			cr.SetLineWidth(2)
			cr.Stroke()

			cr.SetSourceRGBA(d.color.r, d.color.g, d.color.b, .75)
			cr.MoveTo(center.X(), center.Y())
			cr.Arc(center.X(), center.Y(), dot_radius, 0, 2*math.Pi)
			cr.Fill()
		}
	}
}

func (v *viewer) draw_window_transparent_background(win *gtk.Window, cr *cairo.Context) {
	w := XY(float64(win.GetAllocatedWidth()), float64(win.GetAllocatedHeight()))
	c := (*ctx)(cr)
	cr.SetOperator(cairo.OPERATOR_SOURCE)
	c.rect(0, w)
	cr.Fill()
}
