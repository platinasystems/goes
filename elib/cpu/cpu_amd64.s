// func TimeNow() uint64
TEXT ·TimeNow(SB),$0-0
        RDTSC
	SHLQ $32, DX
	ADDQ DX, AX
        MOVQ AX, ret+0(FP)
        RET

TEXT ·GetCallerPC(SB),4,$8-16
	MOVQ	argp+0(FP),AX		// addr of first arg
	MOVQ	-8(AX),AX		// get calling pc
	MOVQ	AX, ret+8(FP)
	RET
