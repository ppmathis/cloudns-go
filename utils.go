package cloudns

import (
	"fmt"
	"strings"
)

// APIBool is a custom type representing the way how ClouDNS treats booleans in their API, as they usually appear as
// 1 or 0 (as a number or a string) instead of actual JSON booleans
type APIBool bool

func containsString(needle string, haystack []string) bool {
	for _, value := range haystack {
		if needle == value {
			return true
		}
	}

	return false
}

// MarshalJSON converts a APIBool into a 0 or 1 as a number according to the ClouDNS API docs
func (b APIBool) MarshalJSON() ([]byte, error) {
	if b {
		return []byte("1"), nil
	}

	return []byte("0"), nil
}

// UnmarshalJSON converts a boolean from the ClouDNS API into a sanitized Go boolean
func (b *APIBool) UnmarshalJSON(data []byte) error {
	stringValue := strings.ToLower(strings.Trim(string(data), "\""))
	if stringValue == "true" || stringValue == "1" {
		*b = true
	} else if stringValue == "false" || stringValue == "0" || stringValue == "" {
		*b = false
	} else {
		return fmt.Errorf("could not unmarshal boolean from invalid input: %s", stringValue)
	}

	return nil
}
