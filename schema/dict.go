package schema

import (
	"errors"
	"fmt"
)

// Dict validates array values
type Dict struct {
	// KeysValidator is the validator to apply on dict keys
	KeysValidator FieldValidator
	// ValuesValidator is the validator to apply on dict values
	ValuesValidator FieldValidator
	// MinLen defines the minimum number of fields (default 0)
	MinLen int
	// MaxLen defines the maximum number of fields (default no limit)
	MaxLen int
}

// Compile implements Compiler interface
func (v *Dict) Compile() (err error) {
	if c, ok := v.KeysValidator.(Compiler); ok {
		if err = c.Compile(); err != nil {
			return
		}
	}
	if c, ok := v.ValuesValidator.(Compiler); ok {
		if err = c.Compile(); err != nil {
			return
		}
	}
	return
}

// Validate implements FieldValidator interface
func (v Dict) Validate(value interface{}) (interface{}, error) {
	dict, ok := value.(map[string]interface{})
	if !ok {
		return nil, errors.New("not a dict")
	}
	dest := map[string]interface{}{}
	for key, val := range dict {
		if v.KeysValidator != nil {
			nkey, err := v.KeysValidator.Validate(key)
			if err != nil {
				return nil, fmt.Errorf("invalid key `%s': %s", key, err)
			}
			if key, ok = nkey.(string); !ok {
				return nil, errors.New("key validator does not return string")
			}
		}
		if v.ValuesValidator != nil {
			var err error
			val, err = v.ValuesValidator.Validate(val)
			if err != nil {
				return nil, fmt.Errorf("invalid value for key `%s': %s", key, err)
			}
		}
		dest[key] = val
	}
	l := len(dest)
	if l < v.MinLen {
		return nil, fmt.Errorf("has fewer properties than %d", v.MinLen)
	}
	if v.MaxLen > 0 && l > v.MaxLen {
		return nil, fmt.Errorf("has more properties than %d", v.MaxLen)
	}
	return dest, nil
}
