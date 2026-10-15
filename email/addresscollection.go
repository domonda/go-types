package email

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"unicode/utf8"

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
// joined with a following entry that has angle brackets. A comma in an
// RFC 2047 encoded-word splits like any other, because net/mail and mail
// servers read `=?utf-8?q?a@b,?= <c@example.com>` as two addresses.
// Empty entries are skipped.
//
// The commas are found in the list as ParseAddressList reads it, without
// the runes that it drops, so that a list that ParseAddressList parses
// without an error has the entries that it parses.
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
	if trimmed == "" || isPlaceholderList(trimmed) {
		return nil
	}
	var parts []listPart
	add := func(begin, end int) {
		entry := strutil.TrimSpace(trimmed[begin:end])
		if entry != "" {
			begin += strings.Index(trimmed[begin:end], entry)
			parts = append(parts, listPart{begin, begin + len(entry)})
		}
	}
	// The separators are found in the list as ParseAddressList reads it.
	// sanitizeAddr keeps every comma, so the n-th comma of sanitized
	// is the n-th comma of trimmed.
	sanitized := sanitizeAddr(trimmed)
	unchanged := sanitized == trimmed
	start := 0 // of the entry in trimmed
	for {
		i := nextListSeparator(sanitized)
		if i < 0 {
			add(start, len(trimmed))
			break
		}
		end := start + i + 1
		if !unchanged {
			end = start
			for n := strings.Count(sanitized[:i], ",") + 1; n > 0; n-- {
				end += strings.IndexByte(trimmed[end:], ',') + 1
			}
		}
		add(start, end-1)
		start = end
		sanitized = sanitized[i+1:]
	}
	return joinDisplayNameParts(trimmed, parts)
}

// nextListSeparator returns the index of the first comma in list that
// separates two entries of an address list, outside quotes, angle
// brackets and nested comments, with backslash escapes in quotes and
// comments, or -1 if there is none. SplitAddressList splits at these
// commas, and ParseAddressList skips to them after an invalid entry,
// so that both find the same entries. An RFC 2047 encoded-word of a
// display name holds none of these characters, see phraseEncodedWord,
// so the address pattern and nextListSeparator read it alike.
// list starts at the start of an entry and is sanitized by sanitizeAddr.
func nextListSeparator(list string) int {
	var state listState
	for i := 0; i < len(list); i++ {
		// All delimiters are ASCII, so iterating bytes is safe for UTF-8
		if state.read(list[i]) {
			return i
		}
	}
	return -1
}

// listState is the state of reading an address list byte by byte
// like nextListSeparator. The zero value is the state at the start
// of the list and after every separator.
type listState struct {
	quoted  bool
	escaped bool
	comment int
	angle   int
}

// read reads the next byte c of an address list and reports whether
// it is a comma that separates two entries.
func (s *listState) read(c byte) (separator bool) {
	switch {
	case s.escaped:
		s.escaped = false
	case s.quoted:
		switch c {
		case '\\':
			s.escaped = true
		case '"':
			s.quoted = false
		}
	case s.comment > 0:
		switch c {
		case '\\':
			s.escaped = true
		case '(':
			s.comment++
		case ')':
			s.comment--
		}
	case c == '"':
		s.quoted = true
	case c == '(':
		s.comment++
	case c == '<':
		s.angle++
	case c == '>' && s.angle > 0:
		s.angle--
	case c == ',' && s.angle == 0:
		return true
	}
	return false
}

// openAtEnd returns what is still open at the end of entry, read like
// nextListSeparator reads an address list: "quote", "comment" or
// "angle bracket", or "" if nothing is open, so that a comma after entry
// separates it from the next entry for SplitAddressList.
func openAtEnd(entry string) string {
	var state listState
	for i := 0; i < len(entry); i++ {
		state.read(entry[i])
	}
	switch {
	case state.quoted:
		return "quote"
	case state.comment > 0:
		return "comment"
	case state.angle > 0:
		return "angle bracket"
	}
	return ""
}

// listEntryEnd returns the index of the comma where SplitAddressList ends
// the first entry of the address list, or -1 if the entry holds the rest
// of the list, and the index of the last part of the entry, if it has more
// than one part, else 0. list starts at the start of an entry and is
// sanitized by sanitizeAddr. A first part up to a comma without an @ or
// angle brackets is joined with the following parts without them up to
// one with angle brackets, see joinDisplayNameParts, so the entry ends
// after that part, which is its last part, like `John <john@example.com>`
// of `Doe, John <john@example.com>`. An empty first part is an entry of
// its own here, unlike for SplitAddressList, which skips it:
// ParseAddressList reads the empty entry before a doubled comma as invalid
// entry and continues after it.
func listEntryEnd(list string) (lastPart, end int) {
	first := nextListSeparator(list)
	if first < 0 || strings.ContainsAny(list[:first], "@<") || strutil.TrimSpace(list[:first]) == "" {
		return 0, first
	}
	for start := first + 1; ; {
		i := nextListSeparator(list[start:])
		part := list[start:]
		if i >= 0 {
			part = part[:i]
		}
		switch {
		case strings.Contains(part, "<"):
			// The display name parts end at this part
			if i < 0 {
				return start, -1
			}
			return start, start + i
		case strings.Contains(part, "@") || i < 0:
			// No display name, the first part is an entry of its own
			return 0, first
		}
		start += i + 1
	}
}

// isSingleListEntry reports whether SplitAddressList reads the address
// list entry as one entry: entry has no separator, or only the commas of
// an unquoted display name, which SplitAddressList joins with the address
// after them.
func isSingleListEntry(entry string) bool {
	if nextListSeparator(entry) < 0 {
		return true
	}
	entries := SplitAddressList(entry)
	return len(entries) == 1 && string(entries[0]) == strutil.TrimSpace(entry)
}

// isPlaceholderList reports whether list is only a placeholder for hidden
// recipients like "undisclosed-recipients:;", read like ParseAddressList
// reads it: sanitized and without trailing commas, so that SplitAddressList
// returns no entries for every list that ParseAddressList parses as empty.
// A placeholder has no @, so a list with one is not sanitized to check it.
func isPlaceholderList(list string) bool {
	if strings.Contains(list, "@") {
		return false
	}
	list = strings.TrimRight(sanitizeAddr(list), ", ")
	return placeholderListRegexp.MatchString(strings.ToLower(list))
}

// hasUnsafeRune reports whether text has a rune that the address
// collections refuse in Validate and in the methods that write their text
// as stored, because the address parser ignores it while a message header
// or a database written from the stored text would not: a control
// character other than the tab, like a line break that injects a header
// or a NUL that PostgreSQL rejects, a bidi embedding, override or isolate
// that shows the text in another order than the parser reads it,
// or invalid UTF-8. All other runes are written as stored, like the
// no-break space, the ideographic space, the zero width non-joiner and
// joiner and the soft hyphen in names, and the left-to-right and
// right-to-left marks.
func hasUnsafeRune(text string) bool {
	for i := 0; i < len(text); {
		c := text[i]
		if c < utf8.RuneSelf {
			if (c < ' ' && c != '\t') || c == 0x7f {
				return true
			}
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(text[i:])
		switch {
		case r == utf8.RuneError && size == 1, // invalid UTF-8
			r <= 0x9f,                      // C1 control character
			r >= '\u202a' && r <= '\u202e', // bidi embedding or override
			r >= '\u2066' && r <= '\u2069': // bidi isolate
			return true
		}
		i += size
	}
	return false
}

// checkSafeRunes returns an error naming text as name if text, trimmed
// like by strutil.TrimSpace, has a rune that hasUnsafeRune reports.
// The surrounding runes that strutil.TrimSpace removes are not checked,
// because the Value methods trim them and don't write them.
// The Validate methods of the address collections call it, so that their
// Value methods and AddressSet.MarshalText, which write the text as
// stored, write every text that Validate accepts and nothing else:
// a display name with a line break must not pass Validate and inject
// a header into a message built from the stored text.
func checkSafeRunes(text, name string) error {
	text = strutil.TrimSpace(text)
	if !hasUnsafeRune(text) {
		return nil
	}
	return errs.Errorf("%s %q has a control character, a bidi embedding, override or isolate, or invalid UTF-8", name, truncateForError(text))
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
