package httpsfv

import "testing"

func TestDecodeError(t *testing.T) {
	_, err := UnmarshalItem([]string{"invalid-é"})

	if err.Error() != "unmarshal error: character 8" {
		t.Error("invalid error")
	}

	_, err = UnmarshalItem([]string{`"é"`})
	if err.Error() != "invalid string format: character 2" {
		t.Error("invalid error")
	}

	if err.(*UnmarshalError).Unwrap().Error() != "invalid string format" {
		t.Error("invalid wrapped error")
	}
}
