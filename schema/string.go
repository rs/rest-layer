package schema

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// String validates string based values
type String struct {
	re      *regexp.Regexp
	Regexp  string
	Allowed []string
	MaxLen  int
	MinLen  int
}

// Compile compiles and validate regexp if any.
func (v *String) Compile(rc ReferenceChecker) (err error) {
	if v.Regexp != "" {
		// Compile and cache regexp, report any compilation error.
		if v.re, err = regexp.Compile(v.Regexp); err != nil {
			err = fmt.Errorf("invalid regexp: %s", err)
		}
	}
	return
}

// ValidateQuery implements schema.FieldQueryValidator interface
func (v String) ValidateQuery(value interface{}) (interface{}, error) {
	s, ok := value.(string)
	if !ok {
		return nil, errors.New("not a string")
	}
	return s, nil
}

// Validate validates and normalize string based value.
func (v String) Validate(value interface{}) (interface{}, error) {
	// Pre-check that compilation was successful.
	if v.Regexp != "" && v.re == nil {
		return nil, errors.New("not successfully compiled")
	}

	s, ok := value.(string)
	if !ok {
		return nil, errors.New("not a string")
	}
	l := len(s)
	if l < v.MinLen {
		return nil, fmt.Errorf("is shorter than %d", v.MinLen)
	}
	if v.MaxLen > 0 && l > v.MaxLen {
		return nil, fmt.Errorf("is longer than %d", v.MaxLen)
	}
	if len(v.Allowed) > 0 {
		found := false
		for _, allowed := range v.Allowed {
			if s == allowed {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("not one of [%s]", strings.Join(v.Allowed, ", "))
		}
	}
	if v.Regexp != "" {
		if !v.re.MatchString(s) {
			return nil, fmt.Errorf("does not match %s", v.Regexp)
		}
	}
	return s, nil
}
