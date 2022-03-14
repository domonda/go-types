package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var emailAddressNormalized = map[Address]string{
	"test@test.com":                  "test@test.com",
	"te+st@test.com":                 "te+st@test.com",
	"test888@test.com":               "test888@test.com",
	"<noname@test.com>":              "noname@test.com",
	"With Name <With.Name@test.com>": `"With Name" <with.name@test.com>`,
}

func TestAddressNormalized(t *testing.T) {
	for addr, expected := range emailAddressNormalized {
		t.Run(string(addr), func(t *testing.T) {
			result, err := addr.Normalized()
			assert.NoError(t, err)
			assert.Equal(t, expected, string(result))
		})
	}
}
