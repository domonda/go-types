package strutil

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const tld = "(?:com|net|org|edu|gov|mil|de|at|eu|ch|es|nl|uk|cz|dk|fi)"

var (
	httpDomainRegex  = regexp.MustCompile(`^http(?:s)?\:\/\/([a-z][a-z0-9\-\.]*\.` + tld + `)(?:\/.*)?$`)
	wwwDomainRegex   = regexp.MustCompile(`^(www\.[a-z][a-z0-9\-]*\.` + tld + `)(?:\/.*)?$`)
	emailDomainRegex = regexp.MustCompile(`^[a-z][a-z0-9\.\-_\+]*@([a-z][a-z0-9\-]*\.` + tld + `)$`)
)

// ParseDomainName extracts and returns a domain name from word.
// It recognizes domains in HTTP URLs, www prefixed domains, and email addresses.
// Returns an empty string if no domain is found.
func ParseDomainName(word string) string {
	word = strings.ToLower(word)

	matches := httpDomainRegex.FindStringSubmatch(word)
	if len(matches) == 2 {
		return matches[1]
	}

	matches = wwwDomainRegex.FindStringSubmatch(word)
	if len(matches) == 2 {
		return matches[1]
	}

	matches = emailDomainRegex.FindStringSubmatch(word)
	if len(matches) == 2 {
		return matches[1]
	}

	return ""
}

// ParseDomainNameIndex extracts a domain name from word and returns both the
// lowercase domain and the byte indices [start, end) where the domain was found in word.
// It recognizes domains in HTTP URLs, www prefixed domains, and email addresses.
// Returns empty string and nil indices if no domain is found.
func ParseDomainNameIndex(word string) (domain string, indices []int) {
	lower := strings.ToLower(word)

	indices = httpDomainRegex.FindStringSubmatchIndex(lower)
	if len(indices) == 4 {
		return lower[indices[2]:indices[3]], []int{wordIndex(word, indices[2]), wordIndex(word, indices[3])}
	}

	indices = wwwDomainRegex.FindStringSubmatchIndex(lower)
	if len(indices) == 4 {
		return lower[indices[2]:indices[3]], []int{wordIndex(word, indices[2]), wordIndex(word, indices[3])}
	}

	indices = emailDomainRegex.FindStringSubmatchIndex(lower)
	if len(indices) == 4 {
		return lower[indices[2]:indices[3]], []int{wordIndex(word, indices[2]), wordIndex(word, indices[3])}
	}

	return "", nil
}

// wordIndex converts the byte index lowerIndex of strings.ToLower(word)
// to the byte index of the same rune in word.
// Lowercasing can change the byte length of a rune, for example
// 'İ' (U+0130) and the Kelvin sign (U+212A) become 'i' and 'k',
// and an invalid UTF-8 byte becomes U+FFFD.
func wordIndex(word string, lowerIndex int) int {
	lowerPos := 0
	for i, r := range word {
		if lowerPos >= lowerIndex {
			return i
		}
		lowerPos += utf8.RuneLen(unicode.ToLower(r))
	}
	return len(word)
}
