package bank

import "github.com/domonda/go-types/country"

var IBANFinder ibanFinder

type ibanFinder struct{}

func (ibanFinder) FindAllIndex(str []byte, n int) (result [][]int) {
	strLen := len(str)
	max := strLen - IBANMinLength
	for i := 0; i <= max; i++ {
		countryCode := country.Code(str[i : i+2])
		countryLength, found := ibanCountryLengthMap[countryCode]
		if found {
			end := i + countryLength
			if end <= strLen {
				if IBAN(str[i:end]).Valid() {
					result = append(result, []int{i, end})
					i = end - 1
					continue
				}
			}
		}
	}
	return result
}
