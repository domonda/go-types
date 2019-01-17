package charset

import (
	"sort"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

func EncodingNames() []string {
	var names []string
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

var sharedEncodings = map[string]encoding.Encoding{
	"ISO-8859-6E": charmap.ISO8859_6,
	"ISO-8859-6I": charmap.ISO8859_6,
	"ISO-8859-8E": charmap.ISO8859_8,
	"ISO-8859-8I": charmap.ISO8859_8,
}

func findEncoding(name string) (e encoding.Encoding, foundName string) {
	nameUpper := strings.ToUpper(name)
	for _, e = range charmap.All {
		if cm, ok := e.(*charmap.Charmap); ok {
			if strings.ToUpper(cm.String()) == nameUpper {
				return e, cm.String()
			}
			if _, other := cm.ID(); other == name {
				return e, name
			}
		}
	}
	return sharedEncodings[name], name
}
