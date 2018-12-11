package schema

import (
	"fmt"
	"sort"
	"strings"
)

// ErrorMap contains a map of errors by field name.
type ErrorMap map[string][]interface{}

// Error implements the built-in error interface.
func (err ErrorMap) Error() string {
	errs := make([]string, 0, len(err))
	for key := range err {
		errs = append(errs, key)
	}
	sort.Strings(errs)
	for i, key := range errs {
		errs[i] = fmt.Sprintf("%s is %s", key, err[key])
	}
	return strings.Join(errs, ", ")
}

// Merge copies all errors from other into err.
func (err ErrorMap) Merge(other ErrorMap) {
	for k, v := range other {
		err[k] = append(err[k], v...)
	}
}

// ErrorSlice contains a concatenation of several errors.
type ErrorSlice []error

// Append adds an error to err and returns a new slice if others is not nil. If
// other is another ErrorSlice it is extended so that all elements are appended.
func (err ErrorSlice) Append(other error) ErrorSlice {
	switch et := other.(type) {
	case nil:
		// don't append nil errors.
	case ErrorSlice:
		// Merge error slices.
		err = append(err, et...)
	default:
		err = append(err, et)
	}
	return err
}

func (err ErrorSlice) Error() string {
	sl := make([]string, 0, len(err))
	for _, err := range err {
		sl = append(sl, err.Error())
	}
	return strings.Join(sl, ", ")
}
