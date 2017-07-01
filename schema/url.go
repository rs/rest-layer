package schema

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// URL validates URLs values.
type URL struct {
	AllowRelative  bool
	AllowLocale    bool
	AllowNonHTTP   bool
	AllowedSchemes []string
}

// Validate validates URL values.
func (v URL) Validate(value interface{}) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return nil, errors.New("invalid type")
	}
	u, err := url.Parse(str)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %s", err.Error())
	}
	if !v.AllowRelative && !u.IsAbs() {
		return nil, errors.New("is relative URL")
	}
	if !v.AllowLocale && strings.IndexByte(u.Host, '.') == -1 {
		return nil, errors.New("invalid domain")
	}
	if len(v.AllowedSchemes) > 0 {
		found := false
		for _, scheme := range v.AllowedSchemes {
			if scheme == u.Scheme {
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("invalid scheme")
		}
	} else if !v.AllowNonHTTP && u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("invalid scheme")
	}
	return u.String(), nil
}
