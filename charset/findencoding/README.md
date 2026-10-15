# findencoding

A debugging aid to find the character encoding of a text file by eye. It prints the file decoded with every legacy single byte charset, so you can pick the one where umlauts and currency signs look right, then decode with that charset via [`charset.GetEncoding`](../README.md#encoding). It is a library package, not a command, and it detects nothing except a Byte Order Mark.

```
import "github.com/domonda/go-types/charset/findencoding"
```

## PrintFileWithAllEncodings

```go
func PrintFileWithAllEncodings(filename string, maxBytes int) error
```

- Reads the whole file into memory and returns read errors.
- `maxBytes` limits how many bytes, including a BOM, are decoded and printed. `0` means 1 MiB, a negative value means no limit.
- If the file starts with a BOM, it prints `Found BOM: <name>` and the content decoded in the encoding of the BOM only. A decoding error is returned, there is no fallback to the other encodings.
- Otherwise it converts `\r\n`, `\n\r`, and `\r` line endings to `\n` and prints the content once for every `*charmap.Charmap` of `golang.org/x/text/encoding/charmap` except `IBM Code Page 037` (EBCDIC), and finally as UTF-8 without decoding.

Every decoding is printed to standard output as one block:

```
====================
<encoding name>
--------------------
<decoded text>
====================
```

## Examples

Call it from a throwaway test next to the code that reads the file, then compare the blocks in the verbose test output:

```go
func TestDebugEncoding(t *testing.T) {
	err := findencoding.PrintFileWithAllEncodings("testdata/unknown.csv", 2000)
	if err != nil {
		t.Fatal(err)
	}
}
```

```
go test -run TestDebugEncoding -v
```

A file with a BOM is only printed in the encoding of the BOM, and `maxBytes` counts the BOM:

```go
package main

import (
	"os"
	"path/filepath"

	"github.com/domonda/go-types/charset/findencoding"
)

func main() {
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

	// 6 bytes are the BOM and "hi"
	err = findencoding.PrintFileWithAllEncodings(file, 6)
	if err != nil {
		panic(err)
	}

	// Output:
	// Found BOM: UTF-16LE
	// ====================
	// UTF-16LE
	// --------------------
	// hi
	// ====================
}
```

## Related

- [`charset`](../README.md): `GetEncoding`, `AutoDecode`, and the BOM functions to decode the file once the encoding is known.
