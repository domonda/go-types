package email

import (
	"net/mail"
	"reflect"
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

func TestAddressList_Split(t *testing.T) {
	tests := []struct {
		l       AddressList
		want    []Address
		wantErr bool
	}{
		{l: ``, want: nil},
		{l: `<hello@example.com>,`, want: []Address{`hello@example.com`}},
		{l: `<Hello@example.com>, World@example.com`, want: []Address{`hello@example.com`, `world@example.com`}},
	}
	for _, tt := range tests {
		t.Run(string(tt.l), func(t *testing.T) {
			got, err := tt.l.Split()
			if (err != nil) != tt.wantErr {
				t.Errorf("AddressList(%#v).Split() error = %v, wantErr %v", string(tt.l), err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddressList(%#v).Split() = %v, want %v", string(tt.l), got, tt.want)
			}
		})
	}
}
