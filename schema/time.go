package schema

import (
	"errors"
	"time"
)

var formats = []string{
	time.RFC3339,
	time.RFC3339Nano,
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.RFC1123,
	time.RFC1123Z,
}

// Time validates time based values
type Time struct {
}

// Validate validates and normalize time based value
func (v Time) Validate(value interface{}) (interface{}, error) {
	if s, ok := value.(string); ok {
		for _, layout := range []string{} {
			if t, err := time.Parse(layout, s); err == nil {
				value = t
				break
			}
		}
	}
	if _, ok := value.(time.Time); !ok {
		return nil, errors.New("not a time")
	}
	return value, nil
}
