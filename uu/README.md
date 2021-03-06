# UUID package for Go language

Base on package [github.com/satori/go.uuid](https://github.com/satori/go.uuid)

This package provides pure Go implementation of Universally Unique Identifier (UUID). Supported both creation and parsing of UUIDs.

With 100% test coverage and benchmarks out of box.

Supported versions:

* Version 1, based on timestamp and MAC address (RFC 4122)
* Version 2, based on timestamp, MAC address and POSIX UID/GID (DCE 1.1)
* Version 3, based on MD5 hashing (RFC 4122)
* Version 4, based on random numbers (RFC 4122)
* Version 5, based on SHA-1 hashing (RFC 4122)

## Installation

Use the `go` command:

```go get github.com/domonda/go-types/uu```

## Example

```go
package main

import (
    "fmt"
    "github.com/domonda/go-types/uu"
)

func main() {
    // Creating UUID Version 4
    u1 := uu.IDV4()
    fmt.Printf("UUIDv4: %s\n", u1)

    // Parsing UUID from string input
    u2, err := uu.IDFromString("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
    if err != nil {
        fmt.Printf("Something gone wrong: %s", err)
    }
    fmt.Printf("Successfully parsed: %s", u2)
}
```

## Documentation

[Documentation](http://godoc.org/github.com/domonda/go-types/uu) is hosted at GoDoc project.

## Links

* [RFC 4122](http://tools.ietf.org/html/rfc4122)
* [DCE 1.1: Authentication and Security Services](http://pubs.opengroup.org/onlinepubs/9696989899/chap5.htm#tagcjh_08_02_01_01)

## Copyright

Copyright (C) 2013-2016 by Maxim Bublis <b@codemonkey.ru>.

UUID package released under MIT License.
See [LICENSE](https://github.com/domonda/go-types/uu/blob/master/LICENSE) for details.
