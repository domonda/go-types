package bank

import (
	"database/sql/driver"

	"github.com/domonda/errors"
)

///////////////////////////////////////////////////////////////////////////////
// AccountType

// AccountType holds the type of a bank account.
// AccountType implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// and will treat an empty string AccountType as SQL NULL value.
type AccountType string

var (
	// AccountTypeCurrent is a checking account type
	AccountTypeCurrent AccountType = "CURRENT"
	// AccountTypeSavings is a savings account type
	AccountTypeSavings AccountType = "SAVINGS"
)

func (t AccountType) Valid() bool {
	return t == AccountTypeCurrent || t == AccountTypeSavings
}

// Scan implements the database/sql.Scanner interface.
func (t *AccountType) Scan(value interface{}) error {
	switch x := value.(type) {
	case string:
		*t = AccountType(x)
	case []byte:
		*t = AccountType(x)
	case nil:
		*t = ""
	default:
		return errors.Errorf("can't scan SQL value of type %T as AccountType", value)
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
func (t AccountType) Value() (driver.Value, error) {
	if t == "" {
		return nil, nil
	}
	return string(t), nil
}

///////////////////////////////////////////////////////////////////////////////
// TransactionType

type TransactionType string

const (
	TransactionTypeIncoming TransactionType = "INCOMING"
	TransactionTypeOutgoing TransactionType = "OUTGOING"
)

///////////////////////////////////////////////////////////////////////////////
// PaymentStatus

type PaymentStatus string

const (
	PaymentStatusCreated  PaymentStatus = "CREATED"
	PaymentStatusFinished PaymentStatus = "FINISHED"
	PaymentStatusFailed   PaymentStatus = "FAILED"
)
