# charset

Character encoding helpers for text from external sources: detect and strip a Unicode Byte Order Mark (BOM), convert between UTF-8 and UTF-16, UTF-32, or a legacy charset, look up an encoding by the charset label of an HTTP or MIME header, and guess the encoding of BOM-less data by counting key words. Every conversion goes through UTF-8 byte slices.

```
import "github.com/domonda/go-types/charset"
```

## Encoding

```go
type Encoding interface {
    Encode(utf8Str []byte) (encodedStr []byte, err error) // UTF-8 to encoded
    Decode(encodedStr []byte) (utf8Str []byte, err error) // encoded to UTF-8
    Name() string
    String() string // Name() + " Encoding"
    BOM() BOM       // NoBOM if the encoding has no BOM
}
```

All encodings returned by this package are safe for concurrent use. The ones based on `golang.org/x/text/encoding` serialize their `Encode` and `Decode` calls with a mutex.

| Function                                           | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `UTF8Encoding() Encoding`                          | UTF-8, see the table below.                        |
| `UTF16Encoding(byteOrder binary.ByteOrder) Encoding` | UTF-16 using `EncodeUTF16` and `DecodeUTF16`. Named `UTF-16BE` for `binary.BigEndian`, `UTF-16LE` for any other byte order. |
| `UTF32Encoding(byteOrder binary.ByteOrder) Encoding` | UTF-32 using `EncodeUTF32` and `DecodeUTF32`. Named `UTF-32BE` for `binary.BigEndian`, `UTF-32LE` otherwise. |
| `GetEncoding(name string) (Encoding, error)`       | Encoding for a name or a charset label, see below. |
| `MustGetEncoding(name string) Encoding`            | `GetEncoding` that panics on error.                |
| `EncodingNames() []string`                         | Sorted names of the encodings listed by this package: `UTF-8`, `UTF-16LE`, `UTF-16BE`, `UTF-32LE`, `UTF-32BE`, the display names of the `golang.org/x/text/encoding/charmap` charsets (`Windows 1252`, `ISO 8859-1`, `KOI8-R`, `Macintosh`, `IBM Code Page 437`, ...), and `ISO-8859-6E`, `ISO-8859-6I`, `ISO-8859-8E`, `ISO-8859-8I`. Every name resolves with `GetEncoding`. |

`GetEncoding` resolves a name in this order:

1. An empty name is an error.
2. The name is matched case-insensitively, and `UTF` without a dash also matches (`utf8` gives `UTF-8`, `utf16le` gives `UTF-16LE`).
3. The five UTF names return the shared UTF encodings of this package.
4. A `charmap` display name (`windows 1252`) returns that charset with its display name.
5. `ISO-8859-6E`, `ISO-8859-6I`, `ISO-8859-8E`, and `ISO-8859-8I` return ISO 8859-6 or ISO 8859-8 under that name.
6. Any other name is looked up as IANA charset label with `golang.org/x/text/encoding/ianaindex`, as found in `Content-Type: text/plain; charset=...` headers. A label of an encoding listed by `EncodingNames` returns that encoding with its display name: `windows-1252` gives `Windows 1252`, `latin1` and `ISO-8859-1` give `ISO 8859-1`, `csUTF8` gives `UTF-8`. Other supported IANA charsets, including multi-byte ones, are named by their MIME name: `US-ASCII`, `Shift_JIS`, `UTF-16`.
7. Everything else returns the error `encoding not found: "cp1252"`.

What the encodings do:

| Encoding                                   | `Name()`               | `BOM()`                    | `Encode`                                           | `Decode`                                           |
|--------------------------------------------|------------------------|----------------------------|----------------------------------------------------|----------------------------------------------------|
| `UTF8Encoding()`                           | `UTF-8`                | `BOMUTF8`                  | Returns the input unchanged.                       | Strips a UTF-8 BOM, returns an error for any other BOM. Does not validate UTF-8. |
| `UTF16Encoding(bo)`                        | `UTF-16LE`, `UTF-16BE` | `BOMUTF16LE`, `BOMUTF16BE` | `EncodeUTF16`, no BOM.                             | `DecodeUTF16`.                                     |
| `UTF32Encoding(bo)`                        | `UTF-32LE`, `UTF-32BE` | `BOMUTF32LE`, `BOMUTF32BE` | `EncodeUTF32`, no BOM.                             | `DecodeUTF32`.                                     |
| `GetEncoding` of a `charmap` name or label | e.g. `Windows 1252`    | `NoBOM`                    | `x/text` encoder, error for runes the charset can't represent. | `x/text` decoder.                                  |
| `GetEncoding("UTF-16")`                    | `UTF-16`               | `NoBOM`                    | Big endian with a BOM.                             | Byte order of a leading BOM, big endian without one. |
| `GetEncoding` of other IANA labels         | MIME name              | `NoBOM`                    | `x/text` encoder, error for runes the charset can't represent. | `x/text` decoder.                                  |

## BOM

`BOM` is a string holding the bytes of a Byte Order Mark.

| Value        | Bytes         | `String()` |
|--------------|---------------|------------|
| `NoBOM`      | empty         | `No BOM`   |
| `BOMUTF8`    | `EF BB BF`    | `UTF-8`    |
| `BOMUTF16BE` | `FE FF`       | `UTF-16BE` |
| `BOMUTF16LE` | `FF FE`       | `UTF-16LE` |
| `BOMUTF32BE` | `00 00 FE FF` | `UTF-32BE` |
| `BOMUTF32LE` | `FF FE 00 00` | `UTF-32LE` |

Any other value prints as `Invalid BOM: [...]`.

| Function / Method                                  | Description                                        |
|----------------------------------------------------|----------------------------------------------------|
| `BOMOfBytes(b) BOM`, `BOMOfString(s) BOM`          | The BOM at the start, `NoBOM` if none. 4 byte BOMs are checked before 2 byte ones. |
| `SplitBOM(b) (BOM, []byte)`                        | The detected BOM and the bytes after it.           |
| `TrimBOM(b, bom) []byte`                           | Removes `bom` if `b` starts with it. Any other BOM is kept, `NoBOM` removes nothing. |
| `DecodeWithBOM(b) ([]byte, error)`, `DecodeStringWithBOM(b) (string, error)` | `SplitBOM`, then decode the rest in the encoding of the BOM. Without a BOM the data is returned as UTF-8 unchanged. |
| `bom.Encoding() (Encoding, error)`                 | `UTF8Encoding()` for `NoBOM` and `BOMUTF8`, the UTF-16 or UTF-32 encoding of the byte order otherwise, an error for invalid values. |
| `bom.Endian() binary.ByteOrder`                    | `binary.LittleEndian` or `binary.BigEndian` for the UTF-16 and UTF-32 BOMs, nil for `NoBOM` and `BOMUTF8`. |
| `bom.Decode(data) ([]byte, error)`, `bom.DecodeString(data) (string, error)` | Decodes from the encoding of `bom`. A leading `bom` in `data` is stripped, a different BOM is an error. |
| `bom.String() string`                              | See the table above.                               |

### UTF-16LE vs. UTF-32LE detection

The UTF-16LE BOM (`FF FE`) is a prefix of the UTF-32LE BOM (`FF FE 00 00`), so
`BOMOfBytes` and `BOMOfString` match the 4 byte BOMs before the 2 byte ones.

These two cases are not distinguishable from the bytes alone: UTF-16LE text
whose first character is U+0000 serializes to the same `FF FE 00 00` and is
reported as `BOMUTF32LE`. Preferring the longer BOM is the conventional
resolution (ICU, .NET and `file(1)` all do it), because a leading NUL is not
plain text while real UTF-32LE data is. The Unicode standard lists the
signatures in table 23-6 without saying how to resolve the overlap.

The ambiguity only applies to detection. If the encoding is already known,
skip detection and pass the known BOM to `bom.Decode` or `bom.DecodeString`,
where `FF FE` can only mean UTF-16LE.

## UTF-16 and UTF-32 functions

Lower level functions for code that doesn't need the `Encoding` interface. The decoders strip a leading BOM that matches `byteOrder` and return an error like `expected UTF-16LE BOM but got UTF-16BE` for any other BOM. The encoders never write a BOM.

| Function                                          | Description                                        |
|---------------------------------------------------|----------------------------------------------------|
| `DecodeUTF16(b, byteOrder) ([]byte, error)`       | UTF-16 to UTF-8. Odd length is an error, unpaired surrogates decode to U+FFFD. |
| `DecodeUTF16String(b, byteOrder) (string, error)` | `DecodeUTF16` as string.                           |
| `EncodeUTF16(b, byteOrder) ([]byte, error)`       | UTF-8 to UTF-16. Invalid UTF-8 is an error, a real U+FFFD is encoded. |
| `DecodeUTF32(b, byteOrder) ([]byte, error)`       | UTF-32 to UTF-8. Trailing bytes and invalid code points decode to U+FFFD. |
| `DecodeUTF32String(b, byteOrder) (string, error)` | `DecodeUTF32` as string.                           |
| `EncodeUTF32(b, byteOrder) ([]byte, error)`       | UTF-8 to UTF-32. Invalid UTF-8 is encoded as U+FFFD. |

## AutoDecode

```go
func AutoDecode(data []byte, encodings []Encoding, keyWords []string) (text []byte, encName string, err error)
```

Guesses the encoding of `data`:

1. Empty data returns `nil, "", nil`.
2. If `data` starts with a BOM, the rest is decoded with the encoding of the BOM and `encName` is the BOM name, like `UTF-16BE`. A decoding error is returned.
3. Otherwise `data` is decoded with every passed encoding. Encodings that return an error are skipped. For each result the occurrences of all key words are counted (case-sensitive bytes). The result with the highest count wins, ties go to the encoding passed first.
4. If no key word is found, `data` is returned unchanged with `encName` `""` and no error.

## Gotchas

- **No `Encode` writes a BOM**, except the one of `GetEncoding("UTF-16")`. `UTF16Encoding(binary.LittleEndian).Encode([]byte("hi"))` returns `68 00 69 00`. Prepend `enc.BOM()` yourself, see the second example.
- **`BOM()` describes the encoding, not the output.** All `GetEncoding` results except the five UTF encodings return `NoBOM`, also `UTF-16`, which writes a BOM. `NoBOM.Encoding()` returns UTF-8, whose `BOM()` is `BOMUTF8`.
- **`UTF8Encoding().Decode` does not validate.** `Decode([]byte{0xFF, 'a'})` returns the bytes unchanged without error. Data that starts with a UTF-16 or UTF-32 BOM is an error (`wrong BOM in data: [255 254], expected: [239 187 191]`).
- **Byte orders other than little and big endian.** UTF-16 functions accept `binary.NativeEndian` for BOM-less data, but data with a BOM returns `invalid binary.ByteOrder: NativeEndian`. UTF-32 functions treat every byte order other than `binary.BigEndian` as little endian.
- **UTF-16 and UTF-32 handle bad input differently.** `EncodeUTF16` returns an error for invalid UTF-8, `EncodeUTF32` writes U+FFFD. `DecodeUTF16` rejects an odd length, `DecodeUTF32` decodes trailing bytes as U+FFFD.
- **Legacy charsets don't replace unknown runes.** `MustGetEncoding("latin1").Encode([]byte("€"))` returns the error `encoding: rune not supported by encoding.`, while `windows-1252` encodes it as `0x80`.
- **IANA, not WHATWG labels.** `GetEncoding` follows the IANA registry like MIME does: `latin1`, `ISO-8859-1`, and `US-ASCII` stay ISO 8859-1 and US-ASCII, while browsers decode them as windows-1252. Browser-only labels like `cp1252` are not found. For HTML documents use [`golang.org/x/net/html/charset`](https://pkg.go.dev/golang.org/x/net/html/charset).
- **`AutoDecode` needs distinctive key words.** UTF-8 decoding never fails, and ASCII key words match the same in all ASCII-compatible encodings. With the key word `Rechnung` and the encodings UTF-8 and Windows 1252, the Windows 1252 data `Rechnung f\xfcr` comes back unchanged and named `UTF-8`. Use key words with non-ASCII characters like `für` or `€`.
- **`FF FE 00 00` is detected as UTF-32LE**, see [UTF-16LE vs. UTF-32LE detection](#utf-16le-vs-utf-32le-detection).

## Examples

Detect a BOM and decode the rest:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/charset"
)

func main() {
	data := []byte{0xFF, 0xFE, 'h', 0, 'i', 0} // UTF-16LE with BOM

	bom, rest := charset.SplitBOM(data)
	fmt.Println(bom)

	text, err := bom.DecodeString(rest)
	if err != nil {
		panic(err)
	}
	fmt.Println(text)

	// Output:
	// UTF-16LE
	// hi
}
```

Decode text with the charset label of a MIME header, and write UTF-16LE with a BOM, as Excel expects for CSV files:

```go
package main

import (
	"encoding/binary"
	"fmt"

	"github.com/domonda/go-types/charset"
)

func main() {
	// Content-Type: text/plain; charset=windows-1252
	enc, err := charset.GetEncoding("windows-1252")
	if err != nil {
		panic(err)
	}
	text, err := enc.Decode([]byte("f\xfcr 10 \x80"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s: %s\n", enc.Name(), text)

	utf16 := charset.UTF16Encoding(binary.LittleEndian)
	encoded, err := utf16.Encode([]byte("hi"))
	if err != nil {
		panic(err)
	}
	file := append([]byte(utf16.BOM()), encoded...)
	fmt.Printf("% X\n", file)

	// Output:
	// Windows 1252: für 10 €
	// FF FE 68 00 69 00
}
```

Guess the encoding of BOM-less data with key words:

```go
package main

import (
	"fmt"

	"github.com/domonda/go-types/charset"
)

func main() {
	data := []byte("Rechnung f\xfcr 100 \x80") // Windows 1252

	candidates := []charset.Encoding{
		charset.UTF8Encoding(),
		charset.MustGetEncoding("latin1"),
		charset.MustGetEncoding("windows-1252"),
	}
	text, name, err := charset.AutoDecode(data, candidates, []string{"für", "€"})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s: %s\n", name, text)

	// Output:
	// Windows 1252: Rechnung für 100 €
}
```

## Subpackage findencoding

[`findencoding`](findencoding/README.md) prints a file decoded with every legacy charset, to find its encoding by eye while debugging.

## Related

- [`email`](../email/README.md): decodes RFC 2047 encoded headers with `golang.org/x/net/html/charset`, not with this package.
- [`golang.org/x/text/encoding`](https://pkg.go.dev/golang.org/x/text/encoding): the charset implementations behind `GetEncoding`.
