// Package money contains decimal-string value types used by Dextri Pay.
package money

import (
	"fmt"
	"regexp"
)

var decimalPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)(\.[0-9]+)?$`)

// Decimal is a canonical non-negative base-10 value. The SDK never converts
// it to binary floating point; chain asset precision remains authoritative.
type Decimal string

// String returns the unchanged base-10 representation.
func (decimal Decimal) String() string { return string(decimal) }

// Validate checks decimal-string syntax without assuming an asset precision.
func (decimal Decimal) Validate() error {
	if !decimalPattern.MatchString(string(decimal)) {
		return fmt.Errorf("decimal value must be a canonical non-negative decimal string")
	}
	return nil
}

// ValidatePositive checks syntax and requires a value greater than zero.
func (decimal Decimal) ValidatePositive() error {
	if err := decimal.Validate(); err != nil {
		return err
	}
	for _, character := range string(decimal) {
		if character >= '1' && character <= '9' {
			return nil
		}
	}
	return fmt.Errorf("decimal value must be greater than zero")
}
