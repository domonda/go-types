package bank

import "github.com/domonda/go-types/country"

// IBANFinder finds valid IBANs within a byte slice by scanning for known country codes
// and validating each candidate against the expected country-specific IBAN length.
var IBANFinder ibanFinder

type ibanFinder struct{}

func (ibanFinder) FindAllIndex(str []byte, n int) (result [][]int) {
	if n == 0 {
		return nil
	}
	strLen := len(str)
	max := strLen - IBANMinLength
	for i := 0; i <= max; i++ {
		countryCode := country.Code(str[i : i+2])
		countryLength, found := countryIBANLength[countryCode]
		if found {
			end := i + countryLength
			if end <= strLen {
				if IBAN(str[i:end]).Valid() {
					result = append(result, []int{i, end})
					if n > 0 && len(result) == n {
						return result
					}
					i = end - 1
					continue
				}
			}
		}
	}
	return result
}
