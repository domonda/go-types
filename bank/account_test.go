package bank

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccount_UnmarshalText(t *testing.T) {
	var a Account
	err := a.UnmarshalText([]byte(`{"iban" : "DE02120300000000202051", "bic" : "BYLADEM1001", "currency" : null, "holder" : null}`))
	require.NoError(t, err, "Account.UnmarshalText")
	require.Equal(t, Account{IBAN: "DE02120300000000202051", BIC: "BYLADEM1001"}, a)

	err = a.UnmarshalText([]byte(`DE02120300000000202051`))
	require.NoError(t, err, "Account.UnmarshalText")
	require.Equal(t, Account{IBAN: "DE02120300000000202051"}, a)
}

func TestAccount_UnmarshalJSON(t *testing.T) {
	var a Account
	err := json.Unmarshal([]byte(`{"iban" : "DE02120300000000202051", "bic" : "BYLADEM1001", "currency" : null, "holder" : null}`), &a)
	require.NoError(t, err, "Account.UnmarshalJSON")
	require.Equal(t, Account{IBAN: "DE02120300000000202051", BIC: "BYLADEM1001"}, a)

	err = a.UnmarshalText([]byte(`"DE02120300000000202051"`))
	require.NoError(t, err, "Account.UnmarshalJSON")
	require.Equal(t, Account{IBAN: "DE02120300000000202051"}, a)

	var as []Account
	err = json.Unmarshal([]byte(`[{"iban" : "DE02120300000000202051", "bic" : "BYLADEM1001", "currency" : null, "holder" : null}]`), &as)
	require.NoError(t, err, "json.Unmarshal")
	require.Equal(t, []Account{{IBAN: "DE02120300000000202051", BIC: "BYLADEM1001"}}, as)
}
