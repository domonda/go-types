package charset

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUTF8Encoding_Decode(t *testing.T) {
	enc := UTF8Encoding()
	assert.Equal(t, "UTF-8 Encoding", enc.String())

	// Files written by Windows tools often start with a UTF-8 BOM,
	// it must not end up as an invisible U+FEFF in the decoded text.
	decoded, err := enc.Decode(append([]byte(BOMUTF8), "Grüße"...))
	require.NoError(t, err)
	assert.Equal(t, "Grüße", string(decoded))

	// Without BOM the bytes are passed through
	decoded, err = enc.Decode([]byte("Grüße"))
	require.NoError(t, err)
	assert.Equal(t, "Grüße", string(decoded))

	decoded, err = enc.Decode(nil)
	require.NoError(t, err)
	assert.Empty(t, decoded)

	// Only a leading BOM is removed
	decoded, err = enc.Decode(append([]byte("a"), BOMUTF8...))
	require.NoError(t, err)
	assert.Equal(t, append([]byte("a"), BOMUTF8...), decoded)

	// The BOM of another encoding means the data is not UTF-8
	// and must not be passed through as if it was
	_, err = enc.Decode(append([]byte(BOMUTF16LE), 'h', 0))
	assert.Error(t, err)
}

func TestUTF8Encoding_EncodeDecodeRoundTrip(t *testing.T) {
	enc := UTF8Encoding()
	for _, text := range []string{"", "ASCII", "Grüße 100 €", "世界 😀"} {
		encoded, err := enc.Encode([]byte(text))
		require.NoError(t, err)
		// Encode does not add a BOM
		assert.Equal(t, text, string(encoded))
		decoded, err := enc.Decode(encoded)
		require.NoError(t, err)
		assert.Equal(t, text, string(decoded))
	}
}
