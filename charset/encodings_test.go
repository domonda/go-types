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
