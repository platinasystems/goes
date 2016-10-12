package elib

import (
	"fmt"
	"strconv"
)

// Count implements the Value interface so flags can be specified as
// either integer (1000000) or sometimes more conveniently as floating point (1e6).
type Count int

func (t *Count) Set(s string) (err error) {
	v, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		v, e := strconv.ParseFloat(s, 64)
		*t = Count(v)
		if e == nil {
			err = nil
			return
		}
	}
	*t = Count(v)
	return
}

func (t *Count) String() string { return fmt.Sprintf("%v", *t) }
