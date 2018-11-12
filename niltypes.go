package types

func Int64OrZeroFromPtr(ptr *int64) int64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}

func IntOrZeroFromPtr(ptr *int) int {
	if ptr == nil {
		return 0
	}
	return *ptr
}

func Float64OrZeroFromPtr(ptr *float64) float64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}
