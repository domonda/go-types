package email

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"
	"strings"

	"github.com/domonda/go-errs"

	"github.com/domonda/go-types/strutil"
)

// SplitAddressList splits an address list at the commas between its
// entries and returns the entries unchanged, only with the surrounding
// whitespace trimmed: they are not parsed, normalized or validated.
// Commas inside double quoted strings, angle brackets and comments
// in parentheses don't split, so `"Doe, John" <john@example.com>` is one
// entry. Neither do the commas of an unquoted display name, which are
// followed by an address in angle brackets, so `Doe, John <john@example.com>`
// is one entry too, like for ParseAddressList: entries without an @ are
// joined with a following entry that has angle brackets.
// Empty entries are skipped.
//
// An empty or whitespace only list and the placeholders for hidden
// recipients like "undisclosed-recipients:;" return nil, like for
// ParseAddressList. SplitAddressList never fails: after an unbalanced
// quote, bracket or parenthesis the rest of the list is one entry.
// Group syntax like "Team: a@example.com;" is not supported,
// like by ParseAddressList.
//
// Use ParseAddressList or AddressList.SplitNormalized to get the
// entries parsed and normalized.
func SplitAddressList(list string) []Address {
	trimmed := strutil.TrimSpace(list)
	if trimmed == "" || placeholderListRegexp.MatchString(strings.ToLower(trimmed)) {
		return nil
	}
	var (
		parts   []listPart
		start   int
		quoted  bool
		escaped bool
		comment int
		angle   int
	)
	add := func(begin, end int) {
		entry := strutil.TrimSpace(trimmed[begin:end])
		if entry != "" {
			begin += strings.Index(trimmed[begin:end], entry)
			parts = append(parts, listPart{begin, begin + len(entry)})
		}
	}
	for i := 0; i < len(trimmed); i++ {
		// All delimiters are ASCII, so iterating bytes is safe for UTF-8
		c := trimmed[i]
		switch {
		case escaped:
			escaped = false
		case quoted:
			switch c {
			case '\\':
				escaped = true
			case '"':
				quoted = false
			}
		case comment > 0:
			switch c {
			case '\\':
				escaped = true
			case '(':
				comment++
			case ')':
				comment--
			}
		case c == '"':
			quoted = true
		case c == '(':
			comment++
		case c == '<':
			angle++
		case c == '>' && angle > 0:
			angle--
		case c == ',' && angle == 0:
			add(start, i)
			start = i + 1
		}
	}
	add(start, len(trimmed))
	return joinDisplayNameParts(trimmed, parts)
}

// checkSanitized returns an error naming text as name if sanitizeAddr
// changes text, which is trimmed like by strutil.TrimSpace, because text
// has whitespace other than the space, like a line break, a control or
// format character, like a NUL or a bidi override, or invalid UTF-8.
// Validate parses the sanitized text, which doesn't have these runes,
// while the Value methods of the address collections and
// AddressSet.MarshalText write the text as stored. They call
// checkSanitized so that they only write text that was validated as it
// is written, else a valid display name could carry a line break that
// injects a header into a message, or a NUL that PostgreSQL rejects.
func checkSanitized(text, name string) error {
	if sanitizeAddr(text) == text {
		return nil
	}
	return errs.Errorf("%s %q has whitespace other than spaces, a control or format character or invalid UTF-8, which the address parser ignores", name, truncateForError(text))
}

// listPart is the byte range of an entry of an address list,
// without its surrounding whitespace.
type listPart struct{ begin, end int }

// joinDisplayNameParts returns the parts of list split by SplitAddressList
// at all top level commas, with every run of parts without an @ or angle
// brackets joined with the following part if that has angle brackets,
// because their commas are part of an unquoted display name, like in
// `Doe, Jr., John <john@example.com>`. Other parts are kept separate.
// A joined entry is the text of list from the first to the last part.
func joinDisplayNameParts(list string, parts []listPart) []Address {
	if len(parts) == 0 {
		return nil
	}
	text := func(p listPart) string { return list[p.begin:p.end] }
	entries := make([]Address, 0, len(parts))
	for i := 0; i < len(parts); i++ {
		// next is the first part from i on with an @ or angle brackets
		next := i
		for next < len(parts) && !strings.ContainsAny(text(parts[next]), "@<") {
			next++
		}
		if next > i && next < len(parts) && strings.Contains(text(parts[next]), "<") {
			entries = append(entries, Address(list[parts[i].begin:parts[next].end]))
		} else {
			// Every part up to and including next is an entry,
			// so the parts are scanned only once
			for _, part := range parts[i:min(next+1, len(parts))] {
				entries = append(entries, Address(text(part)))
			}
		}
		i = next
	}
	return entries
}

// addressCollection is the text of an address collection
// as read by readAddressCollection. The zero value is null.
type addressCollection struct {
	// list is set for a plain address list and the content of a JSON
	// string, both kept as given, if entries is nil.
	list string
	// entries are the elements of a JSON array or a PostgreSQL array
	// or the keys of a JSON object, a non nil empty slice for an empty
	// array or object.
	entries []string
}

// readAddressCollection reads the text of an address collection for the
// Scan and UnmarshalJSON methods of AddressSet, AddressList and
// NullableAddressList, so that both accept every shape the collection
// types write or that is found in text, text[], json and jsonb columns.
// On the trimmed text, in this order:
//
//   - "", whitespace only, "null": null
//   - "[" and valid JSON: a JSON array of strings, elements kept exactly
//   - "{" and valid JSON: a JSON object, keys kept exactly in sorted order
//   - "{...}" otherwise: a PostgreSQL array literal, elements trimmed
//     and empty ones skipped
//   - '"' and valid JSON: a JSON string, its content as address list
//   - anything else: an address list, the text kept as given
//
// "{}" is a valid JSON object and the empty PostgreSQL array, both are
// the empty collection. A PostgreSQL array with elements is never valid
// JSON, and `"John Doe" <john@example.com>` is no valid JSON string.
// A JSON array with other elements than strings is an error, like a
// malformed PostgreSQL array literal. typeName is used only in errors.
//
// text is a string or a []byte, so that neither the string of Scan nor
// the []byte of UnmarshalJSON is copied into the other type just to be
// read; the JSON decoder reads a []byte.
func readAddressCollection[T string | []byte](text T, typeName string) (addressCollection, error) {
	trimmed := trimSpace(text)
	switch {
	case len(trimmed) == 0 || string(trimmed) == "null":
		return addressCollection{}, nil

	case trimmed[0] == '[':
		entries := []string{}
		if isJSON, err := unmarshalIfJSON([]byte(trimmed), &entries); isJSON {
			if err != nil {
				return addressCollection{}, errs.Errorf("can't read JSON array as email.%s because of: %w", typeName, err)
			}
			return addressCollection{entries: entries}, nil
		}

	case trimmed[0] == '{':
		if canBeJSONObject(trimmed) {
			var object map[string]json.RawMessage
			if isJSON, err := unmarshalIfJSON([]byte(trimmed), &object); isJSON {
				if err != nil {
					return addressCollection{}, errs.Errorf("can't read JSON object as email.%s because of: %w", typeName, err)
				}
				entries := make([]string, 0, len(object))
				for key := range object {
					entries = append(entries, key)
				}
				slices.Sort(entries)
				return addressCollection{entries: entries}, nil
			}
		}
		if trimmed[len(trimmed)-1] == '}' {
			array, err := scanAddressArray(string(trimmed), typeName)
			if err != nil {
				return addressCollection{}, err
			}
			entries := make([]string, 0, len(array))
			for _, addr := range array {
				// scanAddressArray trims, so an element that was only
				// padding is empty here and is not an entry
				if addr != "" {
					entries = append(entries, addr)
				}
			}
			return addressCollection{entries: entries}, nil
		}

	case trimmed[0] == '"' && trimmed[len(trimmed)-1] == '"':
		var list string
		if isJSON, err := unmarshalIfJSON([]byte(trimmed), &list); isJSON {
			if err != nil {
				return addressCollection{}, errs.Errorf("can't read JSON string as email.%s because of: %w", typeName, err)
			}
			return addressCollection{list: list}, nil
		}
	}
	return addressCollection{list: string(text)}, nil
}

// trimSpace returns text trimmed like by strutil.TrimSpace.
func trimSpace[T string | []byte](text T) T {
	if b, ok := any(text).([]byte); ok {
		return T(strutil.TrimSpaceBytes(b))
	}
	return T(strutil.TrimSpace(string(text)))
}

// unmarshalIfJSON unmarshals data into v and reports whether data is
// valid JSON, so that readAddressCollection reads text that is no JSON
// as another shape of an address collection without an error, but returns
// the error for JSON of another type than v, like [1] for a []string.
// json.Unmarshal checks the whole data before it decodes and reports a
// syntax error instead of any other error, so its *json.SyntaxError tells
// that data is no JSON without another pass over data with json.Valid.
func unmarshalIfJSON(data []byte, v any) (isJSON bool, err error) {
	err = json.Unmarshal(data, v)
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return false, nil
	}
	return true, err
}

// canBeJSONObject reports whether text, which starts with '{', can be
// a JSON object, because its first byte after JSON whitespace is the quote
// of a key or the closing brace. A PostgreSQL array literal like
// {a@example.com} can't, so it is not decoded as JSON first.
func canBeJSONObject[T string | []byte](text T) bool {
	for i := 1; i < len(text); i++ {
		switch text[i] {
		case ' ', '\t', '\n', '\r':
			continue
		case '"', '}':
			return true
		}
		return false
	}
	return false
}

// scanAddressCollection reads the SQL value of an address collection
// with readAddressCollection. SQL NULL is null, any other type than
// string and []byte is an error that names typeName.
func scanAddressCollection(value any, typeName string) (addressCollection, error) {
	switch s := value.(type) {
	case nil:
		return addressCollection{}, nil
	case string:
		return readAddressCollection(s, typeName)
	case []byte:
		return readAddressCollection(s, typeName)
	}
	return addressCollection{}, errs.Errorf("can't scan SQL value of type %T as email.%s", value, typeName)
}

// unmarshalJSONAddressCollection reads the JSON value data of an address
// collection with readAddressCollection. A JSON string, array or object
// is accepted, any other JSON value is an error that names typeName.
func unmarshalJSONAddressCollection(data []byte, typeName string) (addressCollection, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || (trimmed[0] != '"' && trimmed[0] != '[' && trimmed[0] != '{') {
		return addressCollection{}, errs.Errorf("can't unmarshal JSON %s as email.%s", data, typeName)
	}
	return readAddressCollection(trimmed, typeName)
}
