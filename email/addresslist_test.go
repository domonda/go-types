package email

import (
	"net/mail"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStandardComparison(t *testing.T) {
	validEmailAddresses := map[string]*mail.Address{
		`"Unger, Erik" <u.erik@domonda.com>`:        {Name: "Unger, Erik", Address: "u.erik@domonda.com"},
		`"Unger, Erik" <"Unger, Erik"@domonda.com>`: {Name: "Unger, Erik", Address: "Unger, Erik@domonda.com"},
	}

	for addr, expected := range validEmailAddresses {
		t.Run(addr, func(t *testing.T) {
			result, err := mail.ParseAddress(addr)
			assert.NoError(t, err, "valid email address")
			assert.Equal(t, expected, result, "expected: %s", expected)
		})
	}

	for addr, expected := range validEmailAddresses {
		t.Run(addr, func(t *testing.T) {
			results, err := mail.ParseAddressList(addr)
			assert.NoError(t, err, "valid email address")
			assert.Len(t, results, 1, "list of one address")
			assert.Equal(t, expected, results[0], "expected: %s", expected)
		})
	}
}
