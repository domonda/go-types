# charset

Character-encoding utilities focused on the cases that show up when ingesting external text: detect a Unicode Byte Order Mark, encode/decode UTF-8, UTF-16, and UTF-32 to/from UTF-8 byte slices, and convert via legacy encodings from `golang.org/x/text/encoding`.

```
import "github.com/domonda/go-types/charset"
```

## Encoding interface

```go
type Encoding interface {
    Encode(utf8Str []byte) ([]byte, error) // UTF-8 → encoded
    Decode(encoded  []byte) ([]byte, error) // encoded → UTF-8
    Name() string
    String() string
    BOM() BOM
}
```

`Encoding` implementations are safe for concurrent use; the package wraps `golang.org/x/text/encoding` encoders/decoders with internal mutexes.

| Constructor                                       | Returns                                     |
|---------------------------------------------------|---------------------------------------------|
| `UTF8Encoding()`                                  | Pass-through `Encoding` for UTF-8.          |
| `UTF16Encoding(binary.LittleEndian \| BigEndian)` | UTF-16 (LE or BE) `Encoding`.               |
| `UTF32Encoding(binary.LittleEndian \| BigEndian)` | UTF-32 (LE or BE) `Encoding`.               |
| `findencoding.FindEncoding(name string)`          | Resolve any IANA / x/text encoding by name. |

## BOM

`BOM` is a string holding the literal BOM bytes. Predefined values:

| Value        | Bytes         |
|--------------|---------------|
| `NoBOM`      | empty         |
| `BOMUTF8`    | `EF BB BF`    |
| `BOMUTF16BE` | `FE FF`       |
| `BOMUTF16LE` | `FF FE`       |
| `BOMUTF32BE` | `00 00 FE FF` |
| `BOMUTF32LE` | `FF FE 00 00` |

| Function / Method                             | Description                                        |
|-----------------------------------------------|----------------------------------------------------|
| `BOMOfBytes(b)` / `BOMOfString(s)`            | Detect a leading BOM (`NoBOM` if none).            |
| `TrimBOM(b, bom)`                             | Drop the UTF-8 BOM prefix if present.              |
| `SplitBOM(b)`                                 | Return `(bom, rest)`.                              |
| `DecodeWithBOM(b)` / `DecodeStringWithBOM(b)` | Split BOM then decode to UTF-8.                    |
| `bom.Encoding()`                              | The `Encoding` matching a BOM value.               |
| `bom.Endian()`                                | `binary.LittleEndian`/`BigEndian` for UTF-16/32 BOMs. |
| `bom.Decode(data)` / `bom.DecodeString(data)` | Decode using the BOM's encoding.                   |

## UTF-16 / UTF-32 helpers

Lower-level functions if you don't need the `Encoding` interface:

```go
charset.DecodeUTF16(b, byteOrder)       // []byte UTF-16 → []byte UTF-8
charset.DecodeUTF16String(b, byteOrder) // []byte UTF-16 → string UTF-8
charset.EncodeUTF16(b, byteOrder)       // []byte UTF-8  → []byte UTF-16
charset.DecodeUTF32(b, byteOrder)
charset.DecodeUTF32String(b, byteOrder)
charset.EncodeUTF32(b, byteOrder)
```

All UTF-16/32 helpers tolerate a leading BOM but error if the BOM disagrees with the requested byte order.

## Example

```go
data := []byte{0xFF, 0xFE, 'h', 0, 'i', 0} // UTF-16LE with BOM

bom, rest := charset.SplitBOM(data)
fmt.Println(bom) // "UTF-16LE"

utf8Bytes, err := bom.Decode(rest)
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(utf8Bytes)) // "hi"
```

## Subcommands

`findencoding/` contains a small CLI for resolving an encoding by name against the `golang.org/x/text/encoding` registry. Useful when probing files of unknown origin.
