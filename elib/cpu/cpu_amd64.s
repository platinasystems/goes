// func TimeNow() uint64
TEXT ·TimeNow(SB),$0-0
        RDTSC
	SHLQ $32, DX
	ADDQ DX, AX
        MOVQ AX, ret+0(FP)
        RET
