package config

import (
	"encoding/json"
	"errors"
	"fmt"
)

// CopyPreference represents user preferences for copy behavior.
type CopyPreference struct {
	value string
}

// Predefined CopyPreferences (Strongly Typed)
var (
	AskEveryTime      = CopyPreference{"ask-every-time"}     // Always prompt before copying
	AskOnce           = CopyPreference{"ask-once"}           // Ask once per session
	CopyAutomatically = CopyPreference{"copy-automatically"} // Always copy without prompting
	DoNotCopy         = CopyPreference{"do-not-copy"}        // Never copy
)

// AllowedValues returns all valid enum options.
func AllowedValues() []CopyPreference {
	return []CopyPreference{AskEveryTime, AskOnce, CopyAutomatically, DoNotCopy}
}

// IsValid checks if the CopyPreference is valid.
func (c CopyPreference) IsValid() bool {
	for _, v := range AllowedValues() {
		if c == v {
			return true
		}
	}
	return false
}

// Default returns the recommended default value.
func DefaultCopyPreference() CopyPreference {
	return AskEveryTime
}

// String returns the string representation.
func (c CopyPreference) String() string {
	return c.value
}

// ParseCopyPreference safely converts a string to a CopyPreference.
func ParseCopyPreference(s string) (CopyPreference, error) {
	for _, v := range AllowedValues() {
		if v.value == s {
			return v, nil
		}
	}
	return CopyPreference{}, errors.New("invalid copy preference")
}

// JSON Serialization Support
func (c CopyPreference) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CopyPreference) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseCopyPreference(s)
	if err != nil {
		return err
	}
	*c = parsed
	return nil
} 