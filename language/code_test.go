package language

import (
	"database/sql/driver"
	"maps"
	"regexp"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodeNormalized(t *testing.T) {
	tests := []struct {
		name string
		in   Code
		want Code
	}{
		// ISO 639-1, case + whitespace.
		{"lowercase passes through", "en", EN},
		{"uppercase", "EN", EN},
		{"mixed case", "De", DE},
		{"trim whitespace", "  en  ", EN},
		{"trim + case", " De ", DE},

		// ISO 639-2/T and 639-3 (terminologic codes equal the 639-3 form).
		{"639-3 eng", "eng", EN},
		{"639-3 deu", "deu", DE},
		{"639-3 fra", "fra", FR},
		{"639-3 zho", "zho", ZH},
		{"639-3 spa", "spa", ES},
		{"639-3 uppercase", "ENG", EN},

		// ISO 639-2/B (bibliographic) variants where they differ from /T.
		{"639-2/B ger → de", "ger", DE},
		{"639-2/B fre → fr", "fre", FR},
		{"639-2/B chi → zh", "chi", ZH},
		{"639-2/B dut → nl", "dut", NL},
		{"639-2/B alb → sq", "alb", SQ},

		// BCP-47: keep only the language subtag.
		{"BCP-47 en-US", "en-US", EN},
		{"BCP-47 zh-Hant-CN", "zh-Hant-CN", ZH},
		{"BCP-47 sr-Latn", "sr-Latn", SR},
		{"BCP-47 de-AT", "de-AT", DE},
		{"BCP-47 zh-hant-cn", "zh-hant-cn", ZH},

		// POSIX-style underscore separator.
		{"POSIX en_US", "en_US", EN},
		{"POSIX de_AT", "de_AT", DE},

		// 639-3 + BCP-47 region.
		{"eng-US (3-letter + region)", "eng-US", EN},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.in.Normalized()
			require.NoError(t, err, "Normalized(%q)", tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCodeNormalizedErrors(t *testing.T) {
	tests := []struct {
		name string
		in   Code
	}{
		{"empty", ""},
		{"only whitespace", "   "},
		{"single letter", "e"},
		{"unknown 2-letter", "xx"},
		{"unknown 3-letter (no 639-1 equivalent)", "abq"}, // valid 639-3, but no 639-1
		{"gibberish", "qwerty"},
		{"language subtag too long", "english"},
		{"separator at start", "-en"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.in.Normalized()
			require.Error(t, err, "Normalized(%q) should error", tt.in)
			// Errors preserve the original input.
			assert.Equal(t, tt.in, got, "error path should return input unchanged")
		})
	}
}

func TestCodeValid(t *testing.T) {
	// Valid() should still reflect strict ISO 639-1 membership, not the
	// liberal acceptance of Normalized().
	assert.True(t, EN.Valid())
	assert.True(t, Code("en").Valid())
	assert.False(t, Code("EN").Valid(), "Valid() is case-sensitive on the underlying map")
	assert.False(t, Code("eng").Valid(), "Valid() does not accept 639-3 directly")
	assert.False(t, Code("en-US").Valid(), "Valid() does not accept BCP-47 directly")
}

func TestCodeValidAndNormalized(t *testing.T) {
	assert.True(t, Code("en").ValidAndNormalized())
	assert.False(t, Code("EN").ValidAndNormalized())
	assert.False(t, Code("eng").ValidAndNormalized())
	assert.False(t, Code("en-US").ValidAndNormalized())
}

func TestCodeLanguageName(t *testing.T) {
	assert.Equal(t, "English", EN.LanguageName())
	assert.Equal(t, "German", DE.LanguageName())
	assert.Equal(t, "", Code("xx").LanguageName())
	assert.Equal(t, "", Null.LanguageName())

	// Like Valid/Validate/String and country.Code.EnglishName, the name is
	// looked up for the normalized code, so every input that Validate
	// accepts has a name (e.g. language codes read from user input or HTTP headers).
	assert.Equal(t, "English", Code("EN").LanguageName())
	assert.Equal(t, "English", Code(" en ").LanguageName())
	assert.Equal(t, "English", Code("eng").LanguageName())
	assert.Equal(t, "English", Code("en-US").LanguageName())
	assert.Equal(t, "German", Code("ger").LanguageName())
}

func TestParseCode(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want Code
	}{
		// 1. Short path: already a canonical ISO 639-1 code.
		{"canonical 639-1", "en", EN},
		{"canonical 639-1 de", "de", DE},

		// 2. Codes and tags, resolved via Normalized.
		{"uppercase", "EN", EN},
		{"whitespace", "  de  ", DE},
		{"639-3 code", "eng", EN},
		{"639-2/B bibliographic", "ger", DE},
		{"BCP-47 tag", "en-US", EN},
		{"POSIX locale", "zh_Hant", ZH},

		// 3. English language name.
		{"name", "German", DE},
		{"name lowercase", "english", EN},
		{"name with whitespace", "  French  ", FR},
		{"name primary from synonym list", "Catalan", CA},
		{"name synonym from synonym list", "Valencian", CA},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCode(tt.in)
			require.NoError(t, err, "ParseCode(%q)", tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseCodeErrors(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"whitespace only", "   "},
		{"gibberish", "qwerty"},
		{"unknown 2-letter", "xx"},
		{"639-3 without 639-1 equivalent", "abq"},
		{"language name without 639-1 code", "Klingon"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCode(tt.in)
			require.Error(t, err, "ParseCode(%q) should error", tt.in)
			// Errors preserve the original input.
			assert.Equal(t, Code(tt.in), got, "error path should return input unchanged")
		})
	}
}

// TestCode_EnumsMatchCodeNames ties the two lists of language codes together:
// Enums() is generated from the constants in code.go, but Valid() looks up
// codeNames in constants.go. A code added to only one of them would make
// Enums() list a code that Valid() rejects, or miss one that it accepts.
func TestCode_EnumsMatchCodeNames(t *testing.T) {
	assert.ElementsMatch(t, slices.Collect(maps.Keys(codeNames)), Code("").Enums())
}

// TestCode_ValidateStringAndNullSentinel covers the three parts of Code's
// public surface that the //#enum migration created or had to protect.
//
// Validate carries a #custom marker because it normalizes first, like
// country.Code and money.Currency, so input that types.Validate and
// DeepValidate receive from users ("EN", "eng", "en-US") is not rejected.
// The strict exact-match Validate that go-enum would generate instead
// rejects all of them. Valid deliberately stays strict, see TestCodeValid.
// String carries a #custom marker because it normalizes; the plain string
// conversion go-enum generates instead would return "EN" and "eng"
// unchanged. Null is declared as the conversion Code("") rather than
// `Null Code = ""` so that go-enum keeps it out of the enum value lists
// while it stays a typed Code constant with working methods.
//
// Value: protects=language.Code's #custom Validate accepts what Normalized
// accepts, its #custom String normalizes, and Null stays a typed sentinel
// outside Enums and EnumStrings;
// fails_when=go-enum regenerates Validate as the strict lookup or String as
// the plain conversion, or Null is redeclared as `Null Code = ""`;
// why_new=Validate and String have no test at all and
// TestCode_EnumsMatchCodeNames never looks at EnumStrings or Null;
// seam=none
func TestCode_ValidateStringAndNullSentinel(t *testing.T) {
	t.Run("Validate accepts every enumerated code", func(t *testing.T) {
		for _, c := range Code("").Enums() {
			require.NoError(t, c.Validate(), "Code(%q).Validate()", string(c))
		}
	})

	t.Run("Validate accepts what Normalized accepts", func(t *testing.T) {
		// Valid rejects all of these, Validate must not.
		for _, in := range []Code{"EN", " de ", "eng", "ger", "en-US", "zh_Hant"} {
			require.NoError(t, in.Validate(), "Code(%q).Validate()", string(in))
		}
	})

	t.Run("Validate rejects what Normalized rejects", func(t *testing.T) {
		for _, in := range []Code{"", "xx", "zzz", "English"} {
			err := in.Validate()
			require.Error(t, err, "Code(%q).Validate() should error", string(in))
			// The error has to name the offending value to be diagnosable.
			assert.ErrorContains(t, err, string(in))
		}
	})

	t.Run("String normalizes", func(t *testing.T) {
		assert.Equal(t, "en", Code("en").String())
		assert.Equal(t, "en", Code("EN").String(), "String() must normalize the case")
		assert.Equal(t, "en", Code("eng").String(), "String() must normalize 639-3 to 639-1")
		assert.Equal(t, "en", Code("en-US").String(), "String() must drop BCP-47 subtags")
		// No error result, so an unresolvable code comes back unchanged
		// instead of being replaced by the empty string.
		assert.Equal(t, "xx", Code("xx").String())
	})

	t.Run("Null sentinel", func(t *testing.T) {
		// Compile-time: Null stays a typed Code constant, not an untyped
		// string, so it keeps its methods and can be compared to a Code.
		var _ Code = Null

		assert.False(t, Null.Valid(), "language.Null.Valid() must be false")
		assert.Error(t, Null.Validate())
		assert.NotContains(t, Code("").Enums(), Null, "Null is not a valid language code")
		assert.NotContains(t, Code("").EnumStrings(), string(Null))
	})

	t.Run("EnumStrings matches Enums", func(t *testing.T) {
		// Both lists are generated independently from the same constants, so
		// nothing but this assertion keeps them from drifting apart.
		enums := Code("").Enums()
		want := make([]string, len(enums))
		for i, c := range enums {
			want[i] = string(c)
		}
		assert.Equal(t, want, Code("").EnumStrings())
	})
}

// TestCode_ScanValue protects the SQL NULL mapping documented on Code:
// the empty Code (Null) is written as SQL NULL, and SQL NULL is scanned
// back as Null, so a nullable language column round-trips without turning
// NULL into an empty string or vice versa.
func TestCode_ScanValue(t *testing.T) {
	t.Run("Scan", func(t *testing.T) {
		tests := []struct {
			name  string
			value any
			want  Code
		}{
			{name: "string", value: "de", want: DE},
			{name: "bytes", value: []byte("fr"), want: FR},
			{name: "NULL", value: nil, want: Null},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				c := EN // non-zero start value to detect that Scan overwrites it
				require.NoError(t, c.Scan(tt.value))
				assert.Equal(t, tt.want, c)
			})
		}
	})

	t.Run("Scan unsupported type", func(t *testing.T) {
		c := EN
		err := c.Scan(42)
		require.Error(t, err)
		assert.Equal(t, EN, c, "failed Scan must not modify the Code")
	})

	t.Run("Value", func(t *testing.T) {
		v, err := Null.Value()
		require.NoError(t, err)
		assert.Nil(t, v, "Null must be written as SQL NULL, not as empty string")

		v, err = DE.Value()
		require.NoError(t, err)
		assert.Equal(t, driver.Value("de"), v)
	})

	t.Run("round-trip", func(t *testing.T) {
		for _, c := range []Code{Null, EN, DE, ZH} {
			v, err := c.Value()
			require.NoError(t, err)
			var scanned Code
			require.NoError(t, scanned.Scan(v))
			assert.Equal(t, c, scanned)
		}
	})
}

// TestCode_JSONSchema protects that the published JSON schema accepts every
// valid language code and rejects non-normalized forms, so schema based
// validation agrees with Code.ValidAndNormalized.
func TestCode_JSONSchema(t *testing.T) {
	schema := Code("").JSONSchema()
	require.NotNil(t, schema)
	assert.Equal(t, "string", schema.Type)
	pattern := regexp.MustCompile(schema.Pattern)
	for _, c := range Code("").Enums() {
		assert.True(t, pattern.MatchString(string(c)), "schema pattern must accept %q", c)
	}
	for _, s := range []string{"", "EN", "eng", "en-US", "e"} {
		assert.False(t, pattern.MatchString(s), "schema pattern must reject %q", s)
	}
}
