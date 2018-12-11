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

// Time validates time based values.
type Time struct {
	// TimeLayouts is set of time layouts we want to validate.
	TimeLayouts []string

	// Truncate set to truncate all time-stamps to a given precision. Truncate
	// always happens according to the Go zero-time (1 Jan, year 1 at midnight).
	Truncate time.Duration

	// Location, if set, converts all times to the given time-zone.
	Location *time.Location

	layouts []string
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

func (v Time) parse(value interface{}) (time.Time, error) {
	switch vt := value.(type) {
	case time.Time:
		return vt, nil
	case string:
		for _, layout := range v.layouts {
			if t, err := time.Parse(layout, vt); err == nil {
				return t, nil
			}
		}
	}

	return time.Time{}, errors.New("not a time")
}

// ValidateQuery implements the FieldQueryValidator interface.
func (v Time) ValidateQuery(value interface{}) (interface{}, error) {
	t, err := v.parse(value)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// Validate validates and normalize a time based value.
func (v Time) Validate(value interface{}) (interface{}, error) {
	t, err := v.parse(value)
	if err != nil {
		return nil, err
	}

	// We always call Truncate, even if v.Truncate is 0, so that the monotonic
	// time component is always dropped.
	t = t.Truncate(v.Truncate)

	if v.Location != nil {
		t = t.In(v.Location)
	}

	return t, nil
}

func (v Time) get(value interface{}) (time.Time, error) {
	t, ok := value.(time.Time)
	if !ok {
		return t, errors.New("not a time")
	}
	return t, nil
}

// Less implements schema.FieldComparator interface
func (v Time) Less(value, other interface{}) bool {
	t, err := v.get(value)
	o, err1 := v.get(other)
	if err != nil || err1 != nil {
		return false
	}
	return t.Before(o)
}
