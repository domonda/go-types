package email

import (
	"bytes"
	"encoding/json"
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
// as read by readAddressCollection.
type addressCollection struct {
	// null is true for an empty or whitespace only text,
	// "null" and the JSON null value.
	null bool
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
func readAddressCollection(text, typeName string) (addressCollection, error) {
	trimmed := strutil.TrimSpace(text)
	switch {
	case trimmed == "" || trimmed == "null":
		return addressCollection{null: true}, nil

	case trimmed[0] == '[' && json.Valid([]byte(trimmed)):
		entries := []string{}
		err := json.Unmarshal([]byte(trimmed), &entries)
		if err != nil {
			return addressCollection{}, errs.Errorf("can't read JSON array as email.%s because of: %w", typeName, err)
		}
		return addressCollection{entries: entries}, nil

	case trimmed[0] == '{' && json.Valid([]byte(trimmed)):
		var object map[string]json.RawMessage
		err := json.Unmarshal([]byte(trimmed), &object)
		if err != nil {
			return addressCollection{}, errs.Errorf("can't read JSON object as email.%s because of: %w", typeName, err)
		}
		entries := make([]string, 0, len(object))
		for key := range object {
			entries = append(entries, key)
		}
		slices.Sort(entries)
		return addressCollection{entries: entries}, nil

	case trimmed[0] == '{' && trimmed[len(trimmed)-1] == '}':
		array, err := scanAddressArray(trimmed, typeName)
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

	case trimmed[0] == '"' && json.Valid([]byte(trimmed)):
		var list string
		err := json.Unmarshal([]byte(trimmed), &list)
		if err != nil {
			return addressCollection{}, errs.Errorf("can't read JSON string as email.%s because of: %w", typeName, err)
		}
		return addressCollection{list: list, null: list == ""}, nil
	}
	return addressCollection{list: text}, nil
}

// unmarshalJSONAddressCollection reads the JSON value data of an address
// collection with readAddressCollection. A JSON string, array or object
// is accepted, any other JSON value is an error that names typeName.
func unmarshalJSONAddressCollection(data []byte, typeName string) (addressCollection, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || (trimmed[0] != '"' && trimmed[0] != '[' && trimmed[0] != '{') {
		return addressCollection{}, errs.Errorf("can't unmarshal JSON %s as email.%s", data, typeName)
	}
	return readAddressCollection(string(trimmed), typeName)
}
