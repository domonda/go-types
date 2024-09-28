package bank

import (
	"errors"
	"fmt"
	"strings"

	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/nullable"
)

type Account struct {
	IBAN     IBAN
	BIC      NullableBIC
	Currency money.NullableCurrency
	Holder   nullable.TrimmedString
}

func (a *Account) Valid() bool {
	return a != nil && a.IBAN.Valid() && a.BIC.Valid() && a.Currency.Valid()
}

func (a *Account) Validate() error {
	if a == nil {
		return errors.New("nil bank.Account")
	}
	return errors.Join(
		a.IBAN.Validate(),
		a.BIC.Validate(),
		a.Currency.Validate(),
	)
}

func (a *Account) String() string {
	var b strings.Builder
	b.WriteString("bank.Account{")
	b.WriteString(a.IBAN.String())
	if a.BIC.IsNotNull() {
		fmt.Fprintf(&b, ", BIC: %s", a.BIC)
	}
	if a.Currency.IsNotNull() {
		fmt.Fprintf(&b, ", %s", a.Currency)
	}
	if a.Holder.IsNotNull() {
		fmt.Fprintf(&b, ", %q", a.Holder)
	}
	b.WriteByte('}')
	return b.String()
}
