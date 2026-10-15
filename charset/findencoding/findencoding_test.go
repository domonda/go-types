package findencoding

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintFileWithAllEncodings_FileErrors(t *testing.T) {
	// A file that can't be read is reported, not printed as empty text
	err := PrintFileWithAllEncodings(filepath.Join(t.TempDir(), "missing.txt"), 0)
	assert.ErrorIs(t, err, os.ErrNotExist)

	// A BOM is trusted: if the data after it does not decode in the
	// encoding of the BOM, that is an error instead of a fallback to
	// printing the data with all other encodings
	file := filepath.Join(t.TempDir(), "odd-utf16.txt")
	require.NoError(t, os.WriteFile(file, []byte{0xFE, 0xFF, 0x00, 'h', 'i'}, 0o600))
	err = PrintFileWithAllEncodings(file, 0)
	assert.Error(t, err)
}

// TestPrintFileWithAllEncodings_NegativeMaxBytes protects that a negative
// maxBytes means no limit instead of panicking with a slice bounds error,
// so a debugging helper can't crash the program that calls it.
func TestPrintFileWithAllEncodings_NegativeMaxBytes(t *testing.T) {
	file := filepath.Join(t.TempDir(), "utf8bom.txt")
	require.NoError(t, os.WriteFile(file, []byte("\xEF\xBB\xBFhello"), 0o600))
	assert.NotPanics(t, func() {
		assert.NoError(t, PrintFileWithAllEncodings(file, -1))
	})
}

// TestPrintFileWithAllEncodings_WithoutBOM protects the purpose of the
// helper: without a BOM the text is printed in every encoding, so the
// reader can spot the one in which it reads correctly. "Grüße" written
// as ISO 8859-1 bytes must come out readable under that heading and as
// UTF-8, where the single ü and ß bytes are invalid.
func TestPrintFileWithAllEncodings_WithoutBOM(t *testing.T) {
	file := filepath.Join(t.TempDir(), "latin1.txt")
	require.NoError(t, os.WriteFile(file, []byte("Gr\xfc\xdfe"), 0o600))

	stdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	printErr := PrintFileWithAllEncodings(file, 0)
	os.Stdout = stdout
	require.NoError(t, w.Close())
	out, err := io.ReadAll(r)
	require.NoError(t, err)

	require.NoError(t, printErr)
	assert.Contains(t, string(out), "ISO 8859-1\n--------------------\nGrüße\n")
	assert.Contains(t, string(out), "UTF-8\n--------------------\n")
	assert.NotContains(t, string(out), "IBM Code Page 037", "EBCDIC is skipped as never useful")
}

// A file with a BOM is only printed in the encoding of the BOM.
// maxBytes includes the BOM, here it cuts "hi there" to "hi".
func ExamplePrintFileWithAllEncodings() {
	dir, err := os.MkdirTemp("", "findencoding")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	file := filepath.Join(dir, "utf16le.txt")
	data := []byte{0xFF, 0xFE} // UTF-16LE BOM
	for _, c := range "hi there" {
		data = append(data, byte(c), 0)
	}
	err = os.WriteFile(file, data, 0o600)
	if err != nil {
		panic(err)
	}

	err = PrintFileWithAllEncodings(file, 6)
	if err != nil {
		fmt.Println("error:", err)
	}

	// Output:
	// Found BOM: UTF-16LE
	// ====================
	// UTF-16LE
	// --------------------
	// hi
	// ====================
}
