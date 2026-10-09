package bank

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNullableBIC_Null(t *testing.T) {
	// NULL is a valid value for a NullableBIC and must pass validation
	// and normalization unchanged without errors
	var bic NullableBIC
	assert.True(t, bic.IsNull(), "zero value must be NULL")
	assert.False(t, bic.IsNotNull())
	assert.True(t, bic.Valid())
	assert.NoError(t, bic.Validate())
	assert.False(t, bic.ValidAndNotNull())
	assert.True(t, bic.ValidAndNormalized())

	norm, err := bic.Normalized()
	require.NoError(t, err)
	assert.Equal(t, BICNull, norm)
	assert.Equal(t, BICNull, bic.NormalizedOrNull())

	assert.Equal(t, "", bic.String())
	assert.Equal(t, "default", bic.StringOr("default"))
	assert.Equal(t, BIC("BELADEBEXXX"), bic.GetOr("BELADEBEXXX"))
	assert.Panics(t, func() { bic.Get() }, "Get on NULL must panic instead of returning an invalid BIC")
}

func TestNullableBIC_Valid(t *testing.T) {
	const shortBIC NullableBIC = "BELADEBE"
	const normBIC NullableBIC = "BELADEBEXXX"

	assert.True(t, shortBIC.Valid())
	assert.True(t, shortBIC.ValidAndNotNull())
	assert.NoError(t, shortBIC.Validate())
	assert.False(t, shortBIC.ValidAndNormalized(), "8 character BIC is normalized to 11 characters")
	assert.True(t, normBIC.ValidAndNormalized())

	norm, err := shortBIC.Normalized()
	require.NoError(t, err)
	assert.Equal(t, normBIC, norm)
	assert.Equal(t, normBIC, shortBIC.NormalizedOrNull())
	assert.Equal(t, string(normBIC), shortBIC.String())
	assert.Equal(t, string(normBIC), normBIC.StringOr("default"))
}

func TestNullableBIC_Invalid(t *testing.T) {
	// Only the empty string is NULL, everything else must be a valid BIC
	for _, bic := range []NullableBIC{" ", "NULL", "DEUTZZFF", "AMTSGERICHT"} {
		t.Run(string(bic), func(t *testing.T) {
			assert.False(t, bic.IsNull())
			assert.False(t, bic.Valid())
			assert.False(t, bic.ValidAndNotNull())
			assert.Error(t, bic.Validate())
			assert.False(t, bic.ValidAndNormalized())

			norm, err := bic.Normalized()
			assert.Error(t, err)
			assert.Equal(t, bic, norm, "invalid BIC must be returned unchanged")
			assert.Equal(t, BICNull, bic.NormalizedOrNull())
			assert.Equal(t, string(bic), bic.String())
		})
	}
}

func TestNullableBIC_SetGet(t *testing.T) {
	var bic NullableBIC
	bic.Set("BELADEBEXXX")
	assert.True(t, bic.IsNotNull())
	assert.Equal(t, BIC("BELADEBEXXX"), bic.Get())
	assert.Equal(t, BIC("BELADEBEXXX"), bic.GetOr("GENODEF1JEV"))

	bic.SetNull()
	assert.True(t, bic.IsNull())

	// Conversions between BIC and NullableBIC
	assert.Equal(t, NullableBIC("BELADEBEXXX"), BIC("BELADEBEXXX").Nullable())
	assert.True(t, BIC("").Nullable().IsNull())
}

func TestNullableBIC_ScanString(t *testing.T) {
	// Textual NULL representations from CSV or form input must become NULL
	for _, source := range []string{"", "NULL", "null", "nil"} {
		t.Run(source, func(t *testing.T) {
			bic := NullableBIC("BELADEBEXXX")
			require.NoError(t, bic.ScanString(source, true))
			assert.True(t, bic.IsNull())
		})
	}

	t.Run("validate accepts valid BIC", func(t *testing.T) {
		var bic NullableBIC
		require.NoError(t, bic.ScanString("BELADEBEXXX", true))
		assert.Equal(t, NullableBIC("BELADEBEXXX"), bic)
	})
	t.Run("validate rejects invalid BIC", func(t *testing.T) {
		bic := NullableBIC("unchanged")
		require.Error(t, bic.ScanString("DEUTZZFF", true))
		assert.Equal(t, NullableBIC("unchanged"), bic)
	})
	t.Run("without validate assigns invalid BIC as is", func(t *testing.T) {
		var bic NullableBIC
		require.NoError(t, bic.ScanString("DEUTZZFF", false))
		assert.Equal(t, NullableBIC("DEUTZZFF"), bic)
	})
}

func TestNullableBIC_SQL(t *testing.T) {
	t.Run("NULL is written as nil", func(t *testing.T) {
		// NULL must be written to SQL as nil, not as an empty string
		v, err := BICNull.Value()
		require.NoError(t, err)
		assert.Nil(t, v)
	})
	t.Run("value is written as string", func(t *testing.T) {
		v, err := NullableBIC("BELADEBEXXX").Value()
		require.NoError(t, err)
		assert.Equal(t, "BELADEBEXXX", v)
	})

	tests := []struct {
		name    string
		value   any
		want    NullableBIC
		wantErr bool
	}{
		{name: "SQL NULL", value: nil, want: BICNull},
		{name: "string", value: "BELADEBEXXX", want: "BELADEBEXXX"},
		{name: "bytes", value: []byte("BELADEBEXXX"), want: "BELADEBEXXX"},
		{name: "unsupported type", value: int64(1), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bic := NullableBIC("unchanged")
			err := bic.Scan(tt.value)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, NullableBIC("unchanged"), bic)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, bic)
		})
	}

	t.Run("Scan copies bytes", func(t *testing.T) {
		// database/sql may reuse the []byte buffer after Scan returns
		buf := []byte("BELADEBEXXX")
		var bic NullableBIC
		require.NoError(t, bic.Scan(buf))
		copy(buf, "XXXXXXXXXXX")
		assert.Equal(t, NullableBIC("BELADEBEXXX"), bic)
	})
}

func TestNullableBIC_JSON(t *testing.T) {
	type obj struct {
		BIC NullableBIC `json:"bic"`
	}

	j, err := json.Marshal(obj{BIC: BICNull})
	require.NoError(t, err)
	assert.Equal(t, `{"bic":null}`, string(j), "NULL must be marshalled as JSON null, not as empty string")
	var got obj
	require.NoError(t, json.Unmarshal(j, &got))
	assert.True(t, got.BIC.IsNull())

	j, err = json.Marshal(obj{BIC: "BELADEBEXXX"})
	require.NoError(t, err)
	assert.Equal(t, `{"bic":"BELADEBEXXX"}`, string(j))
	require.NoError(t, json.Unmarshal(j, &got))
	assert.Equal(t, NullableBIC("BELADEBEXXX"), got.BIC)
}

func TestNullableBIC_JSONSchema(t *testing.T) {
	// The schema must accept JSON null because NULL is marshalled as null
	s := BICNull.JSONSchema()
	require.Len(t, s.OneOf, 2)
	assert.Equal(t, "string", s.OneOf[0].Type)
	assert.Equal(t, BICRegex, s.OneOf[0].Pattern)
	assert.Equal(t, "null", s.OneOf[1].Type)
}
