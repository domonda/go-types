package language

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestISO6393Name(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{"eng", "English"},
		{"deu", "German"},
		{"fra", "French"},
		{"aaa", "Ghotuo"},   // first code in the index
		{"zzj", "Zuojiang Zhuang"},
		{"zza", "Dimili"},   // code with several names → primary name
		{"abq", "Abaza"},    // valid 639-3 without a 639-1 equivalent

		{"", ""},            // empty
		{"xxx", ""},         // not a known 639-3 code
		{"DEU", ""},         // case-sensitive: canonical form is lower-case
		{"de", ""},          // 639-1, not 639-3
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			assert.Equal(t, tt.want, ISO6393Name(tt.code))
		})
	}
}

func TestISO6393Macrolanguage(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{"twi", "aka"}, // Twi rolls up to Akan
		{"fat", "aka"}, // Fanti rolls up to Akan
		{"zzj", "zha"}, // Zuojiang Zhuang rolls up to Zhuang

		{"deu", ""}, // individual language, member of no macrolanguage
		{"aka", ""}, // a macrolanguage is not itself a member
		{"ajp", ""}, // retired mapping → not a live entry
		{"", ""},    // empty
		{"xxx", ""}, // not a known 639-3 code
		{"TWI", ""}, // case-sensitive: canonical form is lower-case
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			assert.Equal(t, tt.want, ISO6393Macrolanguage(tt.code))
		})
	}
}

// TestISO6393MapsPopulated guards against a generator regression that
// would emit empty or near-empty maps.
func TestISO6393MapsPopulated(t *testing.T) {
	assert.Greater(t, len(iso6393Names), 7000, "iso6393Names should hold the full ISO 639-3 set")
	assert.Greater(t, len(iso6393Macro), 400, "iso6393Macro should hold the active macrolanguage mappings")
}

// TestISO6393MacroMembersHaveNames checks that every macrolanguage code
// referenced by iso6393Macro is itself a known ISO 639-3 code with a
// reference name — i.e. the two generated maps stay consistent.
func TestISO6393MacroMembersHaveNames(t *testing.T) {
	for individual, macro := range iso6393Macro {
		assert.NotEmpty(t, ISO6393Name(individual), "individual code %q has no name", individual)
		assert.NotEmpty(t, ISO6393Name(macro), "macrolanguage code %q has no name", macro)
	}
}
