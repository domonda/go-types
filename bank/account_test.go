package bank

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestAccount_Valid(t *testing.T) {
	var nilAccount *Account
	assert.False(t, nilAccount.Valid(), "nil Account must not be valid")
	assert.Error(t, nilAccount.Validate())

	valid := &Account{IBAN: "DE89370400440532013000", BIC: "COBADEFFXXX", Currency: "EUR", Holder: "John Doe"}
	assert.True(t, valid.Valid())
	assert.NoError(t, valid.Validate())

	// BIC and currency are optional, but the IBAN is required
	assert.True(t, (&Account{IBAN: "DE89370400440532013000"}).Valid())
	assert.False(t, (&Account{}).Valid())
	assert.Error(t, (&Account{}).Validate())

	tests := []struct {
		name    string
		account *Account
	}{
		{name: "invalid IBAN", account: &Account{IBAN: "DE89370400440532013001"}},
		{name: "invalid BIC", account: &Account{IBAN: "DE89370400440532013000", BIC: "DEUTZZFF"}},
		{name: "invalid currency", account: &Account{IBAN: "DE89370400440532013000", Currency: "XYZ"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.False(t, tt.account.Valid())
			assert.Error(t, tt.account.Validate())
		})
	}

	t.Run("all errors are reported", func(t *testing.T) {
		account := &Account{IBAN: "DE89370400440532013001", BIC: "DEUTZZFF", Currency: "XYZ"}
		err := account.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "IBAN")
		assert.Contains(t, err.Error(), "DEUTZZFF")
		assert.Contains(t, err.Error(), "XYZ")
	})
}

func TestAccount_Normalize(t *testing.T) {
	var nilAccount *Account
	assert.Error(t, nilAccount.Normalize())

	a := &Account{IBAN: "DE89 3704 0044 0532 0130 00", BIC: "COBADEFF", Currency: "eur", Holder: "John Doe"}
	require.NoError(t, a.Normalize())
	assert.Equal(t, &Account{IBAN: "DE89370400440532013000", BIC: "COBADEFFXXX", Currency: "EUR", Holder: "John Doe"}, a)

	// Optional NULL values stay NULL
	a = &Account{IBAN: "DE89 3704 0044 0532 0130 00"}
	require.NoError(t, a.Normalize())
	assert.Equal(t, &Account{IBAN: "DE89370400440532013000"}, a)

	// Invalid components are left unchanged and all errors are reported
	a = &Account{IBAN: "DE89 3704 0044 0532 0130 01", BIC: "DEUTZZFF", Currency: "XYZ"}
	err := a.Normalize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "IBAN")
	assert.Contains(t, err.Error(), "DEUTZZFF")
	assert.Contains(t, err.Error(), "XYZ")
	assert.Equal(t, &Account{IBAN: "DE89 3704 0044 0532 0130 01", BIC: "DEUTZZFF", Currency: "XYZ"}, a)
}

// String is for debugging and shows the fields as stored,
// so a non-normalized field is not shown as if it were normalized.
func TestAccount_String(t *testing.T) {
	a := &Account{IBAN: "DE89 3704 0044 0532 0130 00"}
	assert.Equal(t, "bank.Account{DE89 3704 0044 0532 0130 00}", a.String(), "NULL fields are omitted")

	a = &Account{IBAN: "DE89370400440532013000", BIC: "COBADEFF", Currency: "EUR", Holder: "John Doe"}
	assert.Equal(t, `bank.Account{DE89370400440532013000, BIC: COBADEFF, EUR, "John Doe"}`, a.String())
}

func TestAccount_JSONRoundTrip(t *testing.T) {
	// Account has no MarshalJSON, so the output of the default struct
	// marshalling must be readable by the custom UnmarshalJSON
	for _, a := range []Account{
		{IBAN: "DE89370400440532013000"},
		{IBAN: "DE89370400440532013000", BIC: "COBADEFFXXX", Currency: "EUR", Holder: "John Doe"},
	} {
		j, err := json.Marshal(a)
		require.NoError(t, err)
		var got Account
		require.NoError(t, json.Unmarshal(j, &got), string(j))
		assert.Equal(t, a, got)
	}
}

func TestAccount_UnmarshalJSON_Edges(t *testing.T) {
	t.Run("null does not change the Account", func(t *testing.T) {
		a := Account{IBAN: "DE89370400440532013000", BIC: "COBADEFFXXX"}
		require.NoError(t, a.UnmarshalJSON([]byte(`null`)))
		assert.Equal(t, Account{IBAN: "DE89370400440532013000", BIC: "COBADEFFXXX"}, a)
	})

	t.Run("IBAN string resets the other fields", func(t *testing.T) {
		// An IBAN string describes the whole account,
		// so values from a previous account must not survive
		a := Account{IBAN: "AT611904300234573201", BIC: "COBADEFFXXX", Currency: "EUR", Holder: "John Doe"}
		require.NoError(t, a.UnmarshalJSON([]byte(`"DE89 3704 0044 0532 0130 00"`)))
		assert.Equal(t, Account{IBAN: "DE89370400440532013000"}, a)
	})

	for _, j := range []string{
		``,
		`"`,
		`"DE89370400440532013001"`, // invalid check sum
		`"DE89370400440532013000`,  // not a JSON string
		`""`,
		`{"iban": 1}`,
		`{"iban": "DE89370400440532013000"`,
		`[]`,
		`"\x"`,
	} {
		t.Run(j, func(t *testing.T) {
			a := Account{IBAN: "AT611904300234573201"}
			err := a.UnmarshalJSON([]byte(j))
			assert.Error(t, err)
			assert.Equal(t, Account{IBAN: "AT611904300234573201"}, a, "failed unmarshalling must not modify the Account")
		})
	}
}

func TestAccount_UnmarshalText_IBAN(t *testing.T) {
	// Plain text IBANs with spaces and surrounding whitespace are normalized
	var a Account
	require.NoError(t, a.UnmarshalText([]byte(" DE89 3704 0044 0532 0130 00\n")))
	assert.Equal(t, Account{IBAN: "DE89370400440532013000"}, a)

	err := a.UnmarshalText([]byte("DE89 3704 0044 0532 0130 01"))
	assert.Error(t, err)
}

func TestAccount_Scan(t *testing.T) {
	const obj = `{"iban":"DE89370400440532013000","bic":"COBADEFFXXX"}`
	want := Account{IBAN: "DE89370400440532013000", BIC: "COBADEFFXXX"}

	var a Account
	require.NoError(t, a.Scan([]byte(obj)), "JSON column as bytes")
	assert.Equal(t, want, a)

	a = Account{}
	require.NoError(t, a.Scan(obj), "JSON column as string")
	assert.Equal(t, want, a)

	a = Account{}
	require.NoError(t, a.Scan(`"DE89370400440532013000"`), "JSON string with IBAN")
	assert.Equal(t, Account{IBAN: "DE89370400440532013000"}, a)

	assert.Error(t, a.Scan(nil), "a non nullable Account can't be scanned from NULL")
	assert.Error(t, a.Scan(int64(1)))
}
