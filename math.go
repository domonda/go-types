package types

const (
	uvnan    = 0x7FF8000000000001
	uvinf    = 0x7FF0000000000000
	uvneginf = 0xFFF0000000000000
	mask     = 0x7FF
	shift    = 64 - 11 - 1
	bias     = 1023
)

const (
	signMask = 1 << 63
	fracMask = (1 << shift) - 1
	halfMask = 1 << (shift - 1)
	one      = bias << shift
)
