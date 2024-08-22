package bank

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var toNormalizeIBANs = map[string]IBAN{
	// // 2 parts
	"AT74 2011182151456500": "AT742011182151456500",
	"AT 742011182151456500": "AT742011182151456500",
	// 5 parts
	"AT05 1937 0711 1000 0044": "AT051937071110000044",
	// ":AT05 1937 0711 1000 0044.": "AT051937071110000044",
	"FO97 5432 0388 8999 44": "FO9754320388899944",
	// 6 parts
	"DE36 7015 0000 0081 1342 98": "DE36701500000081134298",
	"DE 3670 1500 0000 8113 4298": "DE36701500000081134298",
	"AT 82 38487 000 000 27185":   "AT823848700000027185",
	"LV80 BANK 0000 4351 9500 1":  "LV80BANK0000435195001",
	// 7 parts
	"DE 36 7015 0000 0081 1342 98": "DE36701500000081134298",
	// 8 parts
	"MU17 BOMM 0101 1010 3030 0200 000M UR":  "MU17BOMM0101101030300200000MUR",
	"MT84 MALT 0110 0001 2345 MTLC AST0 01S": "MT84MALT011000012345MTLCAST001S",
	"QA58 DOHB 0000 1234 5678 90AB CDEF G":   "QA58DOHB00001234567890ABCDEFG",
}

var invalidIBANs = []string{
	"at05 1937 0711 1000 0044",
}

// http://www.rbs.co.uk/corporate/international/g0/guide-to-international-business/regulatory-information/iban/iban-example.ashx
var countryIBANTable = []string{
	"Albania", "AL47 2121 1009 0000 0002 3569 8741",
	"Andorra", "AD12 0001 2030 2003 5910 0100",
	"Austria", "AT61 1904 3002 3457 3201",
	"Azerbaijan", "AZ21 NABZ 0000 0000 1370 1000 1944",
	"Bahrain", "BH67 BMAG 0000 1299 1234 56",
	"Belgium", "BE62 5100 0754 7061",
	"Bosniaand Herzegovina", "BA39 1290 0794 0102 8494",
	"Bulgaria", "BG80 BNBG 9661 1020 3456 78",
	"Croatia", "HR12 1001 0051 8630 0016 0",
	"Cyprus", "CY17 0020 0128 0000 0012 0052 7600",
	"Czech Republic", "CZ65 0800 0000 1920 0014 5399",
	"Denmark", "DK50 0040 0440 1162 43",
	"Estonia", "EE38 2200 2210 2014 5685",
	"Faroe Islands", "FO97 5432 0388 8999 44",
	"Finland", "FI21 1234 5600 0007 85",
	"France", "FR14 2004 1010 0505 0001 3M02 606",
	"Georgia", "GE29 NB00 0000 0101 9049 17",
	"Germany", "DE89 3704 0044 0532 0130 00",
	"Gibraltar", "GI75 NWBK 0000 0000 7099 453",
	"Greece", "GR16 0110 1250 0000 0001 2300 695",
	"Greenland", "GL56 0444 9876 5432 10",
	"Hungary", "HU42 1177 3016 1111 1018 0000 0000",
	"Iceland", "IS14 0159 2600 7654 5510 7303 39",
	"Ireland", "IE29 AIBK 9311 5212 3456 78",
	"Israel", "IL62 0108 0000 0009 9999 999",
	"Italy", "IT40 S054 2811 1010 0000 0123 456",
	"Jordan", "JO94 CBJO 0010 0000 0000 0131 0003 02",
	"Kuwait", "KW81 CBKU 0000 0000 0000 1234 5601 01",
	"Latvia", "LV80 BANK 0000 4351 9500 1",
	"Lebanon", "LB62 0999 0000 0001 0019 0122 9114",
	"Liechtenstein", "LI21 0881 0000 2324 013A A",
	"Lithuania", "LT12 1000 0111 0100 1000",
	"Luxembourg", "LU28 0019 4006 4475 0000",
	"Macedonia", "MK07 2501 2000 0058 984",
	"Malta", "MT84 MALT 0110 0001 2345 MTLC AST0 01S",
	"Mauritius", "MU17 BOMM 0101 1010 3030 0200 000M UR",
	"Moldova", "MD24 AG00 0225 1000 1310 4168",
	"Monaco", "MC93 2005 2222 1001 1223 3M44 555",
	"Montenegro", "ME25 5050 0001 2345 6789 51",
	"Netherlands", "NL39 RABO 0300 0652 64",
	"Norway", "NO93 8601 1117 947",
	"Pakistan", "PK36 SCBL 0000 0011 2345 6702",
	"Poland", "PL60 1020 1026 0000 0422 7020 1111",
	"Portugal", "PT50 0002 0123 1234 5678 9015 4",
	"Qatar", "QA58 DOHB 0000 1234 5678 90AB CDEF G",
	"Romania", "RO49 AAAA 1B31 0075 9384 0000",
	"San Marino", "SM86 U032 2509 8000 0000 0270 100",
	"Saudi Arabia", "SA03 8000 0000 6080 1016 7519",
	"Serbia", "RS35 2600 0560 1001 6113 79",
	"Slovak Republic", "SK31 1200 0000 1987 4263 7541",
	"Slovenia", "SI56 1910 0000 0123 438",
	"Spain", "ES80 2310 0001 1800 0001 2345",
	"Sweden", "SE35 5000 0000 0549 1000 0003",
	"Switzerland", "CH93 0076 2011 6238 5295 7",
	"Tunisia", "TN59 1000 6035 1835 9847 8831",
	"Turkey", "TR33 0006 1005 1978 6457 8413 26",
	"UAE", "AE07 0331 2345 6789 0123 456",
	// "United Kingdom", "GB29 RBOS 6016 1331 9268 19", // invalid checksum
}

func Test_NormalizeIBAN(t *testing.T) {
	for testIBAN, expectedIBAN := range toNormalizeIBANs {
		normalized, err := NormalizeIBAN(testIBAN)
		if err != nil {
			t.Errorf("NormalizeIBAN(%s): %s", string(testIBAN), err.Error())
			continue
		}
		if normalized != expectedIBAN {
			t.Errorf("NormalizeIBAN(%s): %s != %s", string(testIBAN), string(normalized), string(expectedIBAN))
		}
	}

	for _, invalidIBAN := range invalidIBANs {
		normalized, err := NormalizeIBAN(invalidIBAN)
		if err == nil {
			t.Errorf("Should NOT be valid NormalizeIBAN(%s): %s", string(invalidIBAN), string(normalized))
			require.Equal(t, invalidIBAN, normalized, "Normalized invalid IBAN returned unchanged")
		}
	}

	for i := 0; i < len(countryIBANTable); i += 2 {
		country, iban := countryIBANTable[i], countryIBANTable[i+1]
		normalized, err := IBAN(iban).NormalizedWithSpaces()
		if err != nil {
			t.Errorf("NormalizedSpacedOrErr(%s): %s [%s]", string(iban), err.Error(), string(country))
			continue
		}
		if string(normalized) != iban {
			t.Errorf("NormalizedSpacedOrErr(%s): %s != %s [%s]", string(iban), string(normalized), string(iban), string(country))
		}
	}
}

var bankAndAccountNumbersTable = map[IBAN][2]string{
	"AT252011183728861100": {"20111", "83728861100"},
}

func Test_IBAN_BankAndAccountNumbers(t *testing.T) {
	for iban, numbers := range bankAndAccountNumbersTable {
		bankNr, accountNr, err := iban.BankAndAccountNumbers()
		if err != nil {
			t.Errorf("error from IBAN.BankAndAccountNumbers(%s): %s", string(iban), err)
			continue
		}
		if bankNr != numbers[0] || accountNr != numbers[1] {
			t.Errorf("IBAN.BankAndAccountNumbers(%s): not correct", string(iban))
		}
	}
}

// var ibanPartsMap = map[string][]string{
// 	"AT051937071110000044":   []string{"AT05", "1937", "0711", "1000", "0044"},
// 	"DE36701500000081134298": []string{"DE36", "7015", "0000", "0081", "1342", "98"},
// 	// "DE19123412341234123412": []string{"DE", "1912", "3412", "3412", "3412", "3412"}, // invalid checksum
// }

// func Test_IBANFromParts(t *testing.T) {
// 	for correctIBAN, parts := range ibanPartsMap {
// 		testIBAN, err := IBANFromParts(parts)
// 		if err != nil {
// 			t.Errorf("IBANFromParts(%v): %s", parts, err.Error())
// 			continue
// 		}
// 		if string(testIBAN) != correctIBAN {
// 			t.Errorf("IBANFromParts(%v): %s != %s", parts, string(testIBAN), string(correctIBAN))
// 		}
// 	}
// }
