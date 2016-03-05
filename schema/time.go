package schema

import (
	"errors"
	"time"

	"golang.org/x/net/context"
)

var (
	// Now is a field hook handler that returns the current time, to be used in
	// schema with OnInit and OnUpdate.
	Now = func(ctx context.Context, value interface{}) interface{} {
		return time.Now()
	}
	// CreatedField is a common schema field configuration for "created" fields. It stores
	// the creation date of the item.
	CreatedField = Field{
		Required:  true,
		ReadOnly:  true,
		OnInit:    &Now,
		Sortable:  true,
		Validator: &Time{},
	}

	// UpdatedField is a common schema field configuration for "updated" fields. It stores
	// the current date each time the item is modified.
	UpdatedField = Field{
		Required:  true,
		ReadOnly:  true,
		OnInit:    &Now,
		OnUpdate:  &Now,
		Sortable:  true,
		Validator: &Time{},
	}

	formats = []string{
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
)

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
