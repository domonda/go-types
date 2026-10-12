// Package charset converts text between UTF-8 and other character encodings
// and handles Unicode Byte Order Marks (BOM).
//
// The package provides:
//   - [Encoding]: an interface to convert text to and from UTF-8,
//     implemented by [UTF8Encoding], [UTF16Encoding], [UTF32Encoding],
//     and the encodings returned by [GetEncoding]
//   - [GetEncoding]: looks up an encoding by display name ("Windows 1252")
//     or by IANA charset label as used in HTTP and MIME headers
//     ("windows-1252", "latin1", "Shift_JIS")
//   - [BOM]: detection with [BOMOfBytes] and [SplitBOM], removal with
//     [TrimBOM], and decoding with [DecodeWithBOM] and [BOM.Decode]
//   - [DecodeUTF16], [EncodeUTF16], [DecodeUTF32], and [EncodeUTF32]
//     for byte slices of a known byte order
//   - [AutoDecode]: guesses the encoding of data without BOM
//     by counting key words in the decoded text
//
// Except for the IANA "UTF-16" encoding, the encoders don't write a BOM,
// prepend the result of [Encoding.BOM] if the receiver needs one.
// The UTF decoders strip a leading BOM that matches the expected encoding
// and return an error for any other BOM.
//
// The subpackage findencoding prints a file decoded with all legacy
// charsets to find its encoding by eye while debugging.
//
// See the README.md file of the package for the complete API reference
// and the known pitfalls.
package charset
