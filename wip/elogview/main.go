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

func (c *ctx) text2(x, a X2, format string, args ...interface{}) (dx X2) {
	cr := c.cr()
	s := fmt.Sprintf(format, args...)
	fe, te := cr.FontExtents(), cr.TextExtents(s)
	dy := te.Height*a.Y() + te.YBearing
	if true {
		dy = fe.Ascent - .5*fe.Height
	}
	x1 := x + XY(-te.Width*a.X()+te.XBearing, dy)
	cr.MoveTo(x1.X(), x1.Y())
	cr.ShowText(s)
	dx = XY(te.Width, fe.Height)
	return
}
func (c *ctx) text(x X2, format string, args ...interface{}) X2 {
	return c.text2(x, XY(.5, .5), format, args...)
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

	lv := &elogViewer{}

	// gui boilerplate
	lv.win, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	lv.win.SetTitle("Event log")
	lv.win.SetDefaultSize(1200, 750)
	lv.win.Connect("destroy", gtk.MainQuit)

	{
		w, _ := gtk.WindowNew(gtk.WINDOW_POPUP)
		lv.p.win = w
		w.SetResizable(false)
		w.SetDecorated(false)
		w.SetSizeRequest(150, 40)
		w.SetTransientFor(lv.win)
		w.SetPosition(gtk.WIN_POS_MOUSE)

		da, _ := gtk.DrawingAreaNew()
		lv.p.da = da
		w.Add(da)
	}

	lv.eb, _ = gtk.EventBoxNew()
	lv.eb.SetCanFocus(true)
	lv.eb.SetEvents(int(gdk.KEY_PRESS_MASK | gdk.ENTER_NOTIFY_MASK | gdk.LEAVE_NOTIFY_MASK | gdk.BUTTON_PRESS_MASK))
	lv.eb.SetBorderWidth(20)
	lv.win.Add(lv.eb)

	lv.da, _ = gtk.DrawingAreaNew()
	lv.eb.Add(lv.da)

	lv.win.ShowAll()

	lv.eb.Connect("draw", lv.draw)
	lv.eb.Connect("enter_notify_event", lv.eb_enter)
	lv.eb.Connect("leave_notify_event", lv.eb_leave)
	lv.eb.Connect("key_press_event", lv.key_press)
	lv.eb.Connect("button_press_event", lv.button_press)
	lv.p.win.Connect("button_press_event", lv.popup_button_press)
	lv.eb.GrabFocus()

	gtk.Main()
}

func (lv *elogViewer) key_press(eb *gtk.EventBox, ev *gdk.Event) {
	keyMap := map[uint]func(){
		KEY_q: func() { gtk.MainQuit() },
	}
	ke := &gdk.EventKey{ev}
	kv := ke.KeyVal()
	if move, found := keyMap[kv]; found {
		move()
		eb.QueueDraw()
	}
}

func (lv *elogViewer) button_press(eb *gtk.EventBox, e *gdk.Event) {
	be := &gdk.EventButton{e}
	x, y := be.MotionVal()
	bv := be.ButtonVal()
	fmt.Println("press", x, y, bv)
	if bv != 1 {
		return
	}

	ev := elog.NewView()
	var tb elog.TimeBounds
	ev.GetTimeBounds(&tb)

	dw := XX(margin)
	w := XY(float64(eb.GetAllocatedWidth()), float64(eb.GetAllocatedHeight()))
	w -= dw

	t := tb.Min + tb.Dt*(x-dw.X())/w.X()
	fmt.Println("time", t)

	var min_e *elog.Event
	min_dt := 1e10
	for i := range ev.Events {
		e := &ev.Events[i]
		if dt := math.Abs(t - ev.Time(e).Sub(tb.Start).Seconds()); dt < min_dt {
			min_dt = dt
			min_e = e
		}
	}
	if min_e != nil {
		fmt.Println("min event", min_dt, ev.EventString(min_e))
	}
	lv.p.show(ev, min_e, &tb)
}

func (lv *elogViewer) popup_button_press(popup *gtk.Window, e *gdk.Event) {
	be := &gdk.EventButton{e}
	bv := be.ButtonVal()
	fmt.Println("popup button press")
	if bv == 1 {
		lv.p.hide()
	}
}

func (lv *elogViewer) eb_enter(eb *gtk.EventBox, ev *gdk.Event) {
	fmt.Println("enter", ev)
	eb.GrabFocus()
}

func (lv *elogViewer) eb_leave(eb *gtk.EventBox, ev *gdk.Event) {
	fmt.Println("leave", ev)
	lv.win.GrabFocus()
}

type popup struct {
	win *gtk.Window
	da  *gtk.DrawingArea
}

func (p *popup) show(v *elog.View, e *elog.Event, tb *elog.TimeBounds) {
	text := ""
	lines := e.Strings()
	for i := range lines {
		if text != "" {
			text += "\n"
		}
		text += lines[i]
	}
	elapsed := v.Time(e).Sub(tb.Start).Seconds()
	text += fmt.Sprintf("\n%.2f%s", elapsed/tb.Unit, tb.UnitName)
	fmt.Println(text)
	p.win.ShowAll()
}
func (p *popup) hide() { p.win.Hide() }

type elogViewer struct {
	win *gtk.Window
	da  *gtk.DrawingArea
	eb  *gtk.EventBox
	p   popup
	m   map[uintptr]*decoration
}

const (
	shape_circle = iota
	shape_square
	shape_diamond
	shape_text
)

type rgba struct{ r, g, b, a float64 }

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
	shape int
}

func (lv *elogViewer) decorationForPc(pc uintptr) (d *decoration, ok bool) {
	if lv.m == nil {
		lv.m = make(map[uintptr]*decoration)
	}
	if d, ok = lv.m[pc]; !ok {
		i := len(lv.m)
		i0, i1 := i/len(standardColors), i%len(standardColors)
		d = &decoration{shape: i0, color: standardColors[i1]}
		d.shape = shape_text
		if d.color.a == 0 {
			d.color.a = .2
		}
		lv.m[pc] = d
	}
	return
}

const margin = 40

func (lv *elogViewer) draw(eb *gtk.EventBox, cr *cairo.Context) {
	ev := elog.NewView()
	var tb elog.TimeBounds
	ev.GetTimeBounds(&tb)
	fmt.Printf("foo %+v\n", &tb)

	c := (*ctx)(cr)

	dw := XX(margin)
	w := XY(float64(eb.GetAllocatedWidth()), float64(eb.GetAllocatedHeight()))
	w -= dw

	// Title
	{
		cr.SetSourceRGB(0, 0, 0)
		x := XY(w.X()/2, 20)
		cr.SetFontSize(18)
		c.text(x, tb.Start.Format("2006-01-02 15:04:05"))
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
		c.text2(dw-10i, XY(0, .5), fmt.Sprintf("%.1f%s", t_min/t_unit, tb.UnitName))
		cr.Stroke()
		c.text2(tmp-10i, XY(1, .5), fmt.Sprintf("%.1f%s", t_max/t_unit, tb.UnitName))
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

		s := e.Strings()[0]
		_, pc := ev.EventPath(e)

		d, _ := lv.decorationForPc(pc)
		center := dw + x
		radius := 4.
		dx := XY(radius, radius)
		cr.SetSourceRGBA(d.color.r, d.color.g, d.color.b, d.color.a)
		switch d.shape {
		case shape_circle:
			cr.Arc(center.X(), center.Y(), radius, 0, 2*math.Pi)
			cr.Fill()
		case shape_square:
			c.rect(center-dx, 2*dx)
			cr.Fill()
		case shape_diamond:
			rx, ry := radius*.9, radius*1.2
			cr.MoveTo(center.X()-rx, center.Y())
			cr.LineTo(center.X(), center.Y()-ry)
			cr.LineTo(center.X()+rx, center.Y())
			cr.LineTo(center.X(), center.Y()+ry)
			cr.ClosePath()
			cr.Fill()
		case shape_text:
			cr.Save()
			cr.SetFontSize(9)
			fe := cr.FontExtents()
			te := cr.TextExtents(s)
			iy := uint(0)
			for {
				x_last.Validate(uint(iy))
				if x_left := center.X() - te.Width/2; x_last[iy]+2*radius < x_left {
					x_last[iy] = x_left + te.Width
					break
				}
				iy++
			}
			// Odd goes above baseline; even goes below baseline.
			var idy int
			iy++
			if iy%2 != 0 {
				idy = int((iy + 1) / 2)
			} else {
				idy = -int(iy / 2)
			}
			rr_center := center - XY(0, (fe.Height+2*radius)*float64(idy))
			cr.SetSourceRGB(0, 0, 0)
			tw := c.text(rr_center, s)
			rr_dx := tw + XY(2*radius, radius)
			cr.Stroke()
			cr.SetSourceRGBA(d.color.r, d.color.g, d.color.b, d.color.a)
			c.roundedRect(rr_center-rr_dx/2, rr_dx, radius)
			cr.Fill()

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

		default:
			panic("ga")
		}
	}

	lv.eb.GrabFocus()
}
