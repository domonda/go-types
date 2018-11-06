package bank

import (
	"database/sql/driver"

	"github.com/guregu/null"
)

///////////////////////////////////////////////////////////////////////////////
// AccountType

// AccountType holds the type of a bank account
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
	var ns null.String
	err := ns.Scan(value)
	if err != nil {
		return err
	}
	if ns.Valid {
		*t = AccountType(ns.String)
	} else {
		*t = ""
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
