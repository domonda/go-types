package strutil

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var domainMap = map[string]string{
	"http://domonda.com":             "domonda.com",
	"https://domonda.com":            "domonda.com",
	"http://domonda.com/":            "domonda.com",
	"https://domonda.com/":           "domonda.com",
	"http://domonda.com/impressum/":  "domonda.com",
	"https://domonda.com/impressum/": "domonda.com",
	"https://www.domonda.com/":       "www.domonda.com",
	"https://app.domonda.com/":       "app.domonda.com",

	"www.domonda.com":            "www.domonda.com",
	"www.domonda.com/":           "www.domonda.com",
	"www.domonda.com/impressum/": "www.domonda.com",

	"erik@domonda.com":          "domonda.com",
	"erik.unger@domonda.com":    "domonda.com",
	"erik-unger@domonda.com":    "domonda.com",
	"erik_unger@domonda.com":    "domonda.com",
	"er+ka+domonda@domonda.com": "domonda.com",

	"HTTPS://DOMONDA.COM/IMPRESSUM/": "domonda.com",
	"WWW.DOMONDA.COM/IMPRESSUM/":     "www.domonda.com",
	"ERIK@DOMONDA.COM":               "domonda.com",
	"ERIK@DOMONDA.at":                "domonda.at",

	"office@adwordsagentur.at": "adwordsagentur.at",
}

func Test_ParseDomainName(t *testing.T) {
	for word, domain := range domainMap {
		result := ParseDomainName(word)
		assert.Equal(t, domain, result, "ParseDomainName(%#v)", word)
	}
}

func Test_ParseDomainNameIndex(t *testing.T) {
	for word, domain := range domainMap {
		result, indices := ParseDomainNameIndex(word)
		assert.Equal(t, domain, result, "ParseDomainNameIndex(%#v)", word)
		assert.Len(t, indices, 2, "ParseDomainNameIndex(%#v)", word)
	}
}

// noDomainWords must not be recognized as domain names because only
// the documented forms are: http(s) URLs, www prefixed domains, and
// email addresses, each with one of the supported TLDs.
var noDomainWords = []string{
	"",
	"domonda",
	"domonda.com",              // bare domain without www, http or email
	"http://domonda.xyz",       // TLD not in the supported list
	"ftp://domonda.com",        // only http and https URLs
	"erik@domonda",             // email without TLD
	"@domonda.com",             // email without local part
	"see https://domonda.com/", // must be a single word, not found inside text
}

func Test_ParseDomainName_NoDomain(t *testing.T) {
	for _, word := range noDomainWords {
		assert.Equal(t, "", ParseDomainName(word), "ParseDomainName(%#v)", word)
	}
}

func Test_ParseDomainNameIndex_NoDomain(t *testing.T) {
	for _, word := range noDomainWords {
		domain, indices := ParseDomainNameIndex(word)
		assert.Equal(t, "", domain, "ParseDomainNameIndex(%#v)", word)
		assert.Nil(t, indices, "ParseDomainNameIndex(%#v)", word)
	}
}

func Test_ParseDomainNameIndex_Indices(t *testing.T) {
	// The indices must locate the domain within the passed word
	// so that callers can mark or replace it in the original text.
	for word, domain := range domainMap {
		_, indices := ParseDomainNameIndex(word)
		if assert.Len(t, indices, 2, "ParseDomainNameIndex(%#v)", word) {
			assert.True(t, strings.EqualFold(domain, word[indices[0]:indices[1]]), "ParseDomainNameIndex(%#v) indices %v", word, indices)
		}
	}
}

func Test_ParseDomainNameIndex_CaseChangingRunes(t *testing.T) {
	// The domain is matched in the lowercased word, but strings.ToLower
	// can change the byte length of a rune: 'İ' (U+0130, 2 bytes) and the
	// Kelvin sign (U+212A, 3 bytes) both become a single ASCII byte.
	// The indices must still locate the domain in the passed word,
	// else callers marking or replacing it in the original text
	// would hit the wrong bytes.
	tests := []struct {
		word       string
		wantDomain string
		wantInWord string
	}{
		{word: "İnfo@domonda.com", wantDomain: "domonda.com", wantInWord: "domonda.com"},
		{word: "https://İ.domonda.com/", wantDomain: "i.domonda.com", wantInWord: "İ.domonda.com"},
		{word: "Karl@Kunde.at", wantDomain: "kunde.at", wantInWord: "Kunde.at"},
		{word: "www.Kunde.at/İmpressum", wantDomain: "www.kunde.at", wantInWord: "www.Kunde.at"},
	}
	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			domain, indices := ParseDomainNameIndex(tt.word)
			assert.Equal(t, tt.wantDomain, domain)
			assert.Equal(t, ParseDomainName(tt.word), domain, "same domain as ParseDomainName")
			if assert.Len(t, indices, 2) {
				assert.Equal(t, tt.wantInWord, tt.word[indices[0]:indices[1]])
			}
		})
	}
}
