package strutil

import (
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
