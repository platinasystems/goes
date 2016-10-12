// func LoadUint32(addr *uint32) (data uint32)
TEXT ·LoadUint32(SB),4,$0-12
	MOVQ	addr+0(FP), AX
	MOVL	0(AX), AX
	MOVL	AX, data+8(FP)
	RET
	
// func StoreUint32(addr *uint32, data uint32)
TEXT ·StoreUint32(SB),4,$0-12
	MOVQ	addr+0(FP), AX
	MOVL	val+8(FP), BX
	MOVL	BX, 0(AX)
	RET

// func LoadUint64(addr *uint64) (data uint64)
TEXT ·LoadUint64(SB),4,$0-12
	MOVQ	addr+0(FP), AX
	MOVQ	0(AX), AX
	MOVQ	AX, data+8(FP)
	RET
	
// func StoreUint64(addr *uint64, data uint64)
TEXT ·StoreUint64(SB),4,$0-12
	MOVQ	addr+0(FP), AX
	MOVQ	val+8(FP), BX
	MOVQ	BX, 0(AX)
	RET

// func MemoryBarrier()
TEXT ·MemoryBarrier(SB),4,$0-0
	MFENCE
	RET
