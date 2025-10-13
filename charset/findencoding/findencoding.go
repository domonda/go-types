// Package findencoding provides utilities for detecting and testing character encodings
// in text files, particularly useful for debugging encoding issues and finding the correct
// encoding for unknown text data.
//
// The package includes:
// - File encoding detection and testing
// - Byte Order Mark (BOM) detection and handling
// - Multiple encoding format support (UTF-8, Windows-1252, ISO-8859-1, etc.)
// - Debugging utilities for encoding analysis
// - Safe file reading with size limits
//
// This package is particularly useful for:
// - Debugging encoding issues in text files
// - Finding the correct encoding for legacy files
// - Testing how text appears in different character encodings
package findencoding

import (
	"fmt"
	"os"

	"golang.org/x/text/encoding/charmap"

	"github.com/domonda/go-types/charset"
	"github.com/domonda/go-types/strutil"
)

// PrintFileWithAllEncodings reads a file and prints its content using all available
// character encodings, making it easy to identify the correct encoding for unknown text.
// The function handles BOM detection and provides formatted output for each encoding.
// maxBytes limits the number of bytes to read (0 defaults to 1MB).
func PrintFileWithAllEncodings(filename string, maxBytes int) error { //#nosec G304 -- file inclusion OK
	sourceData, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	if maxBytes == 0 {
		maxBytes = 1024 * 1024
	}
	if len(sourceData) > maxBytes {
		sourceData = sourceData[:maxBytes]
	}

	print := func(charset string, str string) {
		fmt.Println("====================")
		fmt.Println(charset)
		fmt.Println("--------------------")
		fmt.Print(str, "\n")
		fmt.Println("====================")
	}

	bom, sourceData := charset.SplitBOM(sourceData)
	if bom != charset.NoBOM {
		fmt.Println("Found BOM:", bom)

		str, err := bom.DecodeString(sourceData)
		if err != nil {
			return err
		}
		print(bom.String(), str)
		return nil
	}

	sourceData = strutil.SanitizeLineEndingsBytes(sourceData)

	for _, enc := range charmap.All {
		if cm, ok := enc.(*charmap.Charmap); ok {
			decoded, err := enc.NewDecoder().Bytes(sourceData)
			if err != nil {
				return err
			}
			if cm.String() == "IBM Code Page 037" {
				continue
			}
			print(cm.String(), string(decoded))
		}
	}

	print("UTF-8", string(sourceData))

	return nil
}
