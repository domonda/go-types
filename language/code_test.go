package language

import (
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
}
