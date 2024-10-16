package bank

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/nullable"
)

// Account identifies a bank account by its IBAN and optionally BIC.
type Account struct {
	IBAN     IBAN                   `json:"iban"`
	BIC      NullableBIC            `json:"bic,omitempty"`
	Currency money.NullableCurrency `json:"currency,omitempty"`
	Holder   nullable.TrimmedString `json:"holder,omitempty"`
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

// String returns a string representation of the Account
// usabled for debugging.
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

// Scan implements the database/sql.Scanner interface.
func (a *Account) Scan(value any) (err error) {
	switch x := value.(type) {
	case []byte:
		return a.UnmarshalText(x)
	case string:
		return a.UnmarshalText([]byte(x))
	}
	return fmt.Errorf("can't scan value '%#v' of type %T as bank.Account", value, value)
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// The account can be unmarshalled from a JSON object or a IBAN string.
func (a *Account) UnmarshalText(text []byte) error {
	text = bytes.TrimSpace(text)
	if len(text) > 0 && text[0] == '{' {
		// Unmarshal into a struct that does not
		// implmented UnmarshalText to avoid recursion
		var acc struct {
			IBAN     IBAN
			BIC      NullableBIC
			Currency money.NullableCurrency
			Holder   nullable.TrimmedString
		}
		err := json.Unmarshal(text, &acc)
		if err != nil {
			return fmt.Errorf("can't unmarshal `%s` as JSON for bank.Account: %w", text, err)
		}
		*a = Account(acc)
		return nil
	}
	iban, err := IBAN(text).Normalized()
	if err != nil {
		return fmt.Errorf("can't parse `%s` as IBAN for bank.Account: %w", text, err)
	}
	a.IBAN = iban
	// a.BIC.SetNull()
	// a.Currency.SetNull()
	// a.Holder.SetNull()
	return nil
}
