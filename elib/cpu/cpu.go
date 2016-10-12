// +build !amd64

package cpu

import (
	"time"
)

// Cache lines on generic.
const Log2CacheLineBytes = 6

func TimeNow() Time {
	return Time(time.Now().UnixNano())
}
