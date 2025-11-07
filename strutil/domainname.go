package strutil

import (
	"regexp"
	"strings"
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

// ParseDomainNameIndex extracts a domain name from word and returns both the domain
// and the byte indices [start, end] where the domain was found.
// It recognizes domains in HTTP URLs, www prefixed domains, and email addresses.
// Returns empty string and nil indices if no domain is found.
func ParseDomainNameIndex(word string) (domain string, indices []int) {
	word = strings.ToLower(word)

	indices = httpDomainRegex.FindStringSubmatchIndex(word)
	if len(indices) == 4 {
		indices = indices[2:]
		return strings.ToLower(word[indices[0]:indices[1]]), indices
	}

	indices = wwwDomainRegex.FindStringSubmatchIndex(word)
	if len(indices) == 4 {
		indices = indices[2:]
		return strings.ToLower(word[indices[0]:indices[1]]), indices
	}

	indices = emailDomainRegex.FindStringSubmatchIndex(word)
	if len(indices) == 4 {
		indices = indices[2:]
		return strings.ToLower(word[indices[0]:indices[1]]), indices
	}

	return "", nil
}
