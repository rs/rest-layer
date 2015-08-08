package schema

import (
	"errors"
	"fmt"
	"net/url"
)

// URL validates URLs values
type URL struct {
}

// Validate validates URL values
func (v URL) Validate(value interface{}) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return nil, errors.New("invalid type")
	}
	if _, err := url.Parse(str); err != nil {
		return nil, fmt.Errorf("invalid URL: %s", err.Error())
	}
	return value, nil
}
