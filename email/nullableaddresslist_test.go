package email

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNullableAddressList_Scan(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    NullableAddressList
		wantErr bool
	}{
		// The array branch reads a PostgreSQL text[] column. Its elements
		// are quoted and escaped, so they have to be decoded rather than
		// joined verbatim, otherwise the quotes end up inside an address.
		{
			name:  "quoted array elements",
			value: `{"a@example.com","b@example.com"}`,
			want:  "a@example.com, b@example.com",
		},
		{
			name:  "quoted element with a comma and escaped quotes",
			value: `{"\"Doe, John\" <john@example.com>","b@example.com"}`,
			want:  `"Doe, John" <john@example.com>, b@example.com`,
		},
		{
			name:  "unquoted array elements",
			value: `{a@example.com,b@example.com}`,
			want:  "a@example.com, b@example.com",
		},
		{name: "empty array", value: `{}`, want: ""},
		// A hand written literal, unlike PostgreSQL output, may pad its
		// elements. The padding is not part of the address.
		{name: "padded elements", value: `{ a@example.com , b@example.com }`, want: "a@example.com, b@example.com"},
		{name: "whitespace only array", value: `{ }`, want: ""},
		// Regression cases for the array codec swap (nullable.SplitArray ->
		// notnull.StringArray). Each behaves differently than before, so each
		// is pinned rather than left to be discovered in production.
		// A SQL NULL element used to become the address literally spelled NULL.
		{name: "NULL element", value: `{NULL}`, wantErr: true},
		// Quoted, it is the ordinary string "NULL", not a SQL NULL.
		{name: "quoted NULL element", value: `{"NULL"}`, want: "NULL"},
		// A trailing comma used to be accepted and yield one element.
		{name: "trailing comma", value: `{a@example.com,}`, wantErr: true},
		// Backslash escapes are unescaped now; before, both backslashes stayed.
		{name: "escaped backslash", value: `{"a\\b@example.com"}`, want: `a\b@example.com`},
		// A nested literal used to yield the inner braces as a single address.
		{name: "nested array", value: `{{a@example.com}}`, wantErr: true},
		// A non-array string is an already joined list, so it is taken as is.
		{name: "plain string", value: "a@example.com, b@example.com", want: "a@example.com, b@example.com"},
		// Like the padded array elements above and like AddressSet.Scan,
		// the padding of a plain string is not part of the list, else
		// " a@example.com " and "a@example.com" would be different lists
		// and the padding would end up in the first and last address.
		{name: "padded plain string", value: " a@example.com, b@example.com\t\n", want: "a@example.com, b@example.com"},
		{name: "padded plain bytes", value: []byte(" a@example.com "), want: "a@example.com"},
		// Whitespace only is no list, like the empty string
		{name: "whitespace only plain string", value: " \t ", wantErr: true},
		{name: "bytes", value: []byte(`{"a@example.com"}`), want: "a@example.com"},
		// SQL NULL is the empty list, unlike the empty string which is
		// rejected because it can only come from a non-NULL column.
		{name: "null", value: nil, want: ""},
		{name: "empty string", value: "", wantErr: true},
		{name: "invalid array", value: `{,}`, wantErr: true},
		{name: "unsupported type", value: 123, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var n NullableAddressList
			err := n.Scan(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Scan() error = %v, wantErr %t", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if n != tt.want {
				t.Errorf("Scan() = %q, want %q", n, tt.want)
			}
		})
	}
}

func TestNullableAddressList_ScanValueRoundTrip(t *testing.T) {
	// Value always writes the plain joined string, never an array, so a
	// value written by this type has to scan back unchanged.
	lists := []NullableAddressList{
		"",
		"a@example.com",
		"a@example.com, b@example.com",
		`"Doe, John" <john@example.com>, b@example.com`,
	}
	for _, want := range lists {
		t.Run(string(want), func(t *testing.T) {
			value, err := want.Value()
			if err != nil {
				t.Fatalf("Value returned %v", err)
			}
			var got NullableAddressList
			if err := got.Scan(value); err != nil {
				t.Fatalf("Scan(%v) returned %v", value, err)
			}
			if got != want {
				t.Errorf("Scan(Value()) = %q, want %q", got, want)
			}
		})
	}
}

// TestNullableAddressList_Null protects that the NULL list (empty string)
// is a valid, already normalized empty list for which parsing yields no
// addresses and no error, so optional headers like Cc can be handled
// without special casing.
func TestNullableAddressList_Null(t *testing.T) {
	var n NullableAddressList

	assert.True(t, n.IsNull())
	assert.False(t, n.IsNotNull())
	assert.NoError(t, n.Validate())
	assert.True(t, n.ValidAndNormalized())

	parsed, err := n.Parse()
	assert.NoError(t, err)
	assert.Nil(t, parsed)

	split, err := n.Split()
	assert.NoError(t, err)
	assert.Nil(t, split)

	normalized, err := n.Normalized()
	assert.NoError(t, err)
	assert.True(t, normalized.IsNull())

	assert.Equal(t, "<none>", n.StringOr("<none>"))
	assert.Equal(t, AddressList("default@example.com"), n.GetOr("default@example.com"))
	assert.Panics(t, func() { n.Get() }, "Get on NULL must panic instead of returning an invalid AddressList")
}

func TestNullableAddressList_NotNull(t *testing.T) {
	const n NullableAddressList = `"Doe, John" <John@Example.com>, B@example.com`

	assert.True(t, n.IsNotNull())
	assert.NoError(t, n.Validate())
	assert.False(t, n.ValidAndNormalized(), "upper case address parts are not normalized")

	parsed, err := n.Parse()
	require.NoError(t, err)
	require.Len(t, parsed, 2, "the quoted comma must not split the first address")
	assert.Equal(t, "Doe, John", parsed[0].Name)
	assert.Equal(t, "john@example.com", parsed[0].Address)
	assert.Equal(t, "b@example.com", parsed[1].Address)

	normalized, err := n.Normalized()
	require.NoError(t, err)
	assert.Equal(t, NullableAddressList(`"Doe, John" <john@example.com>, b@example.com`), normalized)
	assert.True(t, normalized.ValidAndNormalized())

	assert.Equal(t, string(n), n.StringOr("<none>"))
	assert.Equal(t, AddressList(n), n.Get())
	assert.Equal(t, AddressList(n), n.GetOr("default@example.com"))
}

func TestNullableAddressList_Invalid(t *testing.T) {
	// A non-NULL invalid list must not be mistaken for NULL.
	const n NullableAddressList = "a@example.com, not an address"

	assert.Error(t, n.Validate())
	assert.False(t, n.ValidAndNormalized())

	normalized, err := n.Normalized()
	assert.Error(t, err)
	assert.Equal(t, n, normalized, "Normalized must return the input unchanged on error")
}

// TestNullableAddressListJoin protects that joining no usable addresses
// results in the NULL list instead of an invalid list of empty entries.
func TestNullableAddressListJoin(t *testing.T) {
	assert.True(t, NullableAddressListJoin().IsNull())
	assert.True(t, NullableAddressListJoin("", "").IsNull())
	assert.Equal(t,
		NullableAddressList(`a@example.com, "Doe, John" <john@example.com>`),
		NullableAddressListJoin("", "a@example.com", "", `"Doe, John" <john@example.com>`),
	)
}

func TestNullableAddressList_Append(t *testing.T) {
	var n NullableAddressList
	n = n.Append("")
	assert.True(t, n.IsNull(), "appending only empty addresses keeps the list NULL")

	n = n.Append("a@example.com")
	assert.Equal(t, NullableAddressList("a@example.com"), n, "appending to NULL must not start with a separator")

	n = n.Append("", "b@example.com")
	assert.Equal(t, NullableAddressList("a@example.com, b@example.com"), n)
}

func TestNullableAddressList_SetAndSetNull(t *testing.T) {
	var n NullableAddressList
	n.Set("a@example.com, b@example.com")
	assert.Equal(t, NullableAddressList("a@example.com, b@example.com"), n)

	n.SetNull()
	assert.True(t, n.IsNull())

	n.Set("a@example.com")
	n.Set("") // documented: setting the empty AddressList means NULL
	assert.True(t, n.IsNull())
}

// TestNullableAddressList_JSON protects the JSON NULL mapping: the NULL list
// is written as JSON null and both null and "" are read as NULL.
func TestNullableAddressList_JSON(t *testing.T) {
	data, err := json.Marshal(NullableAddressList(""))
	require.NoError(t, err)
	assert.Equal(t, `null`, string(data))

	tests := []struct {
		json    string
		want    NullableAddressList
		wantErr bool
	}{
		{json: `null`, want: ""},
		{json: `""`, want: ""},
		{json: `"a@example.com, b@example.com"`, want: "a@example.com, b@example.com"},
		{json: `["a@example.com"]`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.json, func(t *testing.T) {
			n := NullableAddressList("previous@example.com")
			err := json.Unmarshal([]byte(tt.json), &n)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, n)
		})
	}

	t.Run("round-trip", func(t *testing.T) {
		for _, want := range []NullableAddressList{"", `"Doe, John" <john@example.com>, b@example.com`} {
			data, err := json.Marshal(want)
			require.NoError(t, err)
			var got NullableAddressList
			require.NoError(t, json.Unmarshal(data, &got))
			assert.Equal(t, want, got)
		}
	})
}

func TestNullableAddressList_UnmarshalText(t *testing.T) {
	n := NullableAddressList("previous@example.com")
	require.NoError(t, n.UnmarshalText(nil))
	assert.True(t, n.IsNull(), "empty text must be NULL")

	require.NoError(t, n.UnmarshalText([]byte("a@example.com, b@example.com")))
	assert.Equal(t, NullableAddressList("a@example.com, b@example.com"), n)
}
