package cpu

// Cache lines on x86 are 64 bytes.
const Log2CacheLineBytes = 6

func TimeNow() Time
