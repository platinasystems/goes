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
	"os"
	"time"
)

// (x,y) coordinate as complex number for easy arithmetic.
type X2 complex128

func (x X2) X() float64   { return real(x) }
func (x X2) Y() float64   { return imag(x) }
func XY(x, y float64) X2  { return X2(complex(x, y)) }
func XX(x float64) X2     { return XY(x, x) }
func (x X2) Conj() X2     { return XY(real(x), -imag(x)) }
func (x X2) Abs() float64 { return math.Hypot(x.X(), x.Y()) }

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
	win *gtk.Window
	da  *gtk.DrawingArea
	eb  *gtk.EventBox
	ps  map[uint]*popup
	m   map[uintptr]*decoration
}

func main() {
	a := os.Args
	gtk.Init(&a)

	elog.Enable(true)
	for i := 0; i < 10; i++ {
		elog.GenEventf("red %d", i)
		time.Sleep(1 * time.Millisecond)
		elog.GenEventf("green %d", i)
		time.Sleep(1 * time.Millisecond)
		elog.GenEventf("blue %d", i)
		time.Sleep(1 * time.Millisecond)
		elog.GenEventf("yellow %d", i)
		time.Sleep(1 * time.Millisecond)
	}

	v := &viewer{}

	// gui boilerplate
	v.win, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	v.win.SetTitle("Event log")
	v.win.SetDefaultSize(1200, 750)
	v.win.Connect("destroy", gtk.MainQuit)

	v.eb, _ = gtk.EventBoxNew()
	v.eb.SetCanFocus(true)
	v.eb.SetEvents(int(gdk.KEY_PRESS_MASK | gdk.ENTER_NOTIFY_MASK | gdk.LEAVE_NOTIFY_MASK | gdk.BUTTON_PRESS_MASK))
	v.eb.SetBorderWidth(20)
	v.win.Add(v.eb)

	v.da, _ = gtk.DrawingAreaNew()
	v.eb.Add(v.da)

	v.win.ShowAll()

	v.da.Connect("draw", v.draw)
	v.eb.Connect("enter_notify_event", v.eb_enter)
	v.eb.Connect("leave_notify_event", v.eb_leave)
	v.eb.Connect("key_press_event", v.key_press)
	v.eb.Connect("button_press_event", v.button_press)
	v.eb.GrabFocus()
	gtk.Main()
}

func (v *viewer) key_press(eb *gtk.EventBox, ev *gdk.Event) {
	keyMap := map[uint]func(){
		KEY_q:      func() { gtk.MainQuit() },
		KEY_Escape: v.hide_popups,
	}
	ke := &gdk.EventKey{ev}
	kv := ke.KeyVal()
	if move, found := keyMap[kv]; found {
		move()
		eb.QueueDraw()
	}
}

func (v *viewer) button_press(eb *gtk.EventBox, e *gdk.Event) {
	be := &gdk.EventButton{e}
	bv := be.ButtonVal()
	bx, by := be.MotionVal()
	fmt.Println("press", bx, by, bv)
	if bv != 1 {
		return
	}
	mouse_x := XY(bx, by)
	v.do_button_press(eb, mouse_x)
}

func (v *viewer) do_button_press(eb *gtk.EventBox, mouse_x X2) {
	ev := elog.NewView()
	var tb elog.TimeBounds
	ev.GetTimeBounds(&tb)

	dw := XX(margin)
	w := XY(float64(eb.GetAllocatedWidth()), float64(eb.GetAllocatedHeight()))
	w -= dw

	t := tb.Min + tb.Dt*(mouse_x.X()-dw.X())/w.X()
	fmt.Println("time", t)

	var (
		min_e *elog.Event
		min_i int
	)
	min_dt := 1e10
	for i := range ev.Events {
		e := &ev.Events[i]
		if dt := math.Abs(t - ev.Time(e).Sub(tb.Start).Seconds()); dt < min_dt {
			min_dt = dt
			min_e = e
			min_i = i
		}
	}
	if min_e != nil {
		fmt.Println("min event", min_dt, ev.EventString(min_e))
	}
	v.new_popup(ev, min_e, mouse_x, uint(min_i))
}

func (v *viewer) eb_enter(eb *gtk.EventBox, ev *gdk.Event) {
	fmt.Println("enter", ev)
	eb.GrabFocus()
}

func (v *viewer) eb_leave(eb *gtk.EventBox, ev *gdk.Event) {
	fmt.Println("leave", ev)
	v.win.GrabFocus()
}

type popup struct {
	vr    *viewer
	win   *gtk.Window
	da    *gtk.DrawingArea
	e     *elog.Event
	v     *elog.View
	l     []text_line
	x, dx X2
}

func (v *viewer) new_popup(ev *elog.View, e *elog.Event, mouse_x X2, event_index uint) (p *popup) {
	// Event already displayed in popup?
	if _, ok := v.ps[event_index]; ok {
		return
	}

	p = &popup{vr: v, v: ev, e: e}
	w, _ := gtk.WindowNew(gtk.WINDOW_POPUP)
	p.win = w
	w.SetResizable(false)
	w.SetDecorated(false)
	p.dx = XY(400, 300)
	p.x = mouse_x - p.dx/2                         // window is centered on mouse
	w.SetSizeRequest(int(p.dx.X()), int(p.dx.Y())) // max size of drawing area
	w.SetTransientFor(v.win)
	w.SetPosition(gtk.WIN_POS_MOUSE)
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
	p.win.ShowAll()
	return
}

func (p *popup) hide() {
	if p.win == nil {
		return
	}
	p.win.Hide()
	p.win.Destroy()
	v := p.vr
	*p = popup{vr: v}
}
func (v *viewer) hide_popups() {
	for _, p := range v.ps {
		p.hide()
	}
	v.ps = nil
}

func (p *popup) button_press(win *gtk.Window, e *gdk.Event) {
	be := &gdk.EventButton{e}
	bv := be.ButtonVal()
	bx, by := be.MotionVal()
	fmt.Println("popup button press", bx, by, bv)
	if bv != 1 {
		return
	}
	mouse_x := p.x + XY(bx, by)
	p.vr.do_button_press(p.vr.eb, mouse_x)
}

func (p *popup) draw_popup(da *gtk.DrawingArea, cr *cairo.Context) {
	lines := p.e.Strings()
	_, pc := p.v.EventPath(p.e)
	d, _ := p.vr.decorationForPc(pc)

	var tb elog.TimeBounds
	p.v.GetTimeBounds(&tb)
	elapsed := p.v.Time(p.e).Sub(tb.Start).Seconds()
	lines = append(lines,
		fmt.Sprintf("%.2f%s", elapsed/tb.Unit, tb.UnitName))

	if p.l != nil {
		p.l = p.l[:0]
	}
	for i := range lines {
		p.l = append(p.l, text_line{s: lines[i]})
	}

	w := XY(float64(da.GetAllocatedWidth()), float64(da.GetAllocatedHeight()))
	center := w / 2

	c := (*ctx)(cr)
	cr.SetFontSize(10)
	bbox := c.text_box(p.l)

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

const margin = 40

func (v *viewer) draw(da *gtk.DrawingArea, cr *cairo.Context) {
	ev := elog.NewView()
	var tb elog.TimeBounds
	ev.GetTimeBounds(&tb)
	fmt.Printf("foo %+v\n", &tb)

	c := (*ctx)(cr)

	dw := XX(margin)
	w := XY(float64(da.GetAllocatedWidth()), float64(da.GetAllocatedHeight()))
	w -= dw

	// Title
	{
		cr.SetSourceRGB(0, 0, 0)
		x := XY(w.X()/2, 20)
		cr.SetFontSize(18)
		c.textf(x, text_align_center, tb.Start.Format("2006-01-02 15:04:05"))
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
		c.textf(dw-10i, text_align_left, "%.1f%s", t_min/t_unit, tb.UnitName)
		cr.Stroke()
		c.textf(tmp-10i, text_align_right, "%.1f%s", t_max/t_unit, tb.UnitName)
		cr.Stroke()
	}

	var x_last elib.Float64Vec

	cr.SetAntialias(cairo.ANTIALIAS_SUBPIXEL)

	// Draw events.
	for i := range ev.Events {
		e := &ev.Events[i]
		t := ev.Time(e).Sub(tb.Start).Seconds()

		x := XY((t-t_min)/(t_max-t_min)*w.X(), w.Y()/2)
		if x.X() < 0 || x.X() > w.X() {
			continue
		}

		lines := e.Strings()
		_, pc := ev.EventPath(e)
		d, _ := v.decorationForPc(pc)

		center := dw + x
		radius := 4.
		dx := XY(radius, radius)
		cr.SetSourceRGBA(d.color.r, d.color.g, d.color.b, d.color.a)
		switch 3 {
		case 3:
			cr.Save()
			cr.SetFontSize(9)
			fe, te := cr.FontExtents(), cr.TextExtents(lines[0])

			// Choose integer Y such that text will not overlap with other text.
			iy := uint(0)
			for {
				x_last.Validate(uint(iy))
				if x_left := center.X() - te.Width/2; x_last[iy]+2*radius < x_left {
					x_last[iy] = x_left + te.Width
					break
				}
				iy++
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
			cr.SetSourceRGB(0, 0, 0)
			tw := c.textf(rr_center, text_align_center, lines[0])
			rr_dx := tw + XY(2*radius, radius)
			cr.Stroke()
			cr.SetSourceRGBA(d.color.r, d.color.g, d.color.b, d.color.a)
			c.roundedRect(rr_center-rr_dx/2, rr_dx, radius)
			cr.Fill()

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
			cr.Restore()

			// stale code
		case 0:
			cr.Arc(center.X(), center.Y(), radius, 0, 2*math.Pi)
			cr.Fill()
		case 1:
			c.rect(center-dx, 2*dx)
			cr.Fill()
		case 2:
			rx, ry := radius*.9, radius*1.2
			cr.MoveTo(center.X()-rx, center.Y())
			cr.LineTo(center.X(), center.Y()-ry)
			cr.LineTo(center.X()+rx, center.Y())
			cr.LineTo(center.X(), center.Y()+ry)
			cr.ClosePath()
			cr.Fill()
		default:
			panic("ga")
		}
	}

	v.eb.GrabFocus()
}

func (v *viewer) draw_window_transparent_background(win *gtk.Window, cr *cairo.Context) {
	w := XY(float64(win.GetAllocatedWidth()), float64(win.GetAllocatedHeight()))
	c := (*ctx)(cr)
	cr.SetOperator(cairo.OPERATOR_SOURCE)
	c.rect(0, w)
	cr.Fill()
}
