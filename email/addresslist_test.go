package email

import (
	"encoding/json"
	"net/mail"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStandardComparison(t *testing.T) {
	validEmailAddresses := map[string]*mail.Address{
		`"Unger, Erik" <u.erik@domonda.com>`:        {Name: "Unger, Erik", Address: "u.erik@domonda.com"},
		`"Unger, Erik" <"Unger, Erik"@domonda.com>`: {Name: "Unger, Erik", Address: "Unger, Erik@domonda.com"},
	}

	for addr, expected := range validEmailAddresses {
		t.Run(addr, func(t *testing.T) {
			result, err := mail.ParseAddress(addr)
			assert.NoError(t, err, "valid email address")
			assert.Equal(t, expected, result, "expected: %s", expected)
		})
	}

	for addr, expected := range validEmailAddresses {
		t.Run(addr, func(t *testing.T) {
			results, err := mail.ParseAddressList(addr)
			assert.NoError(t, err, "valid email address")
			assert.Len(t, results, 1, "list of one address")
			assert.Equal(t, expected, results[0], "expected: %s", expected)
		})
	}
}

func TestAddressList_Split(t *testing.T) {
	tests := []struct {
		l       AddressList
		want    []Address
		wantErr bool
	}{
		{l: ``, want: nil},
		{l: `<hello@example.com>,`, want: []Address{`hello@example.com`}},
		{l: `<Hello@example.com>, World@example.com`, want: []Address{`hello@example.com`, `world@example.com`}},
		// Partial result: parsable addresses are returned alongside the error.
		{l: `alice@example.com, broken address, bob@example.com`, want: []Address{`alice@example.com`, `bob@example.com`}, wantErr: true},
		{l: `@broken`, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(string(tt.l), func(t *testing.T) {
			got, err := tt.l.Split()
			if (err != nil) != tt.wantErr {
				t.Errorf("AddressList(%#v).Split() error = %v, wantErr %v", string(tt.l), err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddressList(%#v).Split() = %v, want %v", string(tt.l), got, tt.want)
			}
		})
	}
}

func TestAddressList_Scan(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    AddressList
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
		// AddressList is the not-null type, so SQL NULL is an error here.
		// NullableAddressList is what scans NULL into the empty list.
		{name: "null", value: nil, wantErr: true},
		{name: "empty string", value: "", wantErr: true},
		{name: "invalid array", value: `{,}`, wantErr: true},
		{name: "unsupported type", value: 123, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var l AddressList
			err := l.Scan(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Scan() error = %v, wantErr %t", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if l != tt.want {
				t.Errorf("Scan() = %q, want %q", l, tt.want)
			}
		})
	}
}

// TestNormalizeAddressList protects deduplication by mailbox: the same
// address with a different display name or case must only appear once
// (keeping the first spelling), otherwise a recipient would get a
// message multiple times.
func TestNormalizeAddressList(t *testing.T) {
	got, err := NormalizeAddressList(`John <John@Example.com>, jane@example.com, "J. Doe" <john@example.com>, JANE@example.com`)
	require.NoError(t, err)
	assert.Equal(t, []string{`"John" <john@example.com>`, `jane@example.com`}, got)

	// A strict normalization must not silently drop invalid entries.
	_, err = NormalizeAddressList(`a@example.com, not an address`)
	assert.Error(t, err)
}

// TestAddressListJoin protects that empty addresses never produce
// empty list entries like "a@example.com, , b@example.com" that
// would make the joined list invalid.
func TestAddressListJoin(t *testing.T) {
	tests := []struct {
		name  string
		addrs []Address
		want  AddressList
	}{
		{name: "none", addrs: nil, want: ""},
		{name: "only empty", addrs: []Address{"", ""}, want: ""},
		{name: "one", addrs: []Address{"a@example.com"}, want: "a@example.com"},
		{name: "empty skipped", addrs: []Address{"", "a@example.com", "", `"Doe, John" <john@example.com>`, ""}, want: `a@example.com, "Doe, John" <john@example.com>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, AddressListJoin(tt.addrs...), "AddressListJoin")

			strs := make([]string, len(tt.addrs))
			for i, a := range tt.addrs {
				strs[i] = string(a)
			}
			assert.Equal(t, tt.want, AddressListJoinStrings(strs...), "AddressListJoinStrings must behave like AddressListJoin")
		})
	}
}

func TestAddressList_Append(t *testing.T) {
	tests := []struct {
		name  string
		list  AddressList
		addrs []Address
		want  AddressList
	}{
		// Appending to an empty list must not start with a separator.
		{name: "to empty", list: "", addrs: []Address{"a@example.com"}, want: "a@example.com"},
		{name: "to non empty", list: "a@example.com", addrs: []Address{"b@example.com", "c@example.com"}, want: "a@example.com, b@example.com, c@example.com"},
		{name: "empty skipped", list: "a@example.com", addrs: []Address{"", "b@example.com", ""}, want: "a@example.com, b@example.com"},
		{name: "nothing", list: "a@example.com", addrs: nil, want: "a@example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.list.Append(tt.addrs...)
			assert.Equal(t, tt.want, got)
			assert.True(t, got.Valid(), "the appended list must stay parsable")
		})
	}
}

// TestAddressList_UniqueAddressParts protects that the set contains bare,
// lower case mailboxes so it can be used for recipient lookups regardless
// of display names and case.
func TestAddressList_UniqueAddressParts(t *testing.T) {
	set, err := AddressList(`John <John@Example.com>, jane@example.com, "J. Doe" <john@example.com>`).UniqueAddressParts()
	require.NoError(t, err)
	assert.Equal(t, MakeAddressSet("john@example.com", "jane@example.com"), set)

	set, err = AddressList(`a@example.com, not an address`).UniqueAddressParts()
	assert.Error(t, err, "an invalid entry must not be silently dropped")
	assert.Nil(t, set)
}

func TestAddressList_ValidateAndNormalized(t *testing.T) {
	tests := []struct {
		list               AddressList
		valid              bool
		normalized         AddressList
		validAndNormalized bool
	}{
		{
			list:               `a@example.com, b@example.com`,
			valid:              true,
			normalized:         `a@example.com, b@example.com`,
			validAndNormalized: true,
		},
		{
			list:       `A@Example.com,b@example.com`,
			valid:      true,
			normalized: `a@example.com, b@example.com`,
		},
		{
			list:       `John <John@Example.com>, <jane@example.com>`,
			valid:      true,
			normalized: `"John" <john@example.com>, jane@example.com`,
		},
		// The quoted comma in the name must survive normalization,
		// otherwise the normalized list would split into a different
		// number of addresses.
		{
			list:               `"Doe, John" <john@example.com>, b@example.com`,
			valid:              true,
			normalized:         `"Doe, John" <john@example.com>, b@example.com`,
			validAndNormalized: true,
		},
		{list: `a@example.com, not an address`, valid: false},
		{list: `not an address`, valid: false},
	}
	for _, tt := range tests {
		t.Run(string(tt.list), func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.list.Valid(), "Valid")
			assert.Equal(t, tt.valid, tt.list.Validate() == nil, "Validate must agree with Valid")
			assert.Equal(t, tt.validAndNormalized, tt.list.ValidAndNormalized(), "ValidAndNormalized")

			normalized, err := tt.list.Normalized()
			if !tt.valid {
				assert.Error(t, err)
				// On error the input is returned unchanged so it isn't lost.
				assert.Equal(t, tt.list, normalized)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.normalized, normalized)
			assert.True(t, normalized.ValidAndNormalized(), "normalizing must be idempotent")
		})
	}
}

// Value and MarshalJSON write an AddressList normalized like Address,
// see TestAddress_ValueAndMarshalJSONNormalize. A list that can't be
// normalized is written unchanged instead of failing the write.
func TestAddressList_ValueAndMarshalJSONNormalize(t *testing.T) {
	tests := []struct {
		l    AddressList
		want string
	}{
		{l: `a@example.com`, want: `a@example.com`},
		{l: `A@Example.com,  Jürgen <J@Example.com>`, want: `a@example.com, "Jürgen" <j@example.com>`},
		{l: `a@example.com, not an address`, want: `a@example.com, not an address`},
		{l: ``, want: ``},
	}
	for _, tt := range tests {
		v, err := tt.l.Value()
		require.NoError(t, err)
		assert.Equal(t, tt.want, v, "AddressList(%q).Value()", string(tt.l))

		j, err := json.Marshal(tt.l)
		require.NoError(t, err)
		var got string
		require.NoError(t, json.Unmarshal(j, &got))
		assert.Equal(t, tt.want, got, "json.Marshal(AddressList(%q))", string(tt.l))
	}
}

// String and ASCII of a list work like those of Address,
// see TestAddress_StringAndASCII.
func TestAddressList_StringAndASCII(t *testing.T) {
	l := AddressList(`A@Example.com, Jürgen <J@Example.com>`)
	assert.Equal(t, `a@example.com, "Jürgen" <j@example.com>`, l.String())
	assert.Equal(t, `a@example.com, =?utf-8?q?J=C3=BCrgen?= <j@example.com>`, l.ASCII())

	invalid := AddressList(`a@example.com, not an address`)
	assert.Equal(t, string(invalid), invalid.String())
	assert.Equal(t, string(invalid), invalid.ASCII())
}

func TestAddressList_Nullable(t *testing.T) {
	const list AddressList = `"Doe, John" <john@example.com>, b@example.com`
	assert.Equal(t, NullableAddressList(list), list.Nullable(), "conversion must keep the exact string")
}
