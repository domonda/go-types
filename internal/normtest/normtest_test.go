package normtest

import (
	"database/sql/driver"
	"encoding/xml"
	"errors"
	"strings"
	"testing"

	"github.com/domonda/go-types/internal"
)

// refCode is a minimal type that implements the reading and writing
// contract with the internal helpers, to check that Run accepts a
// conforming type. The domain types run Run in their own tests.
type refCode string

func (c refCode) Normalized() (refCode, error) {
	norm := refCode(strings.ToUpper(strings.TrimSpace(string(c))))
	if len(norm) != 2 || norm == "XX" {
		return c, errors.New("invalid refCode")
	}
	return norm, nil
}

func (c refCode) Valid() bool              { _, err := c.Normalized(); return err == nil }
func (c refCode) Validate() error          { _, err := c.Normalized(); return err }
func (c refCode) ValidAndNormalized() bool { n, err := c.Normalized(); return err == nil && n == c }

func (c *refCode) Scan(value any) error {
	return internal.ScanNormalized(c, value, refCode.Normalized, nil)
}

func (c refCode) Value() (driver.Value, error) {
	return internal.ValueNormalized(c, refCode.Normalized, nil)
}

func (c refCode) MarshalJSON() ([]byte, error) {
	return internal.MarshalJSONNormalized(c, refCode.Normalized, nil)
}

func (c *refCode) UnmarshalJSON(data []byte) error {
	return internal.UnmarshalJSONNormalized(c, data, false, refCode.Normalized, nil)
}

func (c refCode) MarshalText() ([]byte, error) {
	return internal.MarshalTextNormalized(c, refCode.Normalized, nil)
}

func (c *refCode) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return internal.UnmarshalXMLNormalized(c, d, start, refCode.Normalized, nil)
}

func (c *refCode) UnmarshalXMLAttr(attr xml.Attr) error {
	return internal.UnmarshalXMLAttrNormalized(c, attr, refCode.Normalized, nil)
}

func (c *refCode) ScanString(source string, validate bool) error {
	return internal.ScanStringNormalized(c, source, validate, refCode.Normalized, nil)
}

func TestRun(t *testing.T) {
	Run(t, Spec[refCode]{
		Canonical: "DE",
		Variants:  []refCode{"de", " De "},
		Invalid:   []refCode{"XX", "toolong"},
	})
}
