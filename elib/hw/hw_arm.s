// func LoadUint32(addr uintptr) (data uint32)
TEXT ·LoadUint32(SB),4,$0-8
	MOVW	addr+0(FP), R1
	MOVW	0(R1), R0
	MOVW	R0, data+4(FP)
	RET
	
// func StoreUint32(addr uintptr, data uint32)
TEXT ·StoreUint32(SB),4,$0-8
	MOVW	addr+0(FP), R1
	MOVW	data+4(FP), R0
	MOVW	R0,0(R1)
	RET

// linux kernel helper: void __kuser_memory_barrier(void) at 0xffff0fa0
TEXT memoryBarrier<>(SB),4,$0
	MOVW	$0xffff0fa0, R15 // R15 is hardware PC.
	
// func MemoryBarrier()
TEXT ·MemoryBarrier(SB),4,$0-0
	BL	memoryBarrier<>(SB)
	RET
