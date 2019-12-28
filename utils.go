package cloudns

import (
	"fmt"
	"strings"
)

type APIBool bool

func containsString(needle string, haystack []string) bool {
	for _, value := range haystack {
		if needle == value {
			return true
		}
	}

	return false
}

func (b APIBool) MarshalJSON() ([]byte, error) {
	if b == true {
		return []byte("1"), nil
	}

	return []byte("0"), nil
}

func (b *APIBool) UnmarshalJSON(data []byte) error {
	stringValue := strings.ToLower(strings.Trim(string(data), "\""))
	if stringValue == "true" || stringValue == "1" {
		*b = true
	} else if stringValue == "false" || stringValue == "0" {
		*b = false
	} else {
		return fmt.Errorf("could not unmarshal boolean from invalid input: %s", stringValue)
	}

	return nil
}
