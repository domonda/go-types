package strutil

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/ungerik/go-command"
	fs "github.com/ungerik/go-fs"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

var sharedEncodings = map[string]encoding.Encoding{
	"ISO-8859-6E": charmap.ISO8859_6,
	"ISO-8859-6I": charmap.ISO8859_6,
	"ISO-8859-8E": charmap.ISO8859_8,
	"ISO-8859-8I": charmap.ISO8859_8,
}

func FindEncoding(name string) encoding.Encoding {
	nameUpper := strings.ToUpper(name)
	for _, enc := range charmap.All {
		if cm, ok := enc.(*charmap.Charmap); ok {
			if strings.ToUpper(cm.String()) == nameUpper {
				return enc
			}
			if _, other := cm.ID(); other == name {
				return enc
			}
		}
	}
	return sharedEncodings[name]
}

func AllEncodingNames() (names []string) {
	for _, enc := range charmap.All {
		if cm, ok := enc.(*charmap.Charmap); ok {
			names = append(names, cm.String())
		}
	}
	for name := range sharedEncodings {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

var DebugPrintFileWithAllEncodingsArgs struct {
	command.ArgsDef

	File fs.File `arg:"file"`
}

func DebugPrintFileWithAllEncodings(file fs.File) error {
	sourceData, err := file.ReadAll()
	if err != nil {
		return err
	}

	print := func(charset string, str string) {
		fmt.Println("====================")
		fmt.Println(charset)
		fmt.Println("--------------------")
		fmt.Print(str, "\n")
		fmt.Println("====================")
	}

	bom, sourceData := SplitBOM(sourceData)
	if bom != NoBOM {
		fmt.Println("Found BOM:", bom)

		str, err := bom.DecodeRemaining(sourceData)
		if err != nil {
			return err
		}
		print(string(bom), str)
		return nil
	}

	sourceData = SanitizeLineEndingsBytes(sourceData)

	for _, enc := range charmap.All {
		if cm, ok := enc.(*charmap.Charmap); ok {
			decoded, err := enc.NewDecoder().Bytes(sourceData)
			if err != nil {
				return err
			}
			print(cm.String(), string(decoded))
		}
	}

	print("UTF-8", string(sourceData))

	return nil
}

// SanitizeLineEndings converts all line endings to just '\n'
func SanitizeLineEndings(text string) string {
	// var (
	// 	needsCopy = false
	// 	buf bytes.Buffer
	// 	lastByte byte
	// )

	// for i, b := range text {

	// TODO optimized version

	// 	lastByte = b
	// }

	// if needsCopy {
	// 	return buf.Bytes()
	// }
	// return text

	text = strings.Replace(text, "\r\n", "\n", -1)
	text = strings.Replace(text, "\n\r", "\n", -1)
	text = strings.Replace(text, "\r", "\n", -1)

	return text
}

// SanitizeLineEndingsBytes converts all line endings to just '\n'
func SanitizeLineEndingsBytes(text []byte) []byte {
	// var (
	// 	needsCopy = false
	// 	buf bytes.Buffer
	// 	lastByte byte
	// )

	// for i, b := range text {

	// TODO optimized version

	// 	lastByte = b
	// }

	// if needsCopy {
	// 	return buf.Bytes()
	// }
	// return text

	text = bytes.Replace(text, []byte{'\r', '\n'}, []byte{'\n'}, -1)
	text = bytes.Replace(text, []byte{'\n', '\r'}, []byte{'\n'}, -1)
	text = bytes.Replace(text, []byte{'\r'}, []byte{'\n'}, -1)

	return text
}
