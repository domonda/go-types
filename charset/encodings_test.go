package charset

import (
	"encoding/binary"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEncoding(t *testing.T) {
	// Encoding names come from user input, HTTP headers and email MIME
	// parts, so common spelling variants must resolve to the same Encoding.
	tests := []struct {
		name     string
		wantName string
	}{
		{name: "UTF-8", wantName: "UTF-8"},
		{name: "utf-8", wantName: "UTF-8"},
		{name: "utf8", wantName: "UTF-8"},
		{name: "UTF8", wantName: "UTF-8"},
		{name: "UTF-16LE", wantName: "UTF-16LE"},
		{name: "utf16le", wantName: "UTF-16LE"},
		{name: "utf-16be", wantName: "UTF-16BE"},
		{name: "UTF32LE", wantName: "UTF-32LE"},
		{name: "utf-32be", wantName: "UTF-32BE"},
		{name: "Windows 1252", wantName: "Windows 1252"},
		{name: "windows 1252", wantName: "Windows 1252"},
		{name: "WINDOWS 1252", wantName: "Windows 1252"},
		{name: "ISO 8859-1", wantName: "ISO 8859-1"},
		{name: "iso 8859-15", wantName: "ISO 8859-15"},
		{name: "KOI8-R", wantName: "KOI8-R"},
		{name: "Macintosh", wantName: "Macintosh"},
		// Encodings that share a charmap with another name
		{name: "ISO-8859-6E", wantName: "ISO-8859-6E"},
		{name: "iso-8859-8i", wantName: "ISO-8859-8I"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, err := GetEncoding(tt.name)
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, enc.Name())
		})
	}
}

// TestGetEncoding_EmptyName protects that a missing charset name, like an
// absent charset parameter of a Content-Type header, is an error. It must
// not silently resolve to some encoding: it used to return IBM Code Page 037
// (EBCDIC) with an empty name, which decodes ASCII text to garbage.
func TestGetEncoding_EmptyName(t *testing.T) {
	enc, err := GetEncoding("")
	assert.Error(t, err)
	assert.Nil(t, enc)
	assert.Panics(t, func() { MustGetEncoding("") })

	enc, err = GetEncoding(" ")
	assert.Error(t, err)
	assert.Nil(t, enc)
}

// TestGetEncoding_IANALabels protects that the charset labels registered by
// IANA, which is what HTTP and MIME Content-Type headers contain, resolve
// case-insensitively. A label of an encoding listed by EncodingNames must
// return the same Encoding as its display name, so the Encoding name and
// the encoded bytes don't depend on which label a sender used.
// That includes the BOM handling: a UTF-16BE or UTF-16LE label must not
// return a generic decoder that keeps a BOM as U+FEFF in the text and
// decodes a BOM of the other byte order as garbage instead of an error.
//
// Value: protects=the IANA aliases csUTF16BE and csUTF16LE return this
// package's UTF-16BE/LE Encoding with its BOM, like their display names;
// fails_when=GetEncoding drops its UTF16 IgnoreBOM cases and wraps x/text
// UTF-16 with the MIME name, which has the same name and bytes but No BOM;
// why_new=no UTF-16 alias row existed, and name plus encoded bytes can't
// tell the wrapped x/text encoding from this package's UTF-16 encoding;
// seam=none;
// no_claim=the "UTF-16" label with BOM detection, see
// TestGetEncoding_IANAOnlyEncodings
func TestGetEncoding_IANALabels(t *testing.T) {
	const text = "Grüße"
	tests := []struct {
		label    string
		wantName string
	}{
		{label: "windows-1252", wantName: "Windows 1252"},
		{label: "WINDOWS-1252", wantName: "Windows 1252"},
		{label: "ISO-8859-1", wantName: "ISO 8859-1"},
		{label: "iso-8859-1", wantName: "ISO 8859-1"},
		{label: "latin1", wantName: "ISO 8859-1"},
		{label: "Latin1", wantName: "ISO 8859-1"},
		{label: "ISO-8859-15", wantName: "ISO 8859-15"},
		{label: "csUTF8", wantName: "UTF-8"},
		{label: "ISO-8859-6-E", wantName: "ISO-8859-6E"},
		{label: "csUTF16BE", wantName: "UTF-16BE"},
		{label: "csUTF16LE", wantName: "UTF-16LE"},
	}
	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			enc, err := GetEncoding(tt.label)
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, enc.Name())

			byName := MustGetEncoding(tt.wantName)
			want, wantErr := byName.Encode([]byte(text))
			got, err := enc.Encode([]byte(text))
			assert.Equal(t, wantErr != nil, err != nil, "same encoding error as %s", byName)
			assert.Equal(t, want, got, "same encoded bytes as %s", byName)
			assert.Equal(t, byName.BOM(), enc.BOM(), "same BOM as %s", byName)
		})
	}
}

// TestGetEncoding_IANAOnlyEncodings protects that IANA charsets without
// an own implementation in this package, including multi-byte ones,
// are usable and named by their MIME name.
func TestGetEncoding_IANAOnlyEncodings(t *testing.T) {
	t.Run("US-ASCII", func(t *testing.T) {
		for _, label := range []string{"US-ASCII", "us-ascii", "csASCII"} {
			enc, err := GetEncoding(label)
			require.NoError(t, err, label)
			assert.Equal(t, "US-ASCII", enc.Name())
		}
		enc := MustGetEncoding("us-ascii")
		encoded, err := enc.Encode([]byte("Hello"))
		require.NoError(t, err)
		assert.Equal(t, []byte("Hello"), encoded)
		// Data loss must be an error instead of a silent replacement.
		_, err = enc.Encode([]byte("Grüße"))
		assert.Error(t, err)
	})

	t.Run("Shift_JIS", func(t *testing.T) {
		enc, err := GetEncoding("shift_jis")
		require.NoError(t, err)
		assert.Equal(t, "Shift_JIS", enc.Name())
		encoded, err := enc.Encode([]byte("日本"))
		require.NoError(t, err)
		assert.Equal(t, []byte{0x93, 0xFA, 0x96, 0x7B}, encoded)
		decoded, err := enc.Decode(encoded)
		require.NoError(t, err)
		assert.Equal(t, "日本", string(decoded))
	})

	// RFC 2781: "UTF-16" takes the byte order from a leading BOM
	// and is big endian without one. Encoding writes a big endian BOM
	// so the byte order is unambiguous for the receiver.
	t.Run("UTF-16", func(t *testing.T) {
		for _, label := range []string{"UTF-16", "utf-16", "utf16"} {
			enc, err := GetEncoding(label)
			require.NoError(t, err, label)
			assert.Equal(t, "UTF-16", enc.Name())
		}
		enc := MustGetEncoding("UTF-16")
		for name, data := range map[string][]byte{
			"BE BOM": {0xFE, 0xFF, 0x00, 'h', 0x00, 'i'},
			"LE BOM": {0xFF, 0xFE, 'h', 0x00, 'i', 0x00},
			"no BOM": {0x00, 'h', 0x00, 'i'},
		} {
			decoded, err := enc.Decode(data)
			require.NoError(t, err, name)
			assert.Equal(t, "hi", string(decoded), name)
		}
		encoded, err := enc.Encode([]byte("hi"))
		require.NoError(t, err)
		assert.Equal(t, []byte{0xFE, 0xFF, 0x00, 'h', 0x00, 'i'}, encoded)
	})
}

func TestGetEncoding_NotFound(t *testing.T) {
	for _, name := range []string{"UTF-7", "no such encoding", "Windows 9999"} {
		t.Run(name, func(t *testing.T) {
			enc, err := GetEncoding(name)
			assert.Error(t, err)
			assert.Nil(t, enc)
			assert.Panics(t, func() { MustGetEncoding(name) })
		})
	}
}

func TestMustGetEncoding(t *testing.T) {
	assert.Equal(t, "UTF-16BE", MustGetEncoding("utf16be").Name())
	assert.Equal(t, "Windows 1252", MustGetEncoding("windows 1252").Name())
}

func TestEncodingNames(t *testing.T) {
	names := EncodingNames()
	assert.True(t, slices.IsSorted(names), "names must be sorted")
	assert.Equal(t, len(names), len(slices.Compact(slices.Clone(names))), "names must be unique")
	for _, want := range []string{"UTF-8", "UTF-16LE", "UTF-16BE", "UTF-32LE", "UTF-32BE", "Windows 1252", "ISO 8859-1", "ISO-8859-8I"} {
		assert.Contains(t, names, want)
	}

	// Every advertised name must be resolvable by GetEncoding to an
	// Encoding with exactly that name which can round-trip ASCII text,
	// otherwise listing the names to a user would offer broken choices.
	const ascii = "Hello, World 123!"
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			enc, err := GetEncoding(name)
			require.NoError(t, err)
			assert.Equal(t, name, enc.Name())

			encoded, err := enc.Encode([]byte(ascii))
			require.NoError(t, err)
			decoded, err := enc.Decode(encoded)
			require.NoError(t, err)
			assert.Equal(t, ascii, string(decoded))
		})
	}
}

func TestAutoDecode_Empty(t *testing.T) {
	text, encName, err := AutoDecode(nil, []Encoding{UTF8Encoding()}, []string{"a"})
	assert.NoError(t, err)
	assert.Nil(t, text)
	assert.Equal(t, "", encName)
}

func TestAutoDecode_BOM(t *testing.T) {
	// A BOM is definitive, the passed encodings and key-words are not used
	data := append([]byte(BOMUTF16BE), 0, 'h', 0, 'i')
	text, encName, err := AutoDecode(data, []Encoding{MustGetEncoding("Windows 1252")}, []string{"not found"})
	require.NoError(t, err)
	assert.Equal(t, "hi", string(text))
	assert.Equal(t, "UTF-16BE", encName)

	// Data that does not decode in the BOM's encoding is an error
	// and not silently returned as is
	_, _, err = AutoDecode(append([]byte(BOMUTF16LE), 'h', 0, 'i'), nil, nil)
	assert.Error(t, err)
}

func TestAutoDecode_KeyWords(t *testing.T) {
	windows1252 := MustGetEncoding("Windows 1252")
	latin1 := MustGetEncoding("ISO 8859-1")

	// "für 100 €" in Windows 1252 is 66 FC 72 20 31 30 30 20 80.
	// ISO 8859-1 decodes the ü but not the €, UTF-8 neither.
	data, err := windows1252.Encode([]byte("Rechnung für 100 €"))
	require.NoError(t, err)

	t.Run("most key-word matches wins", func(t *testing.T) {
		text, encName, err := AutoDecode(data, []Encoding{UTF8Encoding(), latin1, windows1252}, []string{"für", "€"})
		require.NoError(t, err)
		assert.Equal(t, "Rechnung für 100 €", string(text))
		assert.Equal(t, "Windows 1252", encName)
	})

	t.Run("tie goes to first encoding", func(t *testing.T) {
		// Both decode "für" equally well. The order of the passed
		// encodings expresses the caller's preference and must be kept.
		_, encName, err := AutoDecode(data, []Encoding{latin1, windows1252}, []string{"für"})
		require.NoError(t, err)
		assert.Equal(t, "ISO 8859-1", encName)
		_, encName, err = AutoDecode(data, []Encoding{windows1252, latin1}, []string{"für"})
		require.NoError(t, err)
		assert.Equal(t, "Windows 1252", encName)
	})

	t.Run("no key-word found returns data unchanged", func(t *testing.T) {
		text, encName, err := AutoDecode(data, []Encoding{latin1, windows1252}, []string{"Größe"})
		require.NoError(t, err)
		assert.Equal(t, data, text)
		assert.Equal(t, "", encName)
	})

	t.Run("no encodings returns data unchanged", func(t *testing.T) {
		text, encName, err := AutoDecode(data, nil, []string{"für"})
		require.NoError(t, err)
		assert.Equal(t, data, text)
		assert.Equal(t, "", encName)
	})

	t.Run("encoding with decode error is skipped", func(t *testing.T) {
		// "für" has an odd number of bytes in Windows 1252,
		// so UTF-16 fails to decode it. That must not fail AutoDecode,
		// even when the failing encoding is the last one tried.
		data, err := windows1252.Encode([]byte("für"))
		require.NoError(t, err)
		text, encName, err := AutoDecode(data, []Encoding{windows1252, UTF16Encoding(binary.LittleEndian)}, []string{"für"})
		require.NoError(t, err)
		assert.Equal(t, "für", string(text))
		assert.Equal(t, "Windows 1252", encName)
	})
}
