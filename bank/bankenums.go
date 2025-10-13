// Package bank provides comprehensive banking data types and utilities for Go applications.
//
// The package includes:
// - IBAN (International Bank Account Number) validation and parsing
// - BIC (Bank Identifier Code) validation and parsing
// - Bank account management with validation
// - CAMT53 bank statement parsing
// - Database integration (Scanner/Valuer interfaces)
// - JSON marshalling/unmarshalling
// - Nullable banking types support
package bank

import (
	"database/sql/driver"
	"fmt"
)

///////////////////////////////////////////////////////////////////////////////
// AccountType

// AccountType represents the type of a bank account.
// AccountType implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// and treats an empty string AccountType as SQL NULL value.
type AccountType string

var (
	// AccountTypeCurrent represents a checking/current account type
	AccountTypeCurrent AccountType = "CURRENT"
	// AccountTypeSavings represents a savings account type
	AccountTypeSavings AccountType = "SAVINGS"
)

// Valid returns true if the AccountType is a valid account type.
func (t AccountType) Valid() bool {
	return t == AccountTypeCurrent || t == AccountTypeSavings
}

// Scan implements the database/sql.Scanner interface.
func (t *AccountType) Scan(value any) error {
	switch x := value.(type) {
	case string:
		*t = AccountType(x)
	case []byte:
		*t = AccountType(x)
	case nil:
		*t = ""
	default:
		return fmt.Errorf("can't scan SQL value of type %T as AccountType", value)
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
// Returns nil for SQL NULL if the AccountType is empty.
func (t AccountType) Value() (driver.Value, error) {
	if t == "" {
		return nil, nil
	}
	return string(t), nil
}

///////////////////////////////////////////////////////////////////////////////
// TransactionType

// TransactionType represents the direction of a bank transaction.
type TransactionType string

const (
	// TransactionTypeIncoming represents an incoming transaction (credit)
	TransactionTypeIncoming TransactionType = "INCOMING"
	// TransactionTypeOutgoing represents an outgoing transaction (debit)
	TransactionTypeOutgoing TransactionType = "OUTGOING"
)

///////////////////////////////////////////////////////////////////////////////
// PaymentStatus

// PaymentStatus represents the status of a payment transaction.
type PaymentStatus string

const (
	// PaymentStatusCreated represents a payment that has been created
	PaymentStatusCreated PaymentStatus = "CREATED"
	// PaymentStatusFinished represents a payment that has been completed
	PaymentStatusFinished PaymentStatus = "FINISHED"
	// PaymentStatusFailed represents a payment that has failed
	PaymentStatusFailed PaymentStatus = "FAILED"
)
