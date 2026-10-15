package strutil

import (
	"html"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestHTMLEscapeSpecialRunes(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want string
	}{
		{name: "empty", str: "", want: ""},
		{name: "plain ASCII unchanged", str: "Hello World 123", want: "Hello World 123"},
		{name: "German umlauts and sharp s", str: "Größe Übermaß", want: "Gr&ouml;&szlig;e &Uuml;berma&szlig;"},
		{name: "euro sign", str: "100 €", want: "100 &euro;"},
		{name: "typographic quotes and dashes", str: "„a“ – b — c…", want: "&bdquo;a&ldquo; &ndash; b &mdash; c&hellip;"},
		{name: "Greek letters", str: "αβΩ", want: "&alpha;&beta;&Omega;"},
		{name: "non-breaking space", str: "a\u00a0b", want: "a&nbsp;b"},
		// The ampersand has to be escaped too, otherwise already escaped
		// looking input could not be told apart from escaped output.
		{name: "ampersand", str: "A & B", want: "A &amp; B"},
		{name: "entity-like input is escaped again", str: "&auml;", want: "&amp;auml;"},
		// Runes without a named entity must pass through as valid UTF-8,
		// including those encoded with 3 or 4 bytes.
		{name: "runes without entity unchanged", str: "世界 😀 ł", want: "世界 😀 ł"},
		{name: "mixed", str: "ä世ü", want: "&auml;世&uuml;"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HTMLEscapeSpecialRunes(tt.str)
			assert.Equal(t, tt.want, got)
			assert.True(t, utf8.ValidString(got), "result must be valid UTF-8")
		})
	}
}

// TestHTMLEscapeSpecialRunes_Lossless checks that escaping can be undone
// by any HTML parser, so that no information is lost by escaping text
// before embedding it in HTML.
func TestHTMLEscapeSpecialRunes_Lossless(t *testing.T) {
	for _, str := range []string{
		"",
		"plain",
		"Größe & Übermaß",
		"&auml; is already escaped looking",
		"© 2024 domonda™ – „Rechnung“ № 1 ½ € ≠ $",
		"<p>markup</p>",
		"世界 😀",
	} {
		assert.Equal(t, str, html.UnescapeString(HTMLEscapeSpecialRunes(str)), "unescaping %q", str)
	}
}

// TestHTMLEscapeSpecialRunes_EntityTable checks every entry of the
// rune to entity table against the HTML entity definitions of the
// standard library so that a typo in the table can't produce entities
// that browsers render as a different character.
func TestHTMLEscapeSpecialRunes_EntityTable(t *testing.T) {
	for r, entity := range runeEntity {
		want := string(r)
		if r == '\u2006' {
			// Deliberate exception: the six-per-em space has no named
			// entity and is mapped to the regular non-breaking space.
			want = "\u00a0"
		}
		assert.Equal(t, want, html.UnescapeString(entity), "entity %s for rune %q (%U)", entity, r, r)
		assert.Equal(t, entity, HTMLEscapeSpecialRunes(string(r)), "rune %q (%U)", r, r)
	}
}
