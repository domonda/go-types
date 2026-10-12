# How to extract identifiers, dates and amounts from text

This guide shows how to find IBANs, BICs, VAT IDs, dates and money amounts in the text of an invoice or a bank document, and how to turn each match into a normalized value.

## Prerequisites

- Go 1.26 or later and the module: `go get github.com/domonda/go-types`.
- The document as UTF-8 text, for example from the text layer of a PDF or from OCR.
- The language of the document, if you know it. It decides how ambiguous dates like `05/06/2024` are read.

All five finders implement [`types.Finder`](reference-types.md#finder):

```go
FindAllIndex(str []byte, n int) [][]int
```

It returns `[start, end)` byte offsets, so `str[start:end]` is the match. `n < 0` returns all matches, `n == 0` none, and `n > 0` at most `n`.

| Finder                    | Declared as                            | Normalize a match with                             |
|---------------------------|----------------------------------------|----------------------------------------------------|
| `bank.IBANFinder`         | package variable of an unexported type | `bank.NormalizeIBAN`                               |
| `bank.BICFinder`          | package variable of an unexported type | `bank.NormalizeBIC`                                |
| `vat.IDFinder`            | package variable of an unexported type | `vat.NormalizeVATID`                               |
| `date.NewFinder(lang...)` | function returning a `*date.Finder`    | `date.Normalize` with the same language            |
| `money.AmountFinder`      | package variable of an unexported type | `money.ParseCentAmount` or `money.ParseDecimalAmount` |

## Steps

The steps build one program. Step 6 shows it complete, with its output.

1. Write a helper that returns the text of every match, so each step can loop over strings:

   ```go
   // findAll returns the text of every match of f in text.
   func findAll(f types.Finder, text []byte) []string {
   	var matches []string
   	for _, loc := range f.FindAllIndex(text, -1) {
   		matches = append(matches, string(text[loc[0]:loc[1]]))
   	}
   	return matches
   }
   ```

2. Find the IBANs and normalize them. Remember their countries for the next step:

   ```go
   ibanCountries := make(map[country.Code]bool)
   for _, m := range findAll(bank.IBANFinder, text) {
   	iban, err := bank.NormalizeIBAN(m)
   	if err != nil {
   		continue
   	}
   	ibanCountries[iban.CountryCode()] = true
   	fmt.Println("IBAN:", iban)
   }
   ```

3. Find the BICs and normalize them. Upper-case words can have BIC syntax, so keep only BICs whose country (characters 5 and 6) matches an IBAN of the document:

   ```go
   for _, m := range findAll(bank.BICFinder, text) {
   	bic, err := bank.NormalizeBIC(m)
   	if err != nil || !ibanCountries[bic.CountryCode()] {
   		fmt.Println("skipped BIC:", m)
   		continue
   	}
   	fmt.Println("BIC:", bic)
   }
   ```

   The country check is a heuristic for domestic payments. Drop it if your documents name foreign banks, and check the context of the match instead, like a preceding `BIC` or `SWIFT` label.

4. Find the VAT IDs and normalize them. A match is the raw text, which can contain spaces:

   ```go
   for _, m := range findAll(vat.IDFinder, text) {
   	id, err := vat.NormalizeVATID(m)
   	if err != nil {
   		continue
   	}
   	fmt.Println("VAT ID:", id)
   }
   ```

5. Find the dates and the amounts. Pass the language of the document to `date.NewFinder` and the same language to `date.Normalize`: the finder only uses it to test candidates and returns the raw text. `money.AmountFinder` only matches amounts with exactly two decimals, so parse them as `CentAmount` with 2 accepted decimals, which also rejects the `NaN` and `Inf` tokens the finder matches:

   ```go
   for _, m := range findAll(date.NewFinder(language.DE), text) {
   	d, err := date.Normalize(m, language.DE)
   	if err != nil {
   		continue
   	}
   	fmt.Println("Date:", d)
   }
   for _, m := range findAll(money.AmountFinder, text) {
   	amount, err := money.ParseCentAmount(m, money.RoundHalfAwayFromZero, 2)
   	if err != nil {
   		continue
   	}
   	fmt.Println("Amount:", amount)
   }
   ```

6. Put the pieces together and run the program:

   ```go
   package main

   import (
   	"fmt"

   	"github.com/domonda/go-types"
   	"github.com/domonda/go-types/bank"
   	"github.com/domonda/go-types/country"
   	"github.com/domonda/go-types/date"
   	"github.com/domonda/go-types/language"
   	"github.com/domonda/go-types/money"
   	"github.com/domonda/go-types/vat"
   )

   // findAll returns the text of every match of f in text.
   func findAll(f types.Finder, text []byte) []string {
   	var matches []string
   	for _, loc := range f.FindAllIndex(text, -1) {
   		matches = append(matches, string(text[loc[0]:loc[1]]))
   	}
   	return matches
   }

   func main() {
   	text := []byte(`Muster GmbH, Hauptstraße 1, 1010 Wien
   UID: ATU 1022 3006
   RECHNUNG Nr. 2023-117 vom 25.12.2023
   IBAN: AT611904300234573201 BIC: GIBAATWWXXX
   Gesamtbetrag: EUR 1.234,50
   Zahlbar bis 08.01.2024 ohne Abzug.`)

   	ibanCountries := make(map[country.Code]bool)
   	for _, m := range findAll(bank.IBANFinder, text) {
   		iban, err := bank.NormalizeIBAN(m)
   		if err != nil {
   			continue
   		}
   		ibanCountries[iban.CountryCode()] = true
   		fmt.Println("IBAN:", iban)
   	}

   	for _, m := range findAll(bank.BICFinder, text) {
   		bic, err := bank.NormalizeBIC(m)
   		if err != nil || !ibanCountries[bic.CountryCode()] {
   			fmt.Println("skipped BIC:", m)
   			continue
   		}
   		fmt.Println("BIC:", bic)
   	}

   	for _, m := range findAll(vat.IDFinder, text) {
   		id, err := vat.NormalizeVATID(m)
   		if err != nil {
   			continue
   		}
   		fmt.Println("VAT ID:", id)
   	}

   	for _, m := range findAll(date.NewFinder(language.DE), text) {
   		d, err := date.Normalize(m, language.DE)
   		if err != nil {
   			continue
   		}
   		fmt.Println("Date:", d)
   	}

   	for _, m := range findAll(money.AmountFinder, text) {
   		amount, err := money.ParseCentAmount(m, money.RoundHalfAwayFromZero, 2)
   		if err != nil {
   			continue
   		}
   		fmt.Println("Amount:", amount)
   	}

   	// Output:
   	// IBAN: AT611904300234573201
   	// skipped BIC: RECHNUNG
   	// BIC: GIBAATWWXXX
   	// VAT ID: ATU10223006
   	// Date: 2023-12-25
   	// Date: 2024-01-08
   	// Amount: 1234.50
   }
   ```

## Verification

The output of step 6 shows the expected results:

- The IBAN, the BIC, the VAT ID and the dates are in their canonical forms. `ValidAndNormalized()` is true for each of them, so they can be stored and compared as they are.
- `RECHNUNG`, German for "invoice", is valid BIC syntax with the country code `NU` (Niue). Step 3 skips it.
- `2023-117` and the postal code `1010` are not reported as dates or amounts.
- The VAT ID was found although the text splits it into three words: `ATU 1022 3006`.

To test whether a document contains an identifier at all, pass `n = 1`: `len(bank.IBANFinder.FindAllIndex(text, 1)) > 0` stops at the first match.

## What each finder misses

| Finder               | Finds                                              | Misses or mismatches                               |
|----------------------|----------------------------------------------------|----------------------------------------------------|
| `bank.IBANFinder`    | Compact upper-case IBANs with a valid check sum, at any position | Grouped IBANs (`AT61 1904 3002 3457 3201`) and lower case. It checks no word boundaries, so `XAT611904300234573201` yields `AT611904300234573201`. |
| `bank.BICFinder`     | Upper-case 8 or 11 character words with BIC syntax and a valid country code, between word separators | Lower-case and spaced BICs. Upper-case words like `RECHNUNG` and `DEUTSCHE` (country `SC`) are found. The deny-list has only 9 German words. |
| `vat.IDFinder`       | Runs of one to three words, split at whitespace and `:`, that form a valid VAT ID | IDs written as four or more words: `DE 167 015 661` is missed, `DE 167015661` is found. Prefixes without a format like `US`, and `GR` (Greek IDs use `EL`). |
| `date.NewFinder`     | Runs of one to three words that `date.Normalize` accepts | Everything `date.Normalize` rejects, like `2014-04` and `20140601`. |
| `money.AmountFinder` | Words with exactly two decimals, like `1.234,50`, `-1234.56` and `99.00`, and the words `NaN`, `Inf`, `+Inf` and `-Inf` | `12.5`, `1234` and amounts with an attached symbol like `€1.234,50`. A day and month like `25.12` look like an amount and are found. |

## Troubleshooting

**A grouped IBAN like `AT61 1904 3002 3457 3201` is not found.** `IBANFinder` only matches the compact form. Find grouped candidates with a regular expression and let `bank.NormalizeIBAN` reject the wrong ones:

```go
package main

import (
	"fmt"
	"regexp"

	"github.com/domonda/go-types/bank"
)

var groupedIBAN = regexp.MustCompile(`\b[A-Z]{2}\d{2}(?: ?[A-Z0-9]{4}){2,7}(?: ?[A-Z0-9]{1,3})?\b`)

func main() {
	text := []byte("IBAN: AT61 1904 3002 3457 3201, Konto DE89 3704 0044 0532 0130 00.")
	for _, loc := range groupedIBAN.FindAllIndex(text, -1) {
		iban, err := bank.NormalizeIBAN(string(text[loc[0]:loc[1]]))
		if err != nil {
			continue
		}
		fmt.Println(iban)
	}

	// Output:
	// AT611904300234573201
	// DE89370400440532013000
}
```

The expression can also swallow a following group of 1 to 4 letters or digits, like an account number written right after the IBAN. `NormalizeIBAN` then fails with `wrong IBAN length` and the IBAN is skipped.

**An amount with a currency symbol like `€1.234,50` is not found.** `AmountFinder` splits words only at whitespace and `:`. Replace the symbol with as many spaces as it has bytes (`€` has 3 in UTF-8). The offsets then still point into the original text:

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/domonda/go-types/money"
)

func main() {
	text := []byte("Summe: €1.234,50 inkl. 20% USt.")
	search := bytes.ReplaceAll(text, []byte("€"), []byte("   "))
	for _, loc := range money.AmountFinder.FindAllIndex(search, -1) {
		fmt.Printf("%s at %d-%d\n", text[loc[0]:loc[1]], loc[0], loc[1])
	}

	// Output:
	// 1.234,50 at 10-18
}
```

`money.ParseCurrencyAmount("€1.234,50")` parses such a word with its currency, if you already have it.

**`date.Normalize` returns a different date than you expect for `05/06/2024`.** Without a language, ambiguous dates are read day first: 2024-06-05. Pass `language.EN` to both `date.NewFinder` and `date.Normalize` for documents with month first dates. To detect the order from the document itself, see [Detecting the date order of a whole document](explanation-lenient-parsing.md#detecting-the-date-order-of-a-whole-document).

**`vat.NormalizeVATID` fails with `invalid VAT ID: "GR094019245" has an unsupported country code: "GR"`.** Greek VAT IDs use the prefix `EL`, and `IDFinder` doesn't find `GR` IDs either. Replace the prefix if your source writes `GR`.

**`FindAllIndex` returns nil although the text has matches.** Check `n`: `0` returns no matches. Pass `-1` for all of them.

## Related

- [`bank`](../bank/README.md): `IBANFinder`, `BICFinder` and the IBAN and BIC rules.
- [`vat`](../vat/README.md): `IDFinder` and the supported prefixes.
- [`date`](../date/README.md): `Finder` and the accepted date shapes.
- [`money`](../money/README.md): `AmountFinder`, `StringIsAmount` and the amount parsers.
- [Root package reference](reference-types.md): the `Finder` interface.
- [Lenient parsing](explanation-lenient-parsing.md): why the parsers accept these spellings and how they decide between readings.
- [Tutorial: build an invoice record](tutorial-invoice-record.md): what to do with the extracted values.
