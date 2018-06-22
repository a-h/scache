package data

import (
	"errors"
	"net/url"
)

// ID uniquely identifies a piece of data.
type ID struct {
	Source string
	ID     string
}

// String returns the id of the data.
func (d ID) String() string {
	v := url.Values{}
	v.Set("s", d.Source)
	v.Set("id", d.ID)
	v.Set("t", "data.ID")
	return v.Encode()
}

// NewID creates a new ID, e.g. source = "mydb.mytable.id", id=12345
func NewID(source, id string) ID {
	return ID{
		Source: source,
		ID:     id,
	}
}

var (
	// ErrMalformed is the error returned when the ID cannot be parsed due to issues with URL parsing.
	ErrMalformed = errors.New("data.ID: unable to parse due to malformed escaping")
	// ErrNotDataID is the error returned when the "t" value in the URL isn't expected, meaning that the value,
	// while being a URL, isn't a data ID.
	ErrNotDataID = errors.New("data.ID: value is not a data ID")
)

// Parse parses an ID.
func Parse(s string) (id ID, err error) {
	var vals url.Values
	vals, err = url.ParseQuery(s)
	if err != nil {
		err = ErrMalformed
		return
	}
	if vals.Get("t") == "" {
		err = ErrNotDataID
		return
	}
	id.Source = vals.Get("s")
	id.ID = vals.Get("id")
	return
}
