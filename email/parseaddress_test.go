package email

import (
	"fmt"
	"net/mail"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	validEmailAddresses = map[string]*mail.Address{
		`<erik@domonda.com>`:                                            {Name: "", Address: "erik@domonda.com"},
		`"Unger, Erik" <u.erik@domonda.com>`:                            {Name: "Unger, Erik", Address: "u.erik@domonda.com"},
		`"Unger, Erik"` + "\t<u.erik@domonda.com>":                      {Name: "Unger, Erik", Address: "u.erik@domonda.com"},
		`Erik Unger <erik@domonda.com>`:                                 {Name: "Erik Unger", Address: "erik@domonda.com"},
		"Erik\tUnger <erik@domonda.com>":                                {Name: "Erik Unger", Address: "erik@domonda.com"}, // Replace tabs in name with spaces
		`Erik Unger    <erik@domonda.com>`:                              {Name: "Erik Unger", Address: "erik@domonda.com"},
		`Erik Unger <Erik.Unger@domonda.com>`:                           {Name: "Erik Unger", Address: "erik.unger@domonda.com"},
		`"Erik Unger" <erik@domonda.com>`:                               {Name: "Erik Unger", Address: "erik@domonda.com"},
		`"Erik Unger" <"Erik.Unger"@domonda.com>`:                       {Name: "Erik Unger", Address: "erik.unger@domonda.com"},
		`" Erik Unger " <erik@domonda.com>`:                             {Name: "Erik Unger", Address: "erik@domonda.com"},
		`@Erik <erik@domonda.com>`:                                      {Name: "@Erik", Address: "erik@domonda.com"},
		`erik.unger@domonda.com <erik@domonda.com>`:                     {Name: "erik.unger@domonda.com", Address: "erik@domonda.com"}, // Use "erik.unger" in name part vs "erik" in address part to test picking up the right part
		`"Erik Unger-Phd </Domonda-IT>" <Erik.Unger-Phd@domonda-it.at>`: {Name: "Erik Unger-Phd </Domonda-IT>", Address: "erik.unger-phd@domonda-it.at"},
		`refill@example24.de`:                                           {Name: "", Address: "refill@example24.de"},
		`x@mail.example.com`:                                            {Name: "", Address: "x@mail.example.com"},
		`x@mail-example.com`:                                            {Name: "", Address: "x@mail-example.com"},
		`er+bill@mail-billwerk.co.uk`:                                   {Name: "", Address: "er+bill@mail-billwerk.co.uk"},
		`some_underscore@msn.com`:                                       {Name: "", Address: "some_underscore@msn.com"},
		`nasa@7examples.com`:                                            {Name: "", Address: "nasa@7examples.com"},
		`a@we-work.com`:                                                 {Name: "", Address: "a@we-work.com"},
		`customerinfo@email.spammers.com`:                               {Name: "", Address: "customerinfo@email.spammers.com"},
		`Domonda < er+vk+baurauslagen+wirklich@domonda.com>`:            {Name: "Domonda", Address: "er+vk+baurauslagen+wirklich@domonda.com"},
		`Domonda < er+vk+baurauslagen+wirklich@domonda.com >`:           {Name: "Domonda", Address: "er+vk+baurauslagen+wirklich@domonda.com"},
		`_underscore@example.com`:                                       {Name: "", Address: "_underscore@example.com"},

		// Not standard conform, but we still have to be able to parse them:
		`"scanner@" <"example.at scanner"@example.at>`:     {Name: "scanner@", Address: "example.at.scanner@example.at"},
		`"Unger, Erik" <"Unger, Erik"@domonda.com>`:        {Name: "Unger, Erik", Address: "unger..erik@domonda.com"},
		`"\"Example\" <ar1@example.com>" <ar@example.com>`: {Name: "Example", Address: "ar@example.com"}, // Use "ar1" in name part vs "ar" in address part to test picking up the right part
		`<xy=erik@example.com>`:                            {Name: "", Address: "xy=erik@example.com"},
		// `A!#$%&'*+-/=?^_` + "`" + `{|}~@example.com`:      {Name: "", Address: "a!#$%&'*+-/=?^_`{|}~@example.com"},
		`A!#$%&'*+-/=?^_{|}~@example.com`: {Name: "", Address: "a!#$%&'*+-/=?^_{|}~@example.com"},

		`"Some.Name1@xbüro-yy-zzz.de" <Some.Name@xbüro-yy-zzz.de>`: {Name: "Some.Name1@xbüro-yy-zzz.de", Address: "some.name@xbüro-yy-zzz.de"}, // Use "Some.Name1" in name part vs "Some.Name" in address part to test picking up the right part

		// RFC 2047 encoding
		`=?utf-8?b?wqFIb2xhLCBzZcOxb3Ih?= <senor@hola.com>`: {Name: `¡Hola, señor!`, Address: "senor@hola.com"}, // mime.BEncoding.Encode("utf-8", `¡Hola, señor!`)

		`'stupid@quoting.me'`:                         {Name: ``, Address: "stupid@quoting.me"},
		`<'stupid@quoting.me'>`:                       {Name: ``, Address: "stupid@quoting.me"},
		`"'stupid@quoting.me'" <'stupid@quoting.me'>`: {Name: `'stupid@quoting.me'`, Address: "stupid@quoting.me"},

		`wow@xx.consulting`: {Name: ``, Address: "wow@xx.consulting"},

		`witha+plus@gmail.com`:        {Name: ``, Address: "witha+plus@gmail.com"},
		`endingwithnums777@gmail.com`: {Name: ``, Address: "endingwithnums777@gmail.com"},

		// `Non standard comma, in name <comma@example.com>`: {Name: `Non standard comma, in name`, Address: "comma@example.com"},
	}

	invalidEmailAddresses = map[string]struct{}{
		``:             {},
		` `:            {},
		"\t":           {},
		`,`:            {},
		`, `:           {},
		` , `:          {},
		`@`:            {},
		`@domonda.com`: {},
		// `.@domonda.com`: {},
		// `+@domonda.com`:       {}, // allowed?
		// `_@domonda.com`:       {}, // allowed?
		`Hello World!`:           {},
		`erik@`:                  {},
		`erik@domonda.com,`:      {},
		`erik@domonda.com, `:     {},
		`erik@domonda.com , `:    {},
		`.erik@domonda.com`:      {},
		`,erik@domonda.com`:      {},
		`, erik@domonda.com`:     {},
		` , erik@domonda.com`:    {},
		`unger erik@domonda.com`: {},
		// `If need of a ''Declaration of Compliance'' please contact us@example.com`: {},
	}
)

func TestParseAddress(t *testing.T) {
	// Test very special case
	result, err := ParseAddress(`"\"Example\" <ar1@example.com>" <ar@example.com>`)
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "Example" || result.Address != "ar@example.com" {
		t.Fatalf("wrong result: %v", result)
	}

	for addr, expected := range validEmailAddresses {
		t.Run(addr, func(t *testing.T) {
			result, err := ParseAddress(addr)
			assert.NoError(t, err, "valid email address")
			assert.Equal(t, expected, result, "address: %s", addr)
		})
	}

	for addr := range invalidEmailAddresses {
		t.Run(addr, func(t *testing.T) {
			result, err := ParseAddress(addr)
			if err == nil {
				t.Errorf("should not be able to be parsed as email address %s, but got: %s\nRegex: %s", addr, result, nameAddressRegex)
			}
		})
	}
}

func TestFindAllAddresses(t *testing.T) {
	tests := []struct {
		text string
		want []Address
	}{
		{
			text: "",
			want: nil,
		},
		{
			text: `Hello@world.com`,
			want: []Address{`Hello@world.com`},
		},
		{
			text: `Some.Name@xbüro-yy-zzz.de`,
			want: []Address{`Some.Name@xbüro-yy-zzz.de`}, // FindAllAddresses does not normalize to lower case
		},
		{
			text: `Hello@world.com for@example.com`,
			want: []Address{`Hello@world.com`, `for@example.com`},
		},
		{
			text: `Me Hello@world.com for nothing! for@example.com some text with@at symbol`,
			want: []Address{`Hello@world.com`, `for@example.com`},
		},
		{
			text: `Me <Hello@world.com> for nothing! <for@example.com> some text with@at symbol`,
			want: []Address{`Hello@world.com`, `for@example.com`},
		},
		{
			text: "Leading spaces Fax +49 (0)66 66 666 666 what@a-test.com\n",
			want: []Address{`what@a-test.com`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			if got := FindAllAddresses(tt.text); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindAllAddresses() = %v, want %v\nUsed regex: %s", got, tt.want, addressRegex)
			}
		})
	}
}

func TestParseAddressList(t *testing.T) {
	emptyLists := []string{
		"",
		"undisclosed-recipients",
		"undisclosed recipients",
		"Undisclosed Recipients",
	}
	for _, tt := range emptyLists {
		parsed, err := ParseAddressList(tt)
		if err != nil {
			t.Errorf("%q parsing error: %s", tt, err)
		}
		if parsed != nil {
			t.Errorf("%q not parsed as empty list", tt)
		}
	}

	// Test specific list we had problems with before
	// const problemList = "x@mail.example.com, x@mail-example.com, wow@xx.consulting, witha+plus@gmail.com, some_underscore@msn.com, refill@example24.de, nasa@7examples.com, erik@domonda.com <erik@domonda.com>, er+bill@mail-billwerk.co.uk, endingwithnums777@gmail.com, customerinfo@email.spammers.com, a@we-work.com, _underscore@example.com, Erik Unger <erik@domonda.com>, Erik Unger <Erik.Unger@domonda.com>, Erik Unger    <erik@domonda.com>, Domonda < er+vk+baurauslagen+wirklich@domonda.com>, Domonda < er+vk+baurauslagen+wirklich@domonda.com >, A!#$%&'*+-/=?^_{|}~@example.com, @Erik <erik@domonda.com>, =?utf-8?b?wqFIb2xhLCBzZcOxb3Ih?= <senor@hola.com>, <xy=erik@example.com>, <erik@domonda.com>, <'stupid@quoting.me'>, 'stupid@quoting.me', \"scanner@\" <example.at scanner@example.at>, \"\\\"Example\\\" <ar@example.com>\" <ar@example.com>, \"Unger, Erik\" <u.erik@domonda.com>, \"Unger, Erik\" <\"Unger, Erik\"@domonda.com>, \"Some.Name@xbüro-yy-zzz.de\" <Some.Name@xbüro-yy-zzz.de>, \"Erik Unger-Phd </Domonda-IT>\" <Erik.Unger-Phd@domonda-it.at>, \"Erik Unger\" <erik@domonda.com>, \"Erik Unger\" <\"Erik.Unger\"@domonda.com>, \"'stupid@quoting.me'\" <'stupid@quoting.me'>, \" Erik Unger \" <erik@domonda.com>" + `, "Unger, Erik"` + "\t<u.erik@domonda.com>, Erik\tUnger <erik@domonda.com>"
	// parsed, err := ParseAddressList(problemList)
	// if err != nil || len(parsed) != len(validEmailAddresses) {
	// 	fmt.Println("Regex:", nameAddressRegex)
	// 	fmt.Println("AddressList:", problemList)
	// 	for _, addr := range parsed {
	// 		fmt.Println(addr)
	// 	}
	// 	t.Fatalf("Error: %v\nlen(parsed): %d, len(validEmailAddresses): %d\nResult: %v\n\nproblemList: '%s'", err, len(parsed), len(validEmailAddresses), parsed, problemList)
	// }

	// The test lists are created by joining all validEmailAddresses.
	// First they are joined sorted and reverse-sorted by name to
	// create reproducable combinations.
	// Then additional random combinations are also created.

	const numRandomCombinations = 0 // TODO use 100

	// For every list create variations with different separators
	separators := []string{", ", ",", " ,", " , "}

	// Map from joined address-list to source addresses
	// which are also keys of validEmailAddresses
	tests := make(map[string][]string)

	{
		var sortedAddrs []string
		for addr := range validEmailAddresses {
			sortedAddrs = append(sortedAddrs, addr)
		}
		sort.Strings(sortedAddrs)
		for _, separator := range separators {
			list := strings.Join(sortedAddrs, separator)
			tests[list] = sortedAddrs
		}
	}

	{
		var reverseSortedAddrs []string
		for addr := range validEmailAddresses {
			reverseSortedAddrs = append(reverseSortedAddrs, addr)
		}
		sort.Sort(sort.Reverse(sort.StringSlice(reverseSortedAddrs)))
		for _, separator := range separators {
			list := strings.Join(reverseSortedAddrs, separator)
			tests[list] = reverseSortedAddrs
		}
	}

	for x := 0; x < numRandomCombinations; x++ {
		// Every range over the validEmailAddresses map
		// will produce a new random order
		var randomAddrs []string
		for addr := range validEmailAddresses {
			randomAddrs = append(randomAddrs, addr)
		}
		for _, separator := range separators {
			list := strings.Join(randomAddrs, separator)
			tests[list] = randomAddrs
		}
	}

	// Run all prepared tests
	for addressList, sourceAddrs := range tests {
		t.Run(addressList, func(t *testing.T) {
			// fmt.Println("\nLIST:", addressList)
			parsed, err := ParseAddressList(addressList)
			if err != nil {
				t.Fatalf("Error: %v\n\nAddressList: %s\n\nRegex: %s\n", err, addressList, nameAddressRegex)
			}
			if len(parsed) != len(sourceAddrs) {
				for _, addr := range parsed {
					fmt.Println(addr)
				}
				t.Fatalf("len(parsed):%d != len(sourceAddrs):%d", len(parsed), len(sourceAddrs))
			}
			for i := range parsed {
				parsedAddr := parsed[i]
				sourceAddr := sourceAddrs[i]
				expected := validEmailAddresses[sourceAddr]
				if expected.Address != parsedAddr.Address || expected.Name != parsedAddr.Name {
					t.Errorf("expected %#v but got %#v from list '%s'", expected, parsedAddr, addressList)
				}
			}
		})
	}

	// ParseAddressList should be able to parse single addresses like ParseEmailAddress
	for addr, expected := range validEmailAddresses {
		t.Run(addr, func(t *testing.T) {
			results, err := ParseAddressList(addr)
			assert.NoError(t, err, "valid email address")
			if assert.Len(t, results, 1, "parsed list length 1 for 1 address") {
				assert.Equal(t, expected, results[0], "expected: '%s'", expected)
			}
		})
	}

	// Invalid addresses are also invalid lists
	// except for empty trimmed strings wich are an empty list
	for addr := range invalidEmailAddresses {
		if strings.TrimSpace(addr) == "" {
			continue
		}
		t.Run(addr, func(t *testing.T) {
			result, err := ParseAddressList(addr)
			if err == nil {
				t.Errorf("should not be able to be parsed as email address %s, but got: %s\nRegex: %s", addr, result, nameAddressRegex)
			}
		})
	}
}
