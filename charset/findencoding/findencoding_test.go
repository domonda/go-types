package findencoding

import (
	"fmt"
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
