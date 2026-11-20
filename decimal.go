package httpsfv

import (
	"errors"
	"io"
	"math"
	"strconv"
	"strings"
)

const maxDecDigit = 3

// ErrInvalidDecimal is returned when a decimal is invalid.
var ErrInvalidDecimal = errors.New("the integer portion is larger than 12 digits: invalid decimal")

// marshalDecimal serializes as defined in
// https://httpwg.org/specs/rfc9651.html#ser-decimal.
//
// Rounding is done on the shortest decimal form of d, not on its binary
// value: 0.5015 is stored as 0.50149999..., but it must serialize as 0.502.
//
// TODO(dunglas): add support for decimal float type when one will be available
// (https://github.com/golang/go/issues/19787)
func marshalDecimal(b io.StringWriter, d float64) error {
	if math.IsNaN(d) || math.IsInf(d, 0) {
		return ErrInvalidDecimal
	}

	s := strconv.FormatFloat(math.Abs(d), 'f', -1, 64)
	intPart, frac, _ := strings.Cut(s, ".")

	if len(frac) > maxDecDigit {
		digits := []byte(intPart + frac[:maxDecDigit])
		rest := frac[maxDecDigit:]

		// Round half to even.
		if rest[0] > '5' || (rest[0] == '5' && (strings.TrimRight(rest[1:], "0") != "" || (digits[len(digits)-1]-'0')%2 == 1)) {
			i := len(digits) - 1
			for ; i >= 0 && digits[i] == '9'; i-- {
				digits[i] = '0'
			}

			if i < 0 {
				digits = append([]byte{'1'}, digits...)
			} else {
				digits[i]++
			}
		}

		intPart = string(digits[:len(digits)-maxDecDigit])
		frac = string(digits[len(digits)-maxDecDigit:])
	}

	if len(intPart) > 12 {
		return ErrInvalidDecimal
	}

	frac = strings.TrimRight(frac, "0")
	// Ensure at least one digit after the decimal point.
	if frac == "" {
		frac = "0"
	}

	if d < 0 && (intPart != "0" || frac != "0") {
		intPart = "-" + intPart
	}

	_, err := b.WriteString(intPart + "." + frac)

	return err
}

func parseDecimal(s *scanner, decSepOff int, str string, neg bool) (float64, error) {
	if decSepOff == s.off-1 {
		return 0, &UnmarshalError{s.off, ErrInvalidDecimalFormat}
	}

	if len(s.data[decSepOff+1:s.off]) > maxDecDigit {
		return 0, &UnmarshalError{s.off, ErrNumberOutOfRange}
	}

	i, err := strconv.ParseFloat(str, 64)
	if err != nil {
		// Should never happen
		return 0, &UnmarshalError{s.off, err}
	}

	if neg {
		i = -i
	}

	return i, nil
}
