// func getPC(argp unsafe.Pointer, PCHashSeed uint64) (Time, PC, PCHash uint64)
TEXT ·getPC(SB),4,$16-24
        RDTSC
	SHLQ	$32, DX
	ADDQ	DX, AX
        MOVQ	AX, ret+16(FP)
	MOVQ	argp+0(FP),AX	// addr of first arg
	MOVQ	-8(AX),AX	// get calling pc
	MOVQ	AX, ret+24(FP)	// return pc
	MOVQ	argp+8(FP),DX	// hash seed
	XORQ	DX, AX
	MOVQ	AX, X0
	AESENC  X0, X0
	MOVQ	X0, ret+32(FP)
	RET
