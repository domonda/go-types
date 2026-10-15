package email

import (
	"encoding"
	"encoding/json"
	"net/mail"
	"reflect"
	"strings"
	"testing"
	"unicode"

	"github.com/domonda/go-types/internal/normtest"
	"github.com/domonda/go-types/strfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var emailAddressNormalized = map[Address]string{
	"test@test.com":                  "test@test.com",
	"te+st@test.com":                 "te+st@test.com",
	"test888@test.com":               "test888@test.com",
	"<noname@test.com>":              "noname@test.com",
	"With Name <With.Name@test.com>": `"With Name" <with.name@test.com>`,
	// Non-ASCII names stay UTF-8, also when they were RFC 2047 encoded
	"Jürgen Müller <J@Example.com>":                     `"Jürgen Müller" <j@example.com>`,
	"=?utf-8?b?wqFIb2xhLCBzZcOxb3Ih?= <senor@hola.com>": `"¡Hola, señor!" <senor@hola.com>`,
}

func TestAddressNormalized(t *testing.T) {
	for addr, expected := range emailAddressNormalized {
		t.Run(string(addr), func(t *testing.T) {
			result, err := addr.Normalized()
			assert.NoError(t, err)
			assert.Equal(t, expected, string(result))
		})
	}
}

func TestAddressFrom(t *testing.T) {
	tests := []struct {
		name string
		addr *mail.Address
		want Address
	}{
		{name: `nil`, addr: nil, want: ``},
		{name: `empty`, addr: &mail.Address{Name: "", Address: ""}, want: ``},
		{name: `erik@domonda.com`, addr: &mail.Address{Name: "", Address: "erik@domonda.com"}, want: `erik@domonda.com`},
		{name: `"Erik Unger" <erik@domonda.com>`, addr: &mail.Address{Name: "Erik Unger", Address: "erik@domonda.com"}, want: `"Erik Unger" <erik@domonda.com>`},
		// Unlike mail.Address.String, a non-ASCII name is not RFC 2047
		// encoded, so the Address stays readable in JSON and databases.
		{name: `non-ASCII name`, addr: &mail.Address{Name: "Jürgen Müller", Address: "j@example.com"}, want: `"Jürgen Müller" <j@example.com>`},
		// Quotes and backslashes are escaped as quoted-pairs like by
		// mail.Address.String, which the parser reads back unescaped, so the
		// name keeps them and an escaped quote doesn't end it. Whitespace
		// becomes a space and other runes that are not graphic, like the
		// control characters that would break a header line, are dropped.
		{name: `quotes and backslashes`, addr: &mail.Address{Name: `Say "Hi" \ Bye`, Address: "j@example.com"}, want: `"Say \"Hi\" \\ Bye" <j@example.com>`},
		{name: `control characters`, addr: &mail.Address{Name: "A\r\nB\x00C", Address: "j@example.com"}, want: `"A  BC" <j@example.com>`},
		{name: `only quotes`, addr: &mail.Address{Name: `""`, Address: "j@example.com"}, want: `"\"\"" <j@example.com>`},
		// A local part that needs quoting is quoted like by mail.Address.String
		{name: `quoted local part`, addr: &mail.Address{Address: "a b@example.com"}, want: `"a b"@example.com`},
		// A hand-built value without @ is no address to format,
		// it is returned unchanged instead of with an appended @
		{name: `no @`, addr: &mail.Address{Address: "foo"}, want: `foo`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AddressFrom(tt.addr); got != tt.want {
				t.Errorf("AddressFrom() = %s, want %s", got, tt.want)
			}
		})
	}
}

// TestAddress_ValidAndValidAndNormalized protects the difference between
// "parsable" and "already in canonical form": Valid accepts the lenient
// formats seen in the wild, while ValidAndNormalized is only true for the
// exact string Normalized would produce, so callers can use it to detect
// values that still need to be rewritten before storing or comparing them.
func TestAddress_ValidAndValidAndNormalized(t *testing.T) {
	tests := []struct {
		addr               Address
		valid              bool
		validAndNormalized bool
	}{
		{addr: `user@example.com`, valid: true, validAndNormalized: true},
		{addr: `"John Doe" <john@example.com>`, valid: true, validAndNormalized: true},
		// Upper case address parts are valid but not normalized because
		// normalized addresses are compared as lower case strings.
		{addr: `User@Example.com`, valid: true, validAndNormalized: false},
		// An unquoted name is valid but the normalized form quotes it.
		{addr: `John Doe <john@example.com>`, valid: true, validAndNormalized: false},
		// Angle brackets without a name are dropped by normalization.
		{addr: `<john@example.com>`, valid: true, validAndNormalized: false},
		{addr: ``, valid: false},
		{addr: `not an address`, valid: false},
		{addr: `@example.com`, valid: false},
		{addr: `user@`, valid: false},
		{addr: `user@example`, valid: false}, // a top level domain is required
		// A backslash can't escape the closing quote of a local part,
		// unescaped it would leave the empty local part of @example.com
		{addr: `"\"@example.com`, valid: false},
	}
	for _, tt := range tests {
		t.Run(string(tt.addr), func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.addr.Valid(), "Valid")
			assert.Equal(t, tt.valid, tt.addr.Validate() == nil, "Validate must agree with Valid")
			assert.Equal(t, tt.validAndNormalized, tt.addr.ValidAndNormalized(), "ValidAndNormalized")
		})
	}
}

// TestAddress_NormalizedIsIdempotent protects the string round-trip:
// a normalized address must parse back to itself, otherwise values stored
// in normalized form would keep changing every time they are normalized.
func TestAddress_NormalizedIsIdempotent(t *testing.T) {
	for _, addr := range []Address{
		`John Doe <John.Doe@Example.com>`,
		`"Doe, John" <john@example.com>`,
		`<Noname@Example.com>`,
		`=?utf-8?b?wqFIb2xhLCBzZcOxb3Ih?= <senor@hola.com>`,
		`Jürgen Müller <J@X.de>`,
		`"Müller, Jürgen" <j@x.de>`,
		// A quoted local part with an angle bracket without name, which
		// was written without angle brackets and read back as the display
		// name `"` of a@b.com, which is an error
		`<"<a@b.com>"@EXAMPLE.com>`,
		`<"x <a@b.com>"@example.com>`,
		// Names holding "=?", which were written as B encoded-word that
		// decodes to ASCII with an @, which the parser refuses, see
		// TestParseAddress_EncodedWordThatPythonReadsAsAddress. The names
		// that Python decodes to an address or another encoded-word after
		// writing them are refused now, see
		// TestParseAddress_EncodedWordThatPythonDecodesAfterWriting
		`"=\?utf-8?q?x?=" <ceo@corp.com>`,
		`"Really=?" <a@example.com>`,
	} {
		t.Run(string(addr), func(t *testing.T) {
			once, err := addr.Normalized()
			require.NoError(t, err)
			twice, err := once.Normalized()
			require.NoError(t, err)
			assert.Equal(t, once, twice)
			assert.True(t, once.ValidAndNormalized())
		})
	}
}

// FuzzAddressNormalizedIsIdempotent checks TestAddress_NormalizedIsIdempotent
// for every address: Normalized of a valid address is valid and normalized,
// so that Value writes an address that reads back as itself. For the seeds
// from TestAddress_NormalizedIsIdempotent, Normalized wrote an address that
// ParseAddress refused.
func FuzzAddressNormalizedIsIdempotent(f *testing.F) {
	for _, addr := range []string{
		`John Doe <John.Doe@Example.com>`,
		`<"<a@b.com>"@EXAMPLE.com>`,
		`<"x <a@b.com>"@example.com>`,
		`"=\? x@evil.com" <ceo@corp.com>`,
		`"=?utf-8?q?=3D=3Futf-8=3Fq=3Fx=3D40evil.com=3F=3D?=" <ceo@corp.com>`,
		"\"\u212Aate\"@Corp.com",
		`=?utf-8?Q?J=C3=BCrgen_=28juergen=40example.com=29?= <ceo@corp.com>`,
		`mike@doe.com (Mike (Sales))`,
		// Ordinary names that the parser refused for a while, see
		// TestParseAddress_OrdinaryEncodedNames
		`=?utf-8?b?SsO8cmdlbiBNw7xsbGVyLCBCdWNoaGFsdHVuZyB1bmQgRmluYW56ZW4gKEFi?= =?utf-8?b?dGVpbHVuZyBOb3JkKQ==?= <j.mueller@example.com>`,
		`=?utf-8?q?J=C3=BCrgen_M=C3=BCller_[Extern]?= <j@example.de>`,
		`"Really=?" <a@example.com>`,
		`M\u00FCller, =?utf-8?q?J=C3=BCrgen?= <j@x.de>`,
		`Doe; =?utf-8?q?J=C3=BCrgen?= <ceo@corp.com>`,
		`[Ext] =?utf-8?q?J=C3=BCrgen?= <j@x.de>`,
		`=?utf-8?q?J=C3=BCrgen_M=C3=BCller=2C_Buchhaltung_und_Finanzen_=28Abteilung_Nord=29?= <j.mueller@example.com>`,
	} {
		f.Add(addr)
	}
	f.Fuzz(func(t *testing.T, addr string) {
		once, err := Address(addr).Normalized()
		if err != nil {
			return
		}
		twice, err := once.Normalized()
		if err != nil {
			t.Fatalf("Normalized(%q) = %q, which is invalid: %v", addr, once, err)
		}
		if twice != once {
			t.Fatalf("Normalized(%q) = %q, but Normalized of it = %q", addr, once, twice)
		}
	})
}

// TestAddress_NormalizedRoundTrip protects that what Value, MarshalJSON
// and String persist is the same mailbox as the valid input: normalizing
// is idempotent, and the written value parses back as exactly one address
// with the same address part. A quoted local part that lost its quotes
// became another mailbox or even another domain, and a display name with
// an escaped quote, which the parser doesn't read, split into two
// addresses, so stored data could name an extra recipient.
func TestAddress_NormalizedRoundTrip(t *testing.T) {
	for _, addr := range []Address{
		`"a b"@example.com`,
		`"Doe, John"@example.com`,
		`"victim@evil.com>"@example.com`,
		`a@trusted.com (x" attacker@evil.com, z)`,
		`=?utf-8?q?Bank=E2=80=AE_x=C2=85y?= <a@example.com>`,
		// Local parts that net/mail writes unquoted but the parser doesn't
		// read back unquoted: a leading apostrophe it strips, and letters
		// it only accepts within quotes
		`"'abc"@example.com`,
		`"дмитрий"@example.com`,
		`Jürgen Müller <J@Example.com>`,
		`"Müller, Jürgen" <j@x.de>`,
		`mike@doe.com (Mike Doe)`,
		// The parser unescapes quoted-pairs and the formatter escapes
		// again only what needs it. If one side is skipped, every write
		// adds backslashes and the stored mailbox drifts away.
		`"a\ b"@example.com`,
		`"a\\b"@example.com`,
		// A name with an =? would be decoded again by every normalization
		// if it was written back as quoted text without the escape `=\?`.
		// One that decodes to an encoded-word is refused, see
		// TestParseAddress_EncodedWordThatPythonDecodesAfterWriting
		`"=\?utf-8?q?x?=" <a@example.com>`,
		`"Really=?" <a@example.com>`,
		// An encoded name with invalid UTF-8 decodes to a replacement
		// character, which the parser removes from the written name
		`=?utf8?B?X700?=<0@0.AA`,
	} {
		t.Run(string(addr), func(t *testing.T) {
			require.NoError(t, addr.Validate())
			wantPart, err := addr.AddressPart()
			require.NoError(t, err)

			norm, err := addr.Normalized()
			require.NoError(t, err)
			again, err := norm.Normalized()
			require.NoError(t, err)
			assert.Equal(t, norm, again, "Normalized must be idempotent")
			for _, r := range string(norm) {
				assert.True(t, r == ' ' || unicode.IsGraphic(r), "rune %U in %q", r, norm)
			}

			v, err := addr.Value()
			require.NoError(t, err)
			written, err := Address(v.(string)).AsList().SplitNormalized()
			require.NoError(t, err)
			require.Len(t, written, 1, "Value %q must be one address", v)
			gotPart, err := written[0].AddressPart()
			require.NoError(t, err)
			assert.Equal(t, wantPart, gotPart, "Value %q must be the same mailbox", v)
		})
	}
}

func TestAddress_Normalized_InvalidReturnsInput(t *testing.T) {
	// On error the original value is returned unchanged so that callers
	// that ignore the error don't lose the data they passed in.
	const addr Address = "not an address"
	norm, err := addr.Normalized()
	assert.Error(t, err)
	assert.Equal(t, addr, norm)
}

// Value and MarshalJSON write an Address normalized, because Validate
// accepts other spellings like an upper case address part or an
// RFC 2047 encoded name: without normalization the same address would
// reach SQL and JSON as different strings. Value refuses an address that
// can't be normalized, because the database is where an invalid value
// would stay, and MarshalJSON writes it unchanged, so a response that
// holds one still renders.
func TestAddress_ValueAndMarshalJSONNormalize(t *testing.T) {
	tests := []struct {
		addr Address
		want string
	}{
		{addr: `john@example.com`, want: `john@example.com`},
		{addr: `John.Doe@Example.com`, want: `john.doe@example.com`},
		{addr: `John Doe <John@Example.com>`, want: `"John Doe" <john@example.com>`},
		{addr: `=?utf-8?q?J=C3=BCrgen?= <j@example.com>`, want: `"Jürgen" <j@example.com>`},
	}
	for _, tt := range tests {
		v, err := tt.addr.Value()
		require.NoError(t, err)
		assert.Equal(t, tt.want, v, "Address(%q).Value()", string(tt.addr))

		j, err := json.Marshal(tt.addr)
		require.NoError(t, err)
		var got string
		require.NoError(t, json.Unmarshal(j, &got))
		assert.Equal(t, tt.want, got, "json.Marshal(Address(%q))", string(tt.addr))
	}

	_, err := Address(`not an address`).Value()
	assert.Error(t, err, "Value() of an invalid address")
	j, err := json.Marshal(Address(`not an address`))
	require.NoError(t, err)
	assert.Equal(t, `"not an address"`, string(j))

	v, err := Address(``).Value()
	require.NoError(t, err)
	assert.Nil(t, v, "an empty Address is written as SQL NULL")
}

// MarshalText writes the normalized address with a non-ASCII name as UTF-8
// for readability, ASCII the same with the name RFC 2047 encoded for
// message headers. Both write the original if it can't be normalized,
// so an invalid address is still shown instead of an empty string.
func TestAddress_MarshalTextAndASCII(t *testing.T) {
	tests := []struct {
		addr      Address
		wantStr   string
		wantASCII string
	}{
		{addr: `John@Example.com`, wantStr: `john@example.com`, wantASCII: `john@example.com`},
		{addr: `John Doe <John@Example.com>`, wantStr: `"John Doe" <john@example.com>`, wantASCII: `"John Doe" <john@example.com>`},
		{addr: `Jürgen Müller <J@Example.com>`, wantStr: `"Jürgen Müller" <j@example.com>`, wantASCII: `=?utf-8?q?J=C3=BCrgen_M=C3=BCller?= <j@example.com>`},
		// The address part is quoted like the parser reads it back,
		// net/mail would write 'abc@example.com, which is abc@example.com
		{addr: `"N" <"'abc"@example.com>`, wantStr: `"N" <"'abc"@example.com>`, wantASCII: `"N" <"'abc"@example.com>`},
		// A quoted local part with an angle bracket without name is written
		// in angle brackets, so that it doesn't read back as a display name,
		// see TestAddress_NormalizedIsIdempotent
		{addr: `<"<a@b.com>"@Example.com>`, wantStr: `<"<a@b.com>"@example.com>`, wantASCII: `<"<a@b.com>"@example.com>`},
		{addr: `not an address`, wantStr: `not an address`, wantASCII: `not an address`},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.wantStr, marshalText(t, tt.addr), "Address(%q).MarshalText()", string(tt.addr))
		assert.Equal(t, tt.wantASCII, tt.addr.ASCII(), "Address(%q).ASCII()", string(tt.addr))
	}

	// The ASCII form parses back to the same normalized address
	ascii := Address(Address(`Jürgen Müller <J@Example.com>`).ASCII())
	assert.Equal(t, `"Jürgen Müller" <j@example.com>`, marshalText(t, ascii))

	// A long non-ASCII name with special characters is written as one
	// B encoded-word, if a word of at most the 75 characters of RFC 2047
	// would hold only ASCII text with an @: mime.BEncoding.Encode split it
	// so that the last word held only the ASCII text
	// "rechnungen@mueller-gmbh.de)", which Python's email package writes
	// unencoded and unquoted, and which the parser refuses for that reason
	const longName = "Müller GmbH Buchhaltung und Rechnungswesen (rechnungen@mueller-gmbh.de)"
	long := Address(`"` + longName + `" <r@mueller-gmbh.de>`)
	parsed, err := ParseAddress(long.ASCII())
	require.NoError(t, err, long.ASCII())
	assert.Equal(t, longName, parsed.Name)
	// Else it is split into words of at most 75 characters like RFC 2047
	// requires, here like a mail client wrote the name, see
	// TestParseAddress_OrdinaryEncodedNames: the ASCII text with only a ")"
	// in the last word Python reads as part of the name
	split := Address(`"Jürgen Müller, Buchhaltung und Finanzen (Abteilung Nord)" <j.mueller@example.com>`)
	assert.Equal(t, `=?utf-8?b?SsO8cmdlbiBNw7xsbGVyLCBCdWNoaGFsdHVuZyB1bmQgRmluYW56ZW4gKEFi?= =?utf-8?b?dGVpbHVuZyBOb3JkKQ==?= <j.mueller@example.com>`, split.ASCII())
	for word := range strings.FieldsSeq(split.ASCII()) {
		assert.LessOrEqual(t, len(word), 75, word)
	}
	assert.Equal(t, string(split), marshalText(t, Address(split.ASCII())))

	// A name holding "=?" is written quoted with every =? escaped as `=\?`
	// by MarshalText and ASCII, so that ParseAddress and net/mail read the
	// name as written instead of decoding it once more and showing other
	// text. It was written as B encoded-word, which Python's email package
	// writes decoded, unencoded and unquoted, so that it read the name
	// "=? x@evil.com" as the address `"=? x"@evil.com`. A name that decodes
	// to an encoded-word is refused, see
	// TestParseAddress_EncodedWordThatPythonDecodesAfterWriting.
	const literal = "=?utf-8?q?x?="
	encoded := Address(`"=\?utf-8?q?x?=" <A@example.com>`)
	assert.Equal(t, `"=\?utf-8?q?x?=" <a@example.com>`, marshalText(t, encoded))
	assert.Equal(t, `"=\?utf-8?q?x?=" <a@example.com>`, encoded.ASCII())
	for _, written := range []string{marshalText(t, encoded), encoded.ASCII()} {
		name, err := Address(written).NamePart()
		require.NoError(t, err)
		assert.Equal(t, literal, name, written)
		stdlib, err := mail.ParseAddress(written)
		require.NoError(t, err)
		assert.Equal(t, literal, stdlib.Name, written)
	}
}

// TestAddress_Parts protects the accessors used to compare and route
// addresses: the name part is returned verbatim (decoded and unquoted),
// while the address and local parts are lower cased so that differently
// cased spellings of the same mailbox compare equal.
func TestAddress_Parts(t *testing.T) {
	tests := []struct {
		addr        Address
		name        string
		addressPart Address
		localPart   string
	}{
		{addr: `user@example.com`, name: ``, addressPart: `user@example.com`, localPart: `user`},
		{addr: `<User@Example.com>`, name: ``, addressPart: `user@example.com`, localPart: `user`},
		{addr: `Jane <Jane+Invoices@Example.com>`, name: `Jane`, addressPart: `jane+invoices@example.com`, localPart: `jane+invoices`},
		// The comma in the quoted name must not split the address.
		{addr: `"Doe, John" <John.Doe@Example.com>`, name: `Doe, John`, addressPart: `john.doe@example.com`, localPart: `john.doe`},
		// RFC 2047 encoded names are decoded.
		{addr: `=?utf-8?b?wqFIb2xhLCBzZcOxb3Ih?= <senor@hola.com>`, name: `¡Hola, señor!`, addressPart: `senor@hola.com`, localPart: `senor`},
		// A quoted local part keeps its quotes, else the address part
		// would read as another mailbox, here one at evil.com, and the
		// local part ends at the last @ because the quoted part has one.
		{addr: `"Victim@Evil.com>"@Example.com`, name: ``, addressPart: `"victim@evil.com>"@example.com`, localPart: `"victim@evil.com>"`},
		{addr: `"a b"@example.com`, name: ``, addressPart: `"a b"@example.com`, localPart: `"a b"`},
	}
	for _, tt := range tests {
		t.Run(string(tt.addr), func(t *testing.T) {
			name, err := tt.addr.NamePart()
			require.NoError(t, err)
			assert.Equal(t, tt.name, name, "NamePart")

			addressPart, err := tt.addr.AddressPart()
			require.NoError(t, err)
			assert.Equal(t, tt.addressPart, addressPart, "AddressPart")

			addressPartString, err := tt.addr.AddressPartString()
			require.NoError(t, err)
			assert.Equal(t, string(tt.addressPart), addressPartString, "AddressPartString")

			localPart, err := tt.addr.LocalPart()
			require.NoError(t, err)
			assert.Equal(t, tt.localPart, localPart, "LocalPart")
		})
	}
}

func TestAddress_Parts_InvalidAddress(t *testing.T) {
	// Invalid addresses must report an error instead of
	// returning a partial or made up part.
	const addr Address = "not an address"

	name, err := addr.NamePart()
	assert.Error(t, err, "NamePart")
	assert.Empty(t, name)

	addressPart, err := addr.AddressPart()
	assert.Error(t, err, "AddressPart")
	assert.Empty(t, addressPart)

	localPart, err := addr.LocalPart()
	assert.Error(t, err, "LocalPart")
	assert.Empty(t, localPart)
}

// TestAddress_DomainPart protects that the domain is taken from the
// actual address and not from an address-like display name,
// because the domain is used to identify the sender's organization.
func TestAddress_DomainPart(t *testing.T) {
	tests := []struct {
		addr Address
		want string
	}{
		{addr: `user@example.com`, want: `example.com`},
		{addr: `John <john@example.com>`, want: `example.com`},
		{addr: `John <john@example.com >`, want: `example.com`},
		{addr: "john@example.com\r\n", want: `example.com`},
		// The last @ wins, so an address in the name part is ignored.
		{addr: `"john@old.example.org" <john@new.example.com>`, want: `new.example.com`},
		// Without an @ there is no domain. Returning the whole string
		// would make arbitrary text look like a sender domain.
		{addr: ``, want: ``},
		{addr: `example.com`, want: ``},
		{addr: ` not an address `, want: ``},
	}
	for _, tt := range tests {
		t.Run(string(tt.addr), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.addr.DomainPart())
		})
	}
}

// TestAddress_Conversions protects that converting to the nullable
// and list types keeps the exact string so no data is lost or reformatted.
func TestAddress_Conversions(t *testing.T) {
	const addr Address = `"Doe, John" <john@example.com>`

	assert.Equal(t, NullableAddress(addr), addr.Nullable())

	list := addr.AsList()
	assert.Equal(t, AddressList(addr), list)
	split, err := list.SplitNormalized()
	require.NoError(t, err)
	assert.Equal(t, []Address{addr}, split, "the comma in the quoted name must not create a second list entry")
}

func TestAddress_JSONSchema(t *testing.T) {
	// The email format lets API clients validate addresses before sending them.
	schema := Address("").JSONSchema()
	assert.Equal(t, "string", schema.Type)
	assert.Equal(t, "email", schema.Format)
}

// marshalText returns the text that MarshalText writes for v,
// which is what encoding/xml and strfmt write.
func marshalText(t *testing.T, v encoding.TextMarshaler) string {
	t.Helper()
	text, err := v.MarshalText()
	require.NoError(t, err)
	return string(text)
}

// TestAddress_ReadingAndWriting checks the reading and writing contract
// shared by the normalizing types: every reading path stores a valid
// address normalized and an invalid one unchanged, Value refuses an invalid
// address, and fmt prints the stored address.
func TestAddress_ReadingAndWriting(t *testing.T) {
	normtest.Run(t, normtest.Spec[Address]{
		Canonical: `"John Doe" <john@example.com>`,
		Variants:  []Address{`John Doe <John@Example.com>`, ` "John Doe" <JOHN@example.com> `},
		Invalid:   []Address{"not an address", "john@"},
	})
}

// strfmt.Scan, which reads CSV files and other text tables, stores an
// address normalized through ScanString, like Scan and UnmarshalJSON do,
// instead of with the spelling of the cell, so that an address read from
// a table compares equal to the same address read from a database.
// Without validation an invalid address is kept, with it it's an error.
func TestAddress_StrfmtScan(t *testing.T) {
	validating := strfmt.NewScanConfig()
	lenient := strfmt.NewScanConfig()
	lenient.ValidateFunc = nil
	for _, config := range []*strfmt.ScanConfig{validating, lenient} {
		var addr Address
		require.NoError(t, strfmt.Scan(reflect.ValueOf(&addr).Elem(), "John Doe <John@Example.com>", config))
		assert.Equal(t, Address(`"John Doe" <john@example.com>`), addr)
	}

	var addr Address
	require.NoError(t, strfmt.Scan(reflect.ValueOf(&addr).Elem(), "not an address", lenient))
	assert.Equal(t, Address("not an address"), addr)
	assert.Error(t, strfmt.Scan(reflect.ValueOf(&addr).Elem(), "not an address", validating))
}
