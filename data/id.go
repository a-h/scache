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
	v.Set("source", d.Source)
	v.Set("id", d.ID)
	return v.Encode()
}

// NewID creates a new ID, e.g. source = "mydb.mytable.id", id=12345
func NewID(source, id string) ID {
	return ID{
		Source: source,
		ID:     id,
	}
}

// Parse parses an ID.
func Parse(s string) (id ID, err error) {
	var vals url.Values
	vals, err = url.ParseQuery(s)
	if err != nil {
		err = errors.New("unable to parse ID")
		return
	}
	id.Source = vals.Get("source")
	id.ID = vals.Get("id")
	return
}
