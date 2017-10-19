package query

// Window defines a view on the resulting payload.
type Window struct {
	// Offset is the 0 based index of the item in the result set to start the
	// window at.
	Offset int

	// Limit is the maximum number of items to return in the result set. A value
	// lower than 0 means no limit.
	Limit int
}

// Page creates a Window using pagination.
func Page(page, perPage, skip int) *Window {
	if perPage < 0 {
		if skip > 0 {
			return &Window{Offset: skip, Limit: perPage}
		}
		return nil
	}
	if page < 1 {
		page = 1
	}
	if skip < 0 {
		skip = 0
	}
	return &Window{
		Offset: (page-1)*perPage + skip,
		Limit:  perPage,
	}
}
