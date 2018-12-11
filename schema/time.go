package schema

import (
	"context"
	"errors"
	"time"
)

var (
	// Now is a field hook handler that returns the current time, to be used in
	// schema with OnInit and OnUpdate.
	Now = func(ctx context.Context, value interface{}) interface{} {
		return time.Now()
	}
	// CreatedField is a common schema field configuration for "created" fields.
	// It stores the creation date of the item.
	CreatedField = Field{
		Description: "The time at which the item has been inserted",
		Required:    true,
		ReadOnly:    true,
		OnInit:      Now,
		Sortable:    true,
		Validator:   &Time{},
	}

	// UpdatedField is a common schema field configuration for "updated" fields.
	// It stores the current date each time the item is modified.
	UpdatedField = Field{
		Description: "The time at which the item has been last updated",
		Required:    true,
		ReadOnly:    true,
		OnInit:      Now,
		OnUpdate:    Now,
		Sortable:    true,
		Validator:   &Time{},
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
	TimeLayouts []string // TimeLayouts is set of time layouts we want to validate.
	layouts     []string
}

// Compile the time formats.
func (v *Time) Compile(rc ReferenceChecker) error {
	if len(v.TimeLayouts) == 0 {
		// default layouts to all formats.
		v.layouts = formats
		return nil
	}
	// User specified list of time layouts.
	for _, layout := range v.TimeLayouts {
		v.layouts = append(v.layouts, string(layout))
	}
	return nil
}

func (v Time) parse(value interface{}) (interface{}, error) {
	if s, ok := value.(string); ok {
		for _, layout := range v.layouts {
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

// ValidateQuery implements schema.FieldQueryValidator interface
func (v Time) ValidateQuery(value interface{}) (interface{}, error) {
	return v.parse(value)
}

// Validate validates and normalize time based value.
func (v Time) Validate(value interface{}) (interface{}, error) {
	return v.parse(value)
}

func (v Time) get(value interface{}) (time.Time, error) {
	t, ok := value.(time.Time)
	if !ok {
		return t, errors.New("not a time")
	}
	return t, nil
}

// LessFunc implements the FieldComparator interface.
func (v Time) LessFunc() LessFunc {
	return v.less
}

func (v Time) less(value, other interface{}) bool {
	t, err1 := v.get(value)
	o, err2 := v.get(other)
	if err1 != nil || err2 != nil {
		return false
	}
	return t.Before(o)
}
