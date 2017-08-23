package r2

import (
	"math"
)

// R^2 vector V (x,y) and scalar S.
// Vector is complex numbers for easy arithmetic.
type V complex128

func (x V) X() float64    { return real(x) }
func (x V) Y() float64    { return imag(x) }
func XY(x, y float64) V   { return V(complex(x, y)) }
func XX(x float64) V      { return XY(x, x) }
func UV(x, y uint) V      { return XY(float64(x), float64(y)) }
func UU(x uint) V         { return UV(x, x) }
func IJ(x, y int) V       { return XY(float64(x), float64(y)) }
func II(x int) V          { return IJ(x, x) }
func (x V) Conj() V       { return XY(x.X(), -x.Y()) }
func (x V) Abs2() float64 { return x.X()*x.X() + x.Y()*x.Y() }
func (x V) Abs() float64  { return math.Hypot(x.X(), x.Y()) }

func (x V) XY() (float64, float64) { return x.X(), x.Y() }
func (x V) U() uint                { return uint(x.X()) }
func (x V) V() uint                { return uint(x.Y()) }
func (x V) UV() (uint, uint)       { return x.U(), x.V() }
func (x V) I() int                 { return int(x.X()) }
func (x V) J() int                 { return int(x.Y()) }
func (x V) IJ() (int, int)         { return x.I(), x.J() }

type Rect struct {
	Center, Size V
}

func (t *Rect) Ratio(c V, y, r float64) {
	t.Center = c
	t.Size = XY(y/r, y)
}

// Square with side y.
func (r *Rect) Square(c V, y float64) { r.Ratio(c, y, 1) }

// Golden rectangle with side y.  y/x = (x + y)/y = math.Phi
func (r *Rect) Golden(c V, a float64) { r.Ratio(c, a, math.Phi) }

func (r *Rect) IsInside(t V) bool {
	c, s := r.Center, r.Size/2
	ll, ur := c-s, c+s // lower left, upper right
	if t.X() < ll.X() || t.X() > ur.X() {
		return false
	}
	if t.Y() < ll.Y() || t.Y() > ur.Y() {
		return false
	}
	return true
}
