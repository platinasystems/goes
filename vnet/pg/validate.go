package pg

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/vnet"

	"fmt"
)

const validate = elib.Debug && true

func (t *buffer_type) validate_ref(r *vnet.Ref) {
	if !validate {
		return
	}
	if got, want := r.DataLen(), uint(len(t.data)); got != want {
		fmt.Printf("%s\n", r.String())
		panic(fmt.Errorf("generate wrong len got %d != want %d", got, want))
	}
	if got, want := string(r.DataSlice()), string(t.data); got != want {
		fmt.Printf("%s\n", r.String())
		panic(fmt.Errorf("generate wrong data got %x != want %x", got, want))
	}
	t.validate_sequence++
}

func (n *node) validate_ref(r *vnet.Ref, s *Stream) {
	if !validate {
		return
	}
	if got, want := r.ChainLen(), s.cur_size; got != want {
		fmt.Printf("%s\n", r.String())
		panic(fmt.Errorf("generate wrong size got %d != want %d", got, want))
	}
	if got, want := r.ChainSlice(n.validate_data), s.data[:s.cur_size]; string(got) != string(want) {
		fmt.Printf("%s\n", r.String())
		panic(fmt.Errorf("generate wrong data got %x != want %x", got, want))
	} else {
		n.validate_data = got // re-use next time.
	}
	n.validate_sequence++
}
