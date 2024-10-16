package bank

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccount_UnmarshalText(t *testing.T) {
	var a Account
	err := a.UnmarshalText([]byte(`{"iban" : "DE02120300000000202051", "bic" : "BYLADEM1001", "currency" : null, "holder" : null}`))
	require.NoError(t, err, "Account.UnmarshalText")
}
