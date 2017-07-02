package resource

// Conf defines the configuration for a given resource.
type Conf struct {
	// AllowedModes is the list of Mode allowed for the resource.
	AllowedModes []Mode
	// DefaultPageSize defines a default number of items per page. By default,
	// no default page size is set resulting in no pagination if no `limit`
	// parameter is provided.
	PaginationDefaultLimit int
	// ForceTotal controls how total number of items on list request is
	// computed. By default (TotalOptIn), if the total cannot be computed by the
	// storage handler for free, no total metadata is returned until the user
	// explicitly request it using the total=1 query-string parameter. Note that
	// if the storage cannot compute the total and does not implement the
	// resource.Counter interface, a "not implemented" error is returned.
	//
	// The TotalAlways mode always force the computation of the total (make sure
	// the storage either compute the total on Find or implement the
	// resource.Counter interface.
	//
	// TotalDenied prevents the user from requesting the total.
	ForceTotal ForceTotalMode
}

// ForceTotalMode defines Conf.ForceTotal modes.
type ForceTotalMode int

const (
	// TotalOptIn allows the end-user to opt-in to forcing the total count by
	// adding the total=1 query-string parameter.
	TotalOptIn ForceTotalMode = iota
	// TotalAlways always force the total number of items on list requests
	TotalAlways
	// TotalDenied disallows forcing of the total count, and returns an error if
	// total=1 is supplied, and the total count is not provided by the Storer's
	// Find method.
	TotalDenied
)

// Mode defines CRUDL modes to be used with Conf.AllowedModes.
type Mode int

const (
	// Create mode represents the POST method on a collection URL or the PUT
	// method on a _non-existing_ item URL.
	Create Mode = iota
	// Read mode represents the GET method on an item URL.
	Read
	// Update mode represents the PATCH on an item URL.
	Update
	// Replace mode represents the PUT methods on an existing item URL.
	Replace
	// Delete mode represents the DELETE method on an item URL.
	Delete
	// Clear mode represents the DELETE method on a collection URL.
	Clear
	// List mode represents the GET method on a collection URL.
	List
)

var (
	// ReadWrite is a shortcut for all modes.
	ReadWrite = []Mode{Create, Read, Update, Replace, Delete, List, Clear}
	// ReadOnly is a shortcut for Read and List modes.
	ReadOnly = []Mode{Read, List}
	// WriteOnly is a shortcut for Create, Update, Delete modes.
	WriteOnly = []Mode{Create, Update, Replace, Delete, Clear}

	// DefaultConf defines a configuration with some sensible default parameters.
	// Mode is read/write and default pagination limit is set to 20 items.
	DefaultConf = Conf{
		AllowedModes:           ReadWrite,
		PaginationDefaultLimit: 20,
	}
)

// IsModeAllowed returns true if the provided mode is allowed in the configuration.
func (c Conf) IsModeAllowed(mode Mode) bool {
	for _, m := range c.AllowedModes {
		if m == mode {
			return true
		}
	}
	return false
}
