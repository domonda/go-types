# go-types

A comprehensive Go library providing enhanced type definitions, validation, and utility functions for common data types and operations.

## Overview

`go-types` is a collection of Go packages that extend the standard library with additional type definitions, validation frameworks, and utility functions. It provides type-safe implementations for common data structures and operations, making it easier to work with validated data, nullable types, and specialized data formats.

## Features

- **Type-Safe Collections**: Generic Set implementation with comprehensive operations
- **Validation Framework**: Flexible validation system with both boolean and error-based validation
- **Nullable Types**: Support for nullable values with proper JSON marshaling/unmarshaling
- **Specialized Data Types**: Email addresses, dates, money amounts, VAT IDs, and more
- **String Utilities**: Enhanced string manipulation and formatting functions
- **Iterator Utilities**: Helper functions for Go's iterator patterns
- **JSON Support**: Comprehensive JSON marshaling/unmarshaling for all types

## Core Packages

### Main Package (`types`)

The main package provides core utilities and type definitions:

#### Collections
- **Set[T]**: A generic set implementation with comprehensive operations (union, intersection, difference, etc.)
- **Slice Utilities**: Functions for checking slice containment (`SliceContainsAll`, `SliceContainsAny`)

#### Validation Framework
- **Validator**: Interface for boolean validation
- **ValidatErr**: Interface for validation with detailed error information
- **DeepValidate**: Recursive validation of nested structures
- **CombinedValidator**: Combine multiple validators

#### Utility Functions
- **Pointer Utilities**: Safe pointer dereferencing (`Ptr`, `FromPtr`, `FromPtrOr`)
- **Iterator Helpers**: Functions for working with Go's iterator patterns (`Yield`, `Yield2`, `YieldErr`)
- **JSON Utilities**: Check if types can be marshaled as JSON (`CanMarshalJSON`)

#### Specialized Types
- **LenString**: String with length constraints and validation
- **Finder**: Interface for pattern matching in byte slices

### Subpackages

#### `email` - Email Address Handling
- **Address**: Email address type with normalization and validation
- **AddressList**: Collection of email addresses
- **AddressSet**: Set of unique email addresses
- **Attachment**: Email attachment handling
- **Message**: Complete email message structure
- **ParseAddress**: Flexible email address parsing

#### `money` - Financial Data Types
- **Amount**: Money amount with decimal precision handling
- **Currency**: Currency codes and information
- **CurrencyAmount**: Combined currency and amount
- **Rate**: Exchange rate handling
- **AmountParser**: Parse monetary values from strings
- **CurrencyParser**: Parse currency information

#### `date` - Date Handling
- **Date**: Date type with comprehensive parsing and formatting
- **NullableDate**: Nullable date type
- **DateNames**: Localized date names
- **DateFinder**: Find dates in text
- **DateFormatter**: Format dates in various styles
- **DateParser**: Parse dates from strings

#### `bank` - Banking Data Types
- **IBAN**: International Bank Account Number handling
- **BIC**: Bank Identifier Code handling
- **Account**: Bank account information
- **CAMT53**: ISO 20022 CAMT.053 message handling
- **IBANParser/BICParser**: Parse banking identifiers

#### `vat` - VAT ID Handling
- **ID**: VAT identification number with country-specific validation
- **IDFinder**: Find VAT IDs in text
- **IDParser**: Parse VAT IDs from strings
- **NullableID**: Nullable VAT ID type

#### `country` - Country Information
- **Code**: ISO country codes
- **NullableCode**: Nullable country code
- **Country Data**: Comprehensive country information

#### `language` - Language Codes
- **Code**: ISO 639 language codes
- **Constants**: Language code constants
- **ISO6393**: ISO 639-3 language names

#### `nullable` - Nullable Types
- **Type[T]**: Generic nullable type wrapper
- **Arrays**: Nullable array types for various data types
- **NonEmptyString**: String that cannot be empty
- **TrimmedString**: String with automatic trimming
- **Time**: Nullable time type

#### `strutil` - String Utilities
- **String Manipulation**: Enhanced string functions
- **DomainName**: Domain name validation and manipulation
- **HTML Utilities**: HTML processing functions
- **Random**: Random string generation
- **StringSet**: Set of strings
- **StrMutex**: String-based mutex for synchronization

#### `strfmt` - String Formatting
- **Format**: String format detection and handling
- **Scanner**: String scanning utilities
- **Formatter**: String formatting functions
- **Detector**: Format detection algorithms

#### `float` - Float Utilities
- **Parse**: Enhanced float parsing with tolerance
- **Format**: Float formatting utilities
- **Tolerant**: Tolerant float operations

#### `queue` - Queue Implementation
- **Queue**: Generic queue interface and implementation
- **RingBuffer**: Ring buffer for efficient queue operations

#### `charset` - Character Set Handling
- **Encoding**: Character encoding detection and conversion
- **UTF8/UTF16/UTF32**: Unicode handling utilities
- **BOM**: Byte Order Mark handling

#### `uu` - UUID Utilities
- **ID**: UUID handling and generation
- **IDContext**: Context-aware UUID operations
- **IDMutex**: UUID-based mutex for synchronization
- **IDSet**: Set of UUIDs
- **IDSlice**: Slice of UUIDs with utilities

## Installation

```bash
go get github.com/domonda/go-types
```

## Usage Examples

### Sets

```go
package main

import (
    "fmt"
    "github.com/domonda/go-types"
)

func main() {
    // Create a set
    set1 := types.NewSet(1, 2, 3, 4, 5)
    set2 := types.NewSet(4, 5, 6, 7, 8)
    
    // Set operations
    union := set1.Union(set2)
    intersection := set1.Intersection(set2)
    difference := set1.Difference(set2)
    
    fmt.Printf("Union: %v\n", union.Sorted())
    fmt.Printf("Intersection: %v\n", intersection.Sorted())
    fmt.Printf("Difference: %v\n", difference.Sorted())
}
```

### Validation

```go
package main

import (
    "fmt"
    "github.com/domonda/go-types"
)

type User struct {
    Name  string
    Email string
    Age   int
}

func (u User) Validate() error {
    if u.Name == "" {
        return fmt.Errorf("name is required")
    }
    if u.Age < 0 {
        return fmt.Errorf("age must be non-negative")
    }
    return nil
}

func main() {
    user := User{Name: "John", Email: "john@example.com", Age: 25}
    
    // Validate single value
    if err := types.Validate(user); err != nil {
        fmt.Printf("Validation error: %v\n", err)
    }
    
    // Deep validation of nested structures
    if err := types.DeepValidate(user); err != nil {
        fmt.Printf("Deep validation error: %v\n", err)
    }
}
```

### Email Addresses

```go
package main

import (
    "fmt"
    "github.com/domonda/go-types/email"
)

func main() {
    // Parse and normalize email address
    addr, err := email.NormalizedAddress("John Doe <john@example.com>")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Normalized: %s\n", addr)
    
    // Create address list
    list := email.AddressList{addr, email.Address("jane@example.com")}
    fmt.Printf("Address list: %v\n", list)
}
```

### Money Amounts

```go
package main

import (
    "fmt"
    "github.com/domonda/go-types/money"
)

func main() {
    // Parse money amount
    amount, err := money.ParseAmount("123.45", 2) // 2 decimal places
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Amount: %v\n", amount)
    
    // Create currency amount
    currencyAmount := money.CurrencyAmount{
        Amount:   amount,
        Currency: money.Currency("USD"),
    }
    
    fmt.Printf("Currency amount: %v\n", currencyAmount)
}
```

### Dates

```go
package main

import (
    "fmt"
    "github.com/domonda/go-types/date"
)

func main() {
    // Parse date from string
    d, err := date.Normalize("2024-01-15")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Date: %s\n", d)
    fmt.Printf("Is valid: %v\n", d.Valid())
}
```

### Nullable Types

```go
package main

import (
    "fmt"
    "github.com/domonda/go-types/nullable"
)

func main() {
    // Create nullable string
    nullableStr := nullable.New("hello")
    
    fmt.Printf("Value: %v\n", nullableStr.Value())
    fmt.Printf("Is null: %v\n", nullableStr.IsNull())
    
    // Set to null
    nullableStr.SetNull()
    fmt.Printf("Is null after SetNull: %v\n", nullableStr.IsNull())
}
```

## JSON Support

All types in `go-types` support JSON marshaling and unmarshaling:

```go
package main

import (
    "encoding/json"
    "fmt"
    "github.com/domonda/go-types"
)

func main() {
    // Set JSON marshaling
    set := types.NewSet(1, 2, 3, 4, 5)
    
    jsonData, err := json.Marshal(set)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("JSON: %s\n", string(jsonData))
    
    // Unmarshal back to set
    var newSet types.Set[int]
    err = json.Unmarshal(jsonData, &newSet)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Unmarshaled set: %v\n", newSet.Sorted())
}
```

## Testing

The library includes comprehensive tests for all functionality:

```bash
go test ./...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Dependencies

- Go 1.24.0 or later
- Standard library packages
- External dependencies are minimal and well-maintained

## Performance

The library is designed with performance in mind:
- Generic types provide compile-time type safety
- Efficient data structures (sets implemented as maps)
- Minimal allocations where possible
- Comprehensive caching for expensive operations

## Security

Security considerations:
- Input validation for all user-provided data
- Safe handling of nullable types
- Proper error handling and reporting
- No unsafe operations exposed in public API