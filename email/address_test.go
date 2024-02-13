package email

import (
	"net/mail"
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

func TestAddressFrom(t *testing.T) {
	tests := []struct {
		name string
		addr *mail.Address
		want Address
	}{
		{name: `nil`, addr: nil, want: ``},
		{name: `empty`, addr: &mail.Address{Name: "", Address: ""}, want: ``},
		{name: `erik@domonda.com`, addr: &mail.Address{Name: "", Address: "erik@domonda.com"}, want: `erik@domonda.com`},
		{name: `"Erik Unger" <erik@domonda.com>`, addr: &mail.Address{Name: "Erik Unger", Address: "erik@domonda.com"}, want: `"Erik Unger" <erik@domonda.com>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AddressFrom(tt.addr); got != tt.want {
				t.Errorf("AddressFrom() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestAddress_IsFromKnownEmailProvider(t *testing.T) {
	tests := []struct {
		a    Address
		want bool
	}{
		{``, false},
		{`example@domonda.com`, false},
		{`"Example User" <example@domonda.com>`, false},
		{`example@gmail.com`, true},
		{`guest123-not-valid@airbnb.com`, true},
		{`"Guest 123" <guest123-not-valid@airbnb.com>`, true},
	}
	for _, tt := range tests {
		t.Run(string(tt.a), func(t *testing.T) {
			if got := tt.a.IsFromKnownEmailProvider(); got != tt.want {
				t.Errorf("Address(%#v).IsFromKnownEmailProvider() = %v, want %v", tt.a, got, tt.want)
			}
		})
	}
}
