package elogview

import (
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/elib/math/r2"

	"fmt"
	"math"
	"sort"
	"sync"
)

type ctx cairo.Context

func (c *ctx) cr() *cairo.Context { return (*cairo.Context)(c) }
func (c *ctx) translate(x r2.V)   { c.cr().Translate(x.X(), x.Y()) }
func (c *ctx) scale(x r2.V)       { c.cr().Scale(x.X(), x.Y()) }

type rgba struct{ r, g, b, a float64 }

func (x rgba) RGBA() (r, g, b, a float64) { r, g, b, a = x.r, x.g, x.b, x.a; return }
func gray(x, a float64) rgba              { return rgba{r: x, g: x, b: x, a: a} }

func (x rgba) darken(f float64) (y rgba) {
	y.r = f * x.r
	y.g = f * x.g
	y.b = f * x.b
	y.a = x.a
	return
}

func (x rgba) lighten(f float64) (y rgba) {
	y.r = 1 - f*(1-x.r)
	y.g = 1 - f*(1-x.g)
	y.b = 1 - f*(1-x.b)
	y.a = x.a
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

const (
	text_align_center = iota
	text_align_left
	text_align_right
)

type text_line struct {
	s string
	e cairo.TextExtents
}

func (c *ctx) text_box(lines []text_line) (bounding_box r2.V) {
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
	bounding_box = r2.XY(max_width, fe.Height*float64(len(lines)))
	return
}

func (c *ctx) text(x r2.V, text_align int, lines ...text_line) r2.V {
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
		x1 := x + r2.XY(-bounding_box.X()*a+lines[i].e.XBearing, y0+float64(i)*fe.Height)
		cr.MoveTo(x1.X(), x1.Y())
		cr.ShowText(lines[i].s)
	}
	return bounding_box
}
func (c *ctx) textf(x r2.V, text_align int, format string, args ...interface{}) r2.V {
	var l [1]text_line
	l[0].s = fmt.Sprintf(format, args...)
	return c.text(x, text_align, l[:]...)
}

func (c *ctx) moveTo(x r2.V) { c.cr().MoveTo(x.X(), x.Y()) }
func (c *ctx) lineTo(x r2.V) { c.cr().LineTo(x.X(), x.Y()) }
func (c *ctx) rect(x0, dx r2.V) {
	c.cr().Rectangle(x0.X(), x0.Y(), dx.X(), dx.Y())
}

func (x *ctx) roundedRect(x0, dx r2.V, r float64) {
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
	win                  *gtk.Window
	screen_dpi           float64
	da                   *gtk.DrawingArea
	eb                   *gtk.EventBox
	eb_dx, eb_border     r2.V
	ev                   *elog.View
	ves                  []visible_event
	popup_by_event_index map[uint]*popup
	popups               []*popup
	m                    map[uintptr]*decoration
	pointer              pointer_state
	Config
}

type Config struct {
	EnableKeyboardQuit bool
	Width, Height      int
}

var initOnce sync.Once

func View(ev *elog.View, cf Config) {
	v := &viewer{ev: ev, Config: cf}

	initOnce.Do(func() { gtk.Init(nil) })

	v.eb_border = r2.XY(40, 60)
	v.eb_dx = r2.IJ(cf.Width, cf.Height)

	v.win, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	v.win.SetTitle("Event log")
	v.win.SetDefaultSize(v.eb_dx.IJ())
	v.win.Connect("destroy", v.quit)
	scr, _ := v.win.GetScreen()
	v.screen_dpi = scr.GetResolution()

	v.eb, _ = gtk.EventBoxNew()
	v.eb.SetCanFocus(true)
	v.eb.SetEvents(int(gdk.KEY_PRESS_MASK |
		gdk.BUTTON_PRESS_MASK | gdk.BUTTON_RELEASE_MASK | gdk.POINTER_MOTION_MASK))
	v.win.Add(v.eb)

	v.da, _ = gtk.DrawingAreaNew()
	v.eb.Add(v.da)

	v.da.Connect("draw", v.draw_events)
	v.eb.Connect("button_press_event", v.button_press)
	v.eb.Connect("button_release_event", v.button_release)
	v.eb.Connect("motion_notify_event", v.pointer_motion)
	v.eb.Connect("key_press_event", v.key_press)

	v.win.ShowAll()
	gtk.Main()
}

func (v *viewer) quit() {
	v.delete_all_popups()
	gtk.MainQuit()
}

func (v *viewer) key_press(eb *gtk.EventBox, ev *gdk.Event) {
	ke := &gdk.EventKey{ev}
	key, state := ke.KeyVal(), gdk.ModifierType(ke.State())
	t := &v.ev.Times

	subview := false
	t_min, t_max := t.Min, t.Max
	switch key {
	case gdk.KEY_Escape:
		break
	case gdk.KEY_q:
		if v.EnableKeyboardQuit { // quit does not work via key inside vnet
			v.quit()
		}
	case gdk.KEY_r:
		v.ev.Reset()
	case gdk.KEY_plus:
		dt := t.Dt * .05 // 10% smaller
		t_min += dt
		t_max -= dt
		subview = true
	case gdk.KEY_minus:
		dt := t.Dt * .05 // 10% bigger
		t_min -= dt
		t_max += dt
		subview = true
	case gdk.KEY_rightarrow, gdk.KEY_Right:
		dt := t.Unit
		if state&gdk.GDK_SHIFT_MASK != 0 {
			dt *= 10
		}
		t_min -= dt
		t_max -= dt
		subview = true
	case gdk.KEY_leftarrow, gdk.KEY_Left:
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

type pointer_state struct {
	val uint
	xs  []r2.V
}

func button_event(e *gdk.Event) (x r2.V, val uint) {
	be := gdk.EventButton{e}
	val = be.ButtonVal()
	bx, by := be.MotionVal()
	x = r2.XY(bx, by)
	return
}

func (v *viewer) button_press(eb *gtk.EventBox, e *gdk.Event) {
	handled, x, val := v.do_button_press(e)
	if !handled {
		p := &v.pointer
		p.val = val
		if p.xs != nil {
			p.xs = p.xs[:0]
		}
		p.xs = append(p.xs, x)
	}
}

func (v *viewer) pointer_motion(eb *gtk.EventBox, e *gdk.Event) {
	x, _ := button_event(e)
	p := &v.pointer
	if p.val != 0 {
		p.xs = append(p.xs, x)
		v.eb.QueueDraw()
	}
}

func (v *viewer) button_release(eb *gtk.EventBox, e *gdk.Event) {
	x, val := button_event(e)
	p := &v.pointer
	if p.val != 0 {
		p.xs = append(p.xs, x)
		p.val = 0
		v.do_pointer(val)
	}
}

func (v *viewer) do_pointer(val uint) {
	p := &v.pointer
	l := len(p.xs)
	if l < 3 { // button press + 1 motion + button release
		return
	}
	switch val {
	case 1:
		t0, t1 := v.x_to_time(p.xs[0].X()), v.x_to_time(p.xs[l-1].X())
		save := v.ev.Times
		if n := v.ev.SubView(t0, t1); n == 0 {
			v.ev.SubView(save.Min, save.Max)
		}
		v.delete_all_popups()
		v.eb.QueueDraw()
	}
}

func (v *viewer) do_button_press(e *gdk.Event) (handled bool, mouse_x r2.V, val uint) {
	mouse_x, val = button_event(e)
	if val != 1 {
		return
	}
	min_vi, min_ds := uint(0), 1e10
	for i := range v.ves {
		ve := &v.ves[i]
		if ds := ve.distance(mouse_x); ds < min_ds {
			min_ds = ds
			min_vi = uint(i)
			if ds == 0 {
				break
			}
		}
	}
	// Only accept when screen distance is less than 1/16 an inch of center.
	if handled = min_ds < v.screen_dpi/16; handled {
		// Event already displayed in popup?
		min_ei := v.ves[min_vi].ei
		if p, is_del := v.popup_by_event_index[min_ei]; is_del {
			p.del()
		} else {
			v.new_popup(min_vi, min_ei)
		}
	}
	return
}

func (v *viewer) order_popups() {
	if v.popups != nil {
		v.popups = v.popups[:0]
	}
	ps := v.popups
	for _, q := range v.popup_by_event_index {
		ps = append(ps, q)
	}
	sort.Slice(ps, func(i, j int) bool {
		pi, pj := ps[i], ps[j]
		return pi.vi < pj.vi
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
	v.popups = ps
}

const dt_prev_invalid = 1e10

type da_window struct {
	win *gtk.Window
	da  *gtk.DrawingArea
	eb  *gtk.EventBox
}

type popup struct {
	da_window
	v          *viewer
	ei, vi     uint
	l          []text_line
	prev, next *popup
}

func (p *popup) get_visible_event() *visible_event { return &p.v.ves[p.vi] }
func (p *popup) get_event() *elog.Event            { return &p.v.ev.Events[p.ei] }

func (v *viewer) new_da_window(d *da_window) {
	w, _ := gtk.WindowNew(gtk.WINDOW_POPUP)
	w.SetResizable(false)
	w.SetDecorated(false)
	w.SetDestroyWithParent(true)
	w.SetSizeRequest(int(v.eb_dx.X()), int(v.eb_dx.Y())) // max size of drawing area
	w.SetTransientFor(v.win)
	w.SetPosition(gtk.WIN_POS_CENTER_ON_PARENT)
	scr, _ := w.GetScreen()
	vis, _ := scr.GetRGBAVisual()
	w.SetVisual(vis)

	eb, _ := gtk.EventBoxNew()
	eb.SetCanFocus(true)
	eb.SetEvents(int(gdk.KEY_PRESS_MASK |
		gdk.BUTTON_PRESS_MASK | gdk.BUTTON_RELEASE_MASK | gdk.POINTER_MOTION_MASK))
	w.Add(eb)

	da, _ := gtk.DrawingAreaNew()
	eb.Add(da)

	d.win = w
	d.eb = eb
	d.da = da

	eb.SetEvents(int(gdk.BUTTON_PRESS_MASK | gdk.BUTTON_RELEASE_MASK | gdk.POINTER_MOTION_MASK))
	w.Connect("draw", v.draw_window_transparent_background)
	eb.Connect("button_press_event", v.button_press)
	eb.Connect("button_release_event", v.button_release)
	eb.Connect("motion_notify_event", v.pointer_motion)
	w.ShowAll()
}

func (v *viewer) draw_window_transparent_background(win *gtk.Window, cr *cairo.Context) {
	w := r2.XY(float64(win.GetAllocatedWidth()), float64(win.GetAllocatedHeight()))
	c := (*ctx)(cr)
	cr.SetOperator(cairo.OPERATOR_SOURCE)
	c.rect(0, w)
	cr.Fill()
}

func (v *viewer) new_popup(vi, ei uint) (p *popup) {
	p = &popup{v: v, vi: vi, ei: ei}
	v.new_da_window(&p.da_window)
	p.da.Connect("draw", p.draw_popup)
	if v.popup_by_event_index == nil {
		v.popup_by_event_index = make(map[uint]*popup)
	}
	v.popup_by_event_index[ei] = p
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
	delete(p.v.popup_by_event_index, p.ei)
	p.v.order_popups()
}
func (v *viewer) delete_all_popups() {
	for _, p := range v.popup_by_event_index {
		p.del()
	}
}

func (p0 *popup) dt(p1 *popup) (dt float64) {
	e0, e1 := p0.get_event(), p1.get_event()
	return p1.v.ev.ElapsedTime(e1) - p0.v.ev.ElapsedTime(e0)
}

func (p *popup) draw_popup(da *gtk.DrawingArea, cr *cairo.Context) {
	e, de := p.get_event(), p.get_visible_event()

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

	center := de.rr.Center
	const radius = 4
	rr_dx := bbox + r2.XY(2*radius, radius)
	c.roundedRect(center-rr_dx/2, rr_dx, radius)
	dc := d.color.lighten(.375)
	cr.SetSourceRGBA(dc.r, dc.g, dc.b, 1)
	cr.Fill()

	cr.SetSourceRGBA(0, 0, 0, 1)
	c.text(center-r2.XY(bbox.X()/2, 0), text_align_left, p.l...)
	cr.Stroke()
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

type visible_event struct {
	// Event index.
	ei uint
	// Center/size of rounded rect containing event text.
	rr r2.Rect
}

func (e *visible_event) distance(t r2.V) (ds float64) {
	if e.rr.IsInside(t) {
		return
	}

	c, dx2 := e.rr.Center, e.rr.Size/2

	// Distance to center.
	ds = (c - t).Abs()
	// Distance to 4 corners.
	for i := 0; i < 4; i++ {
		x, y := dx2.X(), dx2.Y()
		if i&1 != 0 {
			x = -x
		}
		if i&2 != 0 {
			y = -y
		}
		dx2 = r2.XY(x, y)
		if l := (c + dx2 - t).Abs(); l < ds {
			ds = l
		}
	}
	return
}

func (v *viewer) time_to_x(t float64) (x float64) {
	dw := v.eb_border
	w := v.eb_dx - 2*dw
	tb := &v.ev.Times
	x = dw.X() + (t-tb.Min)*w.X()/tb.Dt
	return
}
func (v *viewer) x_to_time(x float64) (t float64) {
	dw := v.eb_border
	w := v.eb_dx - 2*dw
	tb := &v.ev.Times
	t = tb.Min + (x-dw.X())*tb.Dt/w.X()
	return
}

func (c *ctx) time_axis(v *viewer, t float64, dx r2.V, text_align int, bg rgba, line bool, format string, args ...interface{}) {
	dw := v.eb_border
	w := v.eb_dx - 2*dw
	x := r2.XY(v.time_to_x(t), dw.Y())
	s := fmt.Sprintf(format, args...)

	c.cr().SetSourceRGBA(0, 0, 0, 1)
	c.cr().SetFontSize(10)
	c.textf(x+dx, text_align, s)
	c.cr().Stroke()

	if line {
		color := bg.darken(.85)
		c.cr().SetSourceRGBA(color.RGBA())
		c.moveTo(x)
		c.lineTo(x + r2.XY(0, w.Y()))
		c.cr().Stroke()
	}
}

func (v *viewer) draw_events(da *gtk.DrawingArea, cr *cairo.Context) {
	ev, tb := v.ev, &v.ev.Times

	v.eb_dx = r2.IJ(da.GetAllocatedWidth(), da.GetAllocatedHeight())
	dw := v.eb_border
	w := v.eb_dx - 2*dw

	t_min, t_max := tb.Min, tb.Max

	c := (*ctx)(cr)
	cr.SetAntialias(cairo.ANTIALIAS_SUBPIXEL)

	// Background
	bg_color := gray(.9, 1)
	c.rect(dw, w)
	cr.SetSourceRGBA(bg_color.RGBA())
	cr.Fill()

	// Axes
	c.time_axis(v, t_min, r2.XY(0, -10), text_align_center, bg_color, false, "%.0f%s", t_min/tb.Unit, tb.UnitName)
	c.time_axis(v, t_max, r2.XY(0, -10), text_align_center, bg_color, false, "%.0f%s", t_max/tb.Unit, tb.UnitName)

	// Pointer bounds.
	{
		p := &v.pointer
		if l := len(p.xs); p.val != 0 && l >= 2 {
			x0, x1 := p.xs[0].X(), p.xs[l-1].X()
			if x0 > x1 {
				x1, x0 = x0, x1
			}
			hilight_color := bg_color.darken(.9)
			cr.SetSourceRGBA(hilight_color.RGBA())
			c.rect(r2.XY(x0, dw.Y()), r2.XY(x1-x0, w.Y()))
			cr.Fill()
			t0, t1 := v.x_to_time(x0), v.x_to_time(x1)
			c.time_axis(v, t0, r2.XY(-5, 10), text_align_right, hilight_color, true, "%.2f", t0/tb.Unit)
			c.time_axis(v, t1, r2.XY(+5, 10), text_align_left, hilight_color, true, "%.2f", t1/tb.Unit)
		}
	}

	// Title
	{
		cr.SetSourceRGB(0, 0, 0)
		x := r2.XY(w.X()/2, dw.Y()/2)
		cr.SetFontSize(18)
		c.textf(x, text_align_center, "%d events, %s", len(ev.Events), tb.Start.Format("2006-01-02 15:04:05"))
		cr.Fill()
	}

	var x_last elib.Float64Vec

	if v.ves != nil {
		v.ves = v.ves[:0]
	}

	// Draw events.
	for i := range ev.Events {
		e := &ev.Events[i]
		t := ev.ElapsedTime(e)
		var ve visible_event

		x := r2.XY((t-t_min)/(t_max-t_min)*w.X(), w.Y()/2)
		if visible := x.X() >= 0 && x.X() < w.X(); !visible {
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
		if visible := iy <= iy_max; !visible {
			continue
		}

		ve.ei = uint(i)

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
		ve.rr.Center = center - r2.XY(0, (fe.Height+2*radius)*float64(idy))

		if visible := ve.rr.Center.Y() >= dw.Y() && ve.rr.Center.Y()+dw.Y() < w.Y(); !visible {
			continue
		}

		cr.SetSourceRGB(0, 0, 0)
		tw := c.textf(ve.rr.Center, text_align_center, lines[0])
		ve.rr.Size = tw + r2.XY(2*radius, radius)
		cr.Stroke()
		cr.SetSourceRGBA(d.color.r, d.color.g, d.color.b, d.color.a)
		c.roundedRect(ve.rr.Center-ve.rr.Size/2, ve.rr.Size, radius)
		cr.Fill()

		// (A) Plot event points on event line at iy == 0 and lines between text & event line.
		max_idy := 4
		if idy >= -max_idy && idy <= +max_idy {
			dot_radius := .9 * radius
			if idy > 0 {
				c.moveTo(ve.rr.Center + r2.XY(0, (ve.rr.Size/2).Y()))
				c.lineTo(center - r2.XY(0, dot_radius))
			} else {
				c.moveTo(ve.rr.Center - r2.XY(0, (ve.rr.Size/2).Y()))
				c.lineTo(center + r2.XY(0, dot_radius))
			}
			cr.SetOperator(cairo.OPERATOR_ATOP)
			cr.SetLineWidth(2)
			cr.Stroke()

			cr.SetSourceRGBA(d.color.r, d.color.g, d.color.b, .75)
			cr.MoveTo(center.X(), center.Y())
			cr.Arc(center.X(), center.Y(), dot_radius, 0, 2*math.Pi)
			cr.Fill()
		}

		// Save away for later use.
		v.ves = append(v.ves, ve)
	}
}
