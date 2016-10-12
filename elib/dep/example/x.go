package main

import (
	"github.com/platinasystems/go/elib/dep"

	"fmt"
)

type t1Hook func(i int)

//go:generate gentemplate -id t1hook -d Package=main -d DepsType=t1HookVec -d Type=t1Hook -d Data=hooks github.com/platinasystems/go/elib/dep/dep.tmpl

type h0 struct{ x int }
type h1 struct{ x int }

func (h *h0) f(x int) { fmt.Printf("%T %+v\n", h, h) }
func (h *h1) f(x int) { fmt.Printf("%T %+v\n", h, h) }

func main() {
	x0, x1, x2 := &h0{x: 0}, &h1{x: 1}, &h1{x: 2}
	hookVec := &t1HookVec{}
	hs := []dep.Dep{
		dep.Dep{Order: 0},
		dep.Dep{Order: 1},
		dep.Dep{Order: 2},
	}
	if true {
		hs[0].Deps = []*dep.Dep{&hs[1]}
	}
	fs := []t1Hook{x0.f, x1.f, x2.f}
	if false {
		for i := range hs {
			hookVec.Add(fs[i], &hs[i])
		}
	} else {
		for i := range fs {
			hookVec.Add(fs[i])
		}
	}
	for i := range hookVec.hooks {
		hookVec.Get(i)(i)
	}
}
