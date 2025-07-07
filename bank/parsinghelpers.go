package bank

func isInRange(b byte, start, end byte) bool {
	return b >= start && b <= end
}

func isUpperAZ(b byte) bool {
	return isInRange(b, 'A', 'Z')
}

func isNum(b byte) bool {
	return isInRange(b, '0', '9')
}

func isUpperAZ0to9(b byte) bool {
	return isUpperAZ(b) || isNum(b)
}
