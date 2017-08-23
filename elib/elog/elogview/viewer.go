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
	"regexp"
	"runtime"
	"strings"
	"sync"
)

type ctx cairo.Context

func (c *ctx) cr() *cairo.Context { return (*cairo.Context)(c) }
func (c *ctx) translate(x r2.V)   { c.cr().Translate(x.X(), x.Y()) }
func (c *ctx) scale(x r2.V)       { c.cr().Scale(x.X(), x.Y()) }

type rgb struct{ r, g, b float64 }
type rgba struct {
	rgb
	a float64
}

func Rgb(r, g, b float64) rgb    { return rgb{r: r, g: g, b: b} }
func Rgba(r rgb, a float64) rgba { return rgba{rgb: r, a: a} }

func (x *rgb) RGB() (r, g, b float64)      { r, g, b = x.r, x.g, x.b; return }
func (x *rgba) RGBA() (r, g, b, a float64) { r, g, b, a = x.r, x.g, x.b, x.a; return }

func gray(f float64) rgb { return rgb{r: f, g: f, b: f} }
func (x *rgb) gray(f float64) {
	x.r = 1 - f
	x.g = 1 - f
	x.b = 1 - f
}

func grayWheel(f0 float64, i, n int) (r rgb) {
	dx := (.5 - f0) / float64(n/2)
	r.gray(f0 + float64(i)*dx)
	return
}

// Mix primary r0 with a bit of secondary r1 (specified by fraction f1 in [0,1]).
func (r0 rgb) mix(r1 rgb, f1 float64) (r rgb) {
	r.r = f1*r1.r + (1-f1)*r0.r
	r.g = f1*r1.g + (1-f1)*r0.g
	r.b = f1*r1.b + (1-f1)*r0.b
	return
}

var (
	black = rgb{r: 0, g: 0, b: 0}
	white = rgb{r: 1, g: 1, b: 1}
)

func (x rgb) darken(f float64) rgb  { return x.mix(black, f) }
func (x rgb) lighten(f float64) rgb { return x.mix(white, f) }

var standardColors = [...]decoration{
	decoration{ // dark blue
		bg: Rgb(.05, .24, .33),
	},
	decoration{ // dark blue
		bg: Rgb(.06, .36, .47),
	},
	decoration{ // light blue
		bg: Rgb(.07, .47, .60),
	},
	decoration{ // light blue
		bg: Rgb(.07, .58, .73),
	},
	decoration{ // teal
		bg: Rgb(.36, .65, .58),
	},
	decoration{
		bg: Rgb(.64, .72, .42),
	},
	decoration{
		bg: Rgb(.92, .78, .27),
	},
	decoration{ // yellow
		bg: Rgb(.93, .67, .22),
	},
	decoration{ // orange
		bg: Rgb(.94, .55, .17),
	},
	decoration{ // orange
		bg: Rgb(.95, .42, .13),
	},
	decoration{ // red
		bg: Rgb(.85, .31, .12),
	},
	decoration{ // red
		bg: Rgb(.75, .18, .11),
	},
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
		if w := te.Width; w > max_width {
			max_width = w
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
		x1 := x + r2.XY(-lines[i].e.Width*a+lines[i].e.XBearing, y0+float64(i)*fe.Height)
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
	win              *gtk.Window
	vbox             *gtk.Box
	screen_dpi       float64
	heading_da       *gtk.DrawingArea
	da               *gtk.DrawingArea
	eb               *gtk.EventBox
	eb_dx, eb_border r2.V
	ev               *elog.View
	ves              []visible_event
	selected_events  map[uint]struct{}
	hidden_callers   map[uint]struct{}
	m                map[uint64]*decoration
	pointer          pointer_state
	filters
	Config
}

type Config struct {
	EnableKeyboardQuit bool
	Width, Height      int
}

var initOnce sync.Once

func View(ev *elog.View, cf Config) {
	v := &viewer{ev: ev, Config: cf}

	runtime.LockOSThread()
	initOnce.Do(func() { gtk.Init(nil) })

	v.eb_border = r2.XY(40, 20)
	v.eb_dx = r2.IJ(cf.Width, cf.Height)

	v.win, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	{
		title := "Event log"
		if n := ev.Name(); n != "" {
			title += " " + n
		}
		v.win.SetTitle(title)
	}
	v.win.SetDefaultSize(v.eb_dx.IJ())
	v.win.Connect("destroy", v.quit)
	scr, _ := v.win.GetScreen()
	v.screen_dpi = scr.GetResolution()

	v.vbox, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	v.win.Add(v.vbox)

	v.heading_da, _ = gtk.DrawingAreaNew()
	v.heading_da.SetSizeRequest(v.eb_dx.I(), 60)
	v.heading_da.Connect("draw", v.draw_heading)
	v.vbox.PackStart(v.heading_da, false, false, 0)

	v.eb, _ = gtk.EventBoxNew()
	v.eb.SetSizeRequest(v.eb_dx.I(), int(.9*v.eb_dx.Y()))
	v.eb.SetCanFocus(true)
	v.eb.SetEvents(int(gdk.KEY_PRESS_MASK |
		gdk.BUTTON_PRESS_MASK | gdk.BUTTON_RELEASE_MASK | gdk.POINTER_MOTION_MASK))

	v.da, _ = gtk.DrawingAreaNew()
	v.eb.Add(v.da)
	v.vbox.PackStart(v.eb, true, true, 0)

	v.da.Connect("draw", v.draw_events)
	v.eb.Connect("button_press_event", v.button_press)
	v.eb.Connect("button_release_event", v.button_release)
	v.eb.Connect("motion_notify_event", v.pointer_motion)
	v.eb.Connect("key_press_event", v.key_press)

	v.filters.init(v)

	v.win.ShowAll()
	gtk.Main()
}

func (v *viewer) quit() { gtk.MainQuit() }

func (v *viewer) key_press(eb *gtk.EventBox, ev *gdk.Event) {
	ke := &gdk.EventKey{ev}
	key, state := ke.KeyVal(), gdk.ModifierType(ke.State())
	t := &v.ev.Times

	subview := false
	t_min, t_max := t.Min, t.Max
	switch key {
	case gdk.KEY_Escape:
		for ci, _ := range v.hidden_callers {
			delete(v.hidden_callers, ci)
		}
		for ei, _ := range v.selected_events {
			delete(v.selected_events, ei)
		}
		break
	case gdk.KEY_q:
		if v.EnableKeyboardQuit { // quit does not work via key inside vnet
			v.quit()
		}
	case gdk.KEY_BackSpace:
		for ei, _ := range v.selected_events {
			e := &v.ev.Events[ei]
			if v.hidden_callers == nil {
				v.hidden_callers = make(map[uint]struct{})
			}
			v.hidden_callers[e.GetCaller()] = struct{}{}
			delete(v.selected_events, ei)
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
		min_ei := v.ves[min_vi].ei
		if v.is_selected(min_ei) {
			delete(v.selected_events, min_ei)
		} else {
			if v.selected_events == nil {
				v.selected_events = make(map[uint]struct{})
			}
			v.selected_events[min_ei] = struct{}{}
		}
		v.da.QueueDraw()
	}
	return
}

func (v *viewer) is_selected(ei uint) (ok bool) {
	_, ok = v.selected_events[ei]
	return
}

func (v *viewer) draw_selected_event(cr *cairo.Context, e *elog.Event, ve *visible_event) {
	ev := v.ev
	lines := e.Strings(ev.GetContext())

	// Indent lines after first.
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
		if i > 0 {
			lines[i] = "  " + strings.TrimSpace(lines[i])
		}
	}

	ec := ev.EventCaller(e)
	d, _ := v.decorationForCaller(ec)

	sf, _ := ec.ShortPath(ec.File, 32)
	sn, _ := ec.ShortPath(ec.Name, 32)
	lines = append(lines, fmt.Sprintf("%s", sn))
	lines = append(lines, fmt.Sprintf("%s: %d", sf, ec.Line))

	if ve.l != nil {
		ve.l = ve.l[:0]
	}
	for i := range lines {
		ve.l = append(ve.l, text_line{s: lines[i]})
	}

	c := (*ctx)(cr)
	cr.SetFontSize(12)
	bbox := c.text_box(ve.l)

	center := ve.rr.Center
	const radius = 8
	rr_dx := bbox + 2*r2.XY(radius, radius)
	upper_left := center - rr_dx/2
	c.roundedRect(upper_left, rr_dx, radius)
	bg := d.bg.lighten(.75)
	c.set_pattern(upper_left, rr_dx, bg, .15)
	cr.Fill()

	fg := grayWheel(.06, 8, 9)
	fg = fg.mix(d.bg, .25)
	cr.SetSourceRGB(fg.RGB())
	c.text(center-r2.XY(bbox.X()/2, 0), text_align_left, ve.l...)
	cr.Stroke()

	ve.rr.Size = rr_dx
}

type decoration struct {
	bg rgb
}

func (c *ctx) set_pattern(x, dx r2.V, bg rgb, f float64) {
	p, _ := cairo.NewPatternLinear(x.X(), x.Y(), x.X(), x.Y()+dx.Y())
	l, d := bg.lighten(f), bg.darken(f)
	p.AddColorStopRGBA(0, l.r, l.g, l.b, 1)
	p.AddColorStopRGBA(1, d.r, d.g, d.b, 1)
	c.cr().SetSource(p)
}

func (v *viewer) decorationForCaller(ci *elog.CallerInfo) (d *decoration, ok bool) {
	pc := ci.PC
	if v.m == nil {
		v.m = make(map[uint64]*decoration)
	}
	if d, ok = v.m[pc]; !ok {
		i := len(v.m)
		i0 := i % len(standardColors)
		d = &standardColors[i0]
		v.m[pc] = d
	}
	return
}

type visible_event struct {
	// Event index.
	ei uint
	// Center/size of rounded rect containing event text.
	rr r2.Rect
	l  []text_line
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

func (c *ctx) time_axis(v *viewer, t float64, dx r2.V, text_align int, bg rgb, line bool, format string, args ...interface{}) {
	dw := v.eb_border
	w := v.eb_dx - 2*dw
	x := r2.XY(v.time_to_x(t), dw.Y())
	s := fmt.Sprintf(format, args...)

	c.cr().SetSourceRGB(0, 0, 0)
	c.cr().SetFontSize(10)
	c.textf(x+dx, text_align, s)
	c.cr().Stroke()

	if line {
		color := bg.darken(.15)
		c.cr().SetSourceRGB(color.RGB())
		c.moveTo(x)
		c.lineTo(x + r2.XY(0, w.Y()))
		c.cr().Stroke()
	}
}

func (v *viewer) draw_heading(da *gtk.DrawingArea, cr *cairo.Context) {
	c := (*ctx)(cr)
	ev, tb := v.ev, &v.ev.Times
	w := r2.IJ(da.GetAllocatedWidth(), da.GetAllocatedHeight())

	cr.SelectFontFace("Mono", cairo.FONT_SLANT_NORMAL, cairo.FONT_WEIGHT_NORMAL)
	cr.SetFontSize(16)
	cr.SetSourceRGB(black.RGB())
	var l [2]text_line
	if n := ev.Name(); len(n) > 0 {
		l[0].s = n + ": "
	}
	l[0].s += fmt.Sprintf("%d events", len(ev.Events))
	l[1].s = tb.Start.Format("2006-01-02 15:04:05:000000")
	bbox := c.text_box(l[:])
	c.text(r2.XY(w.X()/2, w.Y()-bbox.Y()/2), text_align_center, l[:]...)
	cr.Stroke()
}

func (v *viewer) draw_events(da *gtk.DrawingArea, cr *cairo.Context) {
	ev, tb := v.ev, &v.ev.Times

	v.eb_dx = r2.IJ(da.GetAllocatedWidth(), da.GetAllocatedHeight())
	dw := v.eb_border
	w := v.eb_dx - 2*dw

	t_min, t_max := tb.Min, tb.Max

	c := (*ctx)(cr)
	cr.SelectFontFace("Mono", cairo.FONT_SLANT_NORMAL, cairo.FONT_WEIGHT_NORMAL)
	cr.SetAntialias(cairo.ANTIALIAS_SUBPIXEL)

	// Background
	bg_color := grayWheel(.06, 0, 9)
	draw_axis_line := true

	// Axes
	axis_dx := r2.XY(0, -10)
	c.time_axis(v, t_min, axis_dx, text_align_center, bg_color, draw_axis_line, "%.0f%s", t_min/tb.Unit, tb.UnitName)
	c.time_axis(v, t_max, axis_dx, text_align_center, bg_color, draw_axis_line, "%.0f%s", t_max/tb.Unit, tb.UnitName)

	// Title
	if false {
		cr.SetSourceRGB(0, 0, 0)
		x := r2.XY(w.X()/2, dw.Y()/2)
		cr.SetFontSize(18)
		var l [2]text_line
		if n := ev.Name(); len(n) > 0 {
			l[0].s = n + ": "
		}
		l[0].s += fmt.Sprintf("%d events", len(ev.Events))
		l[1].s = tb.Start.Format("2006-01-02 15:04:05:000000")
		c.text(x, text_align_center, l[:]...)
		cr.Stroke()
	}

	if v.ves != nil {
		v.ves = v.ves[:0]
	}

	// Draw events.
	var x_last elib.Float64Vec
	for i := range ev.Events {
		e := &ev.Events[i]
		t := ev.ElapsedTime(e)
		var ve visible_event

		x := r2.XY((t-t_min)/(t_max-t_min)*w.X(), w.Y()/2)
		if x_visible := x.X() >= 0 && x.X() < w.X(); !x_visible {
			continue
		}

		if _, ok := v.hidden_callers[e.GetCaller()]; ok {
			continue
		}

		lines := e.Strings(ev.GetContext())
		ci := ev.EventCaller(e)
		d, _ := v.decorationForCaller(ci)

		center := dw + x
		radius := 4.
		cr.SetSourceRGB(d.bg.RGB())
		cr.SetFontSize(10)
		fe, te := cr.FontExtents(), cr.TextExtents(lines[0])

		var tl [1]text_line
		ve.l = tl[:]
		ve.l[0].s = lines[0]
		bbox := c.text_box(ve.l)
		ve.rr.Size = bbox + r2.XY(4*radius, radius)
		rr_size2 := ve.rr.Size / 2

		// Choose integer Y such that text will not overlap with other text.
		var (
			idy       int
			y_visible bool
		)
		y_level := uint(0)
		for {
			// Avoid level 0 which will be used to plot event points.  See (A) below.
			idy = int(y_level + 1)
			// Odd levels go above baseline; even go below.
			if idy%2 != 0 {
				idy = (idy + 1) / 2
			} else {
				idy = -idy / 2
			}

			// Rounded rect with event text inside.
			ve.rr.Center = center - r2.XY(0, (fe.Height+2*radius)*float64(idy))
			if y_visible = ve.rr.Center.Y()-rr_size2.Y() >= dw.Y() &&
				ve.rr.Center.Y()+rr_size2.Y() <= w.Y()+dw.Y(); !y_visible {
				break
			}

			x_last.ValidateInit(uint(y_level), -1e10)
			if x_left := center.X() - te.Width/2; x_last[y_level]+2*radius < x_left {
				x_last[y_level] = center.X() + te.Width/2
				break
			}
			y_level++
		}

		ve.ei = uint(i)
		if y_visible {
			upper_left := ve.rr.Center - ve.rr.Size/2
			bg := d.bg.lighten(.5)
			fg := black
			f := .1
			c.set_pattern(upper_left, ve.rr.Size, bg, f)
			c.roundedRect(upper_left, ve.rr.Size, radius)
			cr.Fill()
			cr.SetSourceRGB(fg.RGB())
			c.textf(ve.rr.Center, text_align_center, lines[0])
			cr.Stroke()
		}

		// (A) Plot event points on event line at iy == 0 and lines between text & event line.
		{
			cr.Save()
			const max_idy = 6
			show_line := y_visible && idy >= -max_idy && idy <= +max_idy
			dot_radius := .9 * radius
			if show_line {
				if idy > 0 {
					c.moveTo(ve.rr.Center + r2.XY(0, (ve.rr.Size/2).Y()))
					c.lineTo(center - r2.XY(0, dot_radius))
				} else {
					c.moveTo(ve.rr.Center - r2.XY(0, (ve.rr.Size/2).Y()))
					c.lineTo(center + r2.XY(0, dot_radius))
				}
				cr.SetOperator(cairo.OPERATOR_OVER)
				cr.SetSourceRGBA(d.bg.r, d.bg.g, d.bg.b, .5)
				cr.SetLineWidth(2)
				cr.Stroke()
			}

			cr.MoveTo(center.X(), center.Y())
			cr.Arc(center.X(), center.Y(), dot_radius, 0, 2*math.Pi)
			cr.SetSourceRGBA(d.bg.r, d.bg.g, d.bg.b, 1)
			cr.Fill()
			cr.Restore()
		}

		// Save away for later use.
		if y_visible {
			v.ves = append(v.ves, ve)
		}
	}

	// Selected events.
	for i := range v.ves {
		ve := &v.ves[i]
		if v.is_selected(ve.ei) {
			e := &ev.Events[ve.ei]
			v.draw_selected_event(cr, e, ve)
		}
	}

	// Pointer bounds.
	{
		p := &v.pointer
		if l := len(p.xs); p.val != 0 && l >= 4 {
			x0, x1 := p.xs[0].X(), p.xs[l-1].X()
			if x0 > x1 {
				x1, x0 = x0, x1
			}
			bg, fg := grayWheel(.06, 4, 9), grayWheel(.06, 2, 9)
			cr.SetSourceRGBA(bg.r, bg.g, bg.b, .25)
			c.rect(r2.XY(x0, dw.Y()), r2.XY(x1-x0, w.Y()))
			cr.Fill()
			t0, t1 := v.x_to_time(x0), v.x_to_time(x1)
			const dy = 10
			c.time_axis(v, t0, r2.XY(-dy/2, axis_dx.Y()), text_align_right, fg, true, "%.1f%s", t0/tb.Unit, tb.UnitName)
			c.time_axis(v, t1, r2.XY(+dy/2, axis_dx.Y()), text_align_left, fg, true, "%.1f%s", t1/tb.Unit, tb.UnitName)
			if x1-x0 > 30 {
				c.cr().SetSourceRGBA(0, 0, 0, 1)
				c.textf(r2.XY(.5*(x0+x1), dw.Y()+axis_dx.Y()), text_align_center, "%.2f%s", (t1-t0)/tb.Unit, tb.UnitName)
				cr.Stroke()
			}
		}
	}
}

type filters struct {
	v              *viewer
	hbox           *gtk.Box
	add_button     *gtk.Button
	entry          *gtk.Entry
	filter_by_name map[string]*filter
}

type filter struct {
	m            *filters
	name         string
	check_button *gtk.CheckButton
	re           *regexp.Regexp
}

func (f *filter) matchEvent(e *elog.Event) (ok bool) {
	ev := f.m.v.ev
	ec := ev.EventCaller(e)
	lines := e.Strings(ev.GetContext())
	if ok = f.re.MatchString(lines[0]); ok {
		return
	}
	if ok = f.re.MatchString(ec.Name); ok {
		return
	}
	return
}
func (f *filter) nMatching() (n uint) {
	v := f.m.v
	for i := range v.ev.Events {
		e := &v.ev.Events[i]
		if f.matchEvent(e) {
			n++
		}
	}
	return
}

func (f *filter) check_button_clicked(b *gtk.CheckButton, e *gdk.Event) {
	be := gdk.EventButton{e}
	if be.ButtonVal() == 3 {
		f.check_button.Destroy()
		delete(f.m.filter_by_name, f.name)
		f.m.hbox.ShowAll()
		f.m.v.eb.QueueDraw()
		return
	}
}

func (m *filters) button_clicked(b *gtk.Button, e *gdk.Event) {
	buf, _ := m.entry.GetBuffer()

	name, _ := buf.GetText()
	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return
	}

	if _, ok := m.filter_by_name[name]; ok {
		return
	}

	f := &filter{m: m, name: name}

	var err error
	f.re, err = regexp.Compile(name)

	label := f.name
	if err != nil {
		label += " (invalid: " + err.Error() + ")"
	} else {
		label += fmt.Sprintf(" (%d)", f.nMatching())
	}
	cb, _ := gtk.CheckButtonNewWithLabel(label)
	cb.SetActive(err == nil)
	m.hbox.PackStart(cb, false, false, 10)

	if m.filter_by_name == nil {
		m.filter_by_name = make(map[string]*filter)
	}
	m.filter_by_name[f.name] = f

	f.check_button = cb
	cb.Connect("button_press_event", f.check_button_clicked)
	m.hbox.ShowAll()
	m.v.eb.QueueDraw()
}

func (m *filters) init(v *viewer) {
	m.v = v
	m.hbox, _ = gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)

	b, _ := gtk.ButtonNewWithLabel("Add Filter")
	b.SetSizeRequest(30, 30)
	m.add_button = b
	m.hbox.PackStart(b, false, false, 0)
	b.Connect("button_press_event", m.button_clicked)

	e, _ := gtk.EntryNew()
	e.SetSizeRequest(50, 30)
	m.entry = e
	m.hbox.PackStart(e, false, false, 10)
	v.vbox.PackStart(m.hbox, false, false, 0)
}
