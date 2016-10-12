package elib

import (
	"testing"
)

func TestHeap(t *testing.T) {
	c := testHeap{
		iterations: 1000,
		nObjects:   10,
		verbose:    0,
	}
	err := runHeapTest(&c)
	if err != nil {
		t.Error(err)
	}
}
