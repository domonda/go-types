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
		return a.UnmarshalJSON(x)
	case string:
		return a.UnmarshalJSON([]byte(x))
	}
	return fmt.Errorf("can't scan value '%#v' of type %T as bank.Account", value, value)
}

// UnmarshalJSON implements encoding/json.Unmarshaler
func (a *Account) UnmarshalJSON(j []byte) (err error) {
	if len(j) < 2 {
		return fmt.Errorf("too short to unmarshal as bank.Account: `%s`", j)
	}
	if bytes.Equal(j, []byte("null")) {
		return nil // JSON null does not change the Account
	}
	beg := j[0]
	end := j[len(j)-1]
	if beg == '{' && end == '}' {
		// Unmarshal into a struct that does not
		// implmented UnmarshalText to avoid recursion
		var acc struct {
			IBAN     IBAN
			BIC      NullableBIC
			Currency money.NullableCurrency
			Holder   nullable.TrimmedString
		}
		err = json.Unmarshal(j, &acc)
		if err != nil {
			return fmt.Errorf("can't unmarshal `%s` for bank.Account: %w", j, err)
		}
		*a = Account(acc)
		return nil
	}
	// Unmarshal j as an IBAN string
	var iban IBAN
	if beg == '"' && end == '"' {
		// JSON string
		err := json.Unmarshal(j, &iban)
		if err != nil {
			return fmt.Errorf("can't unmarshal `%s` for bank.Account: %w", j, err)
		}
	} else {
		// Non-JSON text from UnmarshalText
		iban = IBAN(j)
	}
	iban, err = iban.Normalized()
	if err != nil {
		return fmt.Errorf("can't parse `%s` as IBAN for bank.Account: %w", j, err)
	}
	a.IBAN = iban
	a.BIC.SetNull()
	a.Currency.SetNull()
	a.Holder.SetNull()
	return nil
}

// UnmarshalText implements the [encoding.TextUnmarshaler] interface.
// The account can be unmarshalled from a JSON object or a IBAN string.
func (a *Account) UnmarshalText(text []byte) error {
	return a.UnmarshalJSON(bytes.TrimSpace(text))
}
