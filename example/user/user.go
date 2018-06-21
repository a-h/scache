package user

import "github.com/a-h/scache/data"

// UserDataSource is a constant which defines where a datum comes from.
const UserDataSource = "dynamo.scache_user_table.id"

// DataID converts a User ID (e.g. 12345) into a data.ID, e.g. (source="dynamo.scache_user_table.id"&id="12345").
func DataID(id string) data.ID {
	return data.NewID(UserDataSource, id)
}

// User database record.
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ObservableID provides a key which uniquely identifies the data.
func (u User) ObservableID() data.ID {
	return DataID(u.ID)
}
