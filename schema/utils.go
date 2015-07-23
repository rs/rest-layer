package schema

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"time"
)

var (
	// Now is a field hook handler that returns the current time, to be used in
	// schema with OnInit and OnUpdate
	Now = func(value interface{}) interface{} {
		return time.Now()
	}

	// NewID is a field hook handler that generates a new unique id if none exist,
	// to be used in schema with OnInit
	NewID = func(value interface{}) interface{} {
		if value == nil {
			r := make([]byte, 128)
			rand.Read(r)
			value = fmt.Sprintf("%x", md5.Sum(r))
		}
		return value
	}

	// IDField is a common schema field configuration that generate an UUID for new item id.
	IDField = Field{
		Required: true,
		ReadOnly: true,
		OnInit:   &NewID,
		Validator: &String{
			Regexp: "^[0-9a-f]{32}$",
		},
	}

	// CreatedField is a common schema field configuration for "created" fields. It stores
	// the creation date of the item.
	CreatedField = Field{
		Required:  true,
		ReadOnly:  true,
		OnInit:    &Now,
		Validator: &Time{},
	}

	// UpdatedField is a common schema field configuration for "updated" fields. It stores
	// the current date each time the item is modified.
	UpdatedField = Field{
		Required:  true,
		ReadOnly:  true,
		OnInit:    &Now,
		OnUpdate:  &Now,
		Validator: &Time{},
	}
)
