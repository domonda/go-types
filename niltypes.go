package types

import "github.com/domonda/go-types/uu"

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

// TODO remove
// UUIDOrNil returns *uuidPtr if uuidPtr is not nil and not the default zero value,
// else nil is returned.
func UUIDOrNil(uuidPtr *uu.ID) interface{} {
	if uuidPtr == nil || *uuidPtr == uu.IDNil {
		return nil
	}
	return *uuidPtr
}
