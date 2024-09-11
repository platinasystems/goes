package netpdu

import "testing"

// http://www.faqs.org/rfcs/rfc1071.html
// 3. Numerical Examples
func TestChecksum(t *testing.T) {
	for i, x := range []struct {
		sum   uint32
		final uint16
		data  [][]byte
	}{
		{0x2ddf0, 0xddf2, [][]byte{
			[]byte{0, 1, 0xf2, 3, 0xf4, 0xf5, 0xf6, 0xf7},
		}},
		{0x2ddf0, 0xddf2, [][]byte{
			[]byte{0, 1, 0xf2, 3},
			[]byte{0xf4, 0xf5, 0xf6, 0xf7},
		}},
		{0xf201, 0xf201, [][]byte{
			[]byte{0, 1, 0xf2},
		}},
		{0x1f0ea, 0xf0eb, [][]byte{
			[]byte{3, 0xf4, 0xf5, 0xf6, 0xf7},
		}},
	} {
		sum := uint32(0)
		for _, data := range x.data {
			sum = Checksum(sum, data)
		}
		if sum != x.sum {
			t.Errorf("%d: %#x != %#x", i, sum, x.sum)
		}
		if final := CarryOver(sum); final != x.final {
			t.Errorf("%d: %#x != %#x", i, final, x.final)
		}
	}
}
