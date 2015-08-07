package rest

// Conf defines the configuration for a given resource
type Conf struct {
	// AllowedModes is the list of Mode allowed for the resource.
	AllowedModes []Mode
	// DefaultPageSize defines a default number of items per page. By defaut,
	// no default page size is set resulting in no pagination if no `limit` parameter
	// is provided.
	PaginationDefaultLimit int
}

// Mode defines CRUDL modes to be used with Conf.AllowedModes.
type Mode int

const (
	// Create mode represents the POST method on a collection URL or the PUT method
	// on a _non-existing_ item URL.
	Create Mode = iota
	// Read mode represents the GET method on an item URL
	Read
	// Update mode represents the PATCH on an item URL.
	Update
	// Replace mode represents the PUT methods on an existing item URL.
	Replace
	// Delete mode represents the DELETE method on an item URL.
	Delete
	// Clear mode represents the DELETE method on a collection URL
	Clear
	// List mode represents the GET method on a collection URL.
	List
)

var (
	// ReadWrite is a shortcut for all modes
	ReadWrite = []Mode{Create, Read, Update, Replace, Delete, List, Clear}
	// ReadOnly is a shortcut for Read and List modes
	ReadOnly = []Mode{Read, List}
	// WriteOnly is a shortcut for Create, Update, Delete modes
	WriteOnly = []Mode{Create, Update, Replace, Delete, Clear}

	// DefaultConf defines a configuration with some sensible default parameters.
	// Mode is read/write and default pagination limit is set to 20 items.
	DefaultConf = Conf{
		AllowedModes:           ReadWrite,
		PaginationDefaultLimit: 20,
	}
)

func (c Conf) isModeAllowed(mode Mode) bool {
	for _, m := range c.AllowedModes {
		if m == mode {
			return true
		}
	}
	return false
}
