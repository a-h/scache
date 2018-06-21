package user

import (
	"encoding/json"
	"net/http"

	"github.com/a-h/scache"

	"github.com/gorilla/mux"
)

// DatabaseGetter gets Users.
type DatabaseGetter func(id string) (User, error)

// DatabasePutter stores users.
type DatabasePutter func(u User) error

// Handler handles GET and POST of User data.
type Handler struct {
	GetUser DatabaseGetter
	PutUser DatabasePutter
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.Post(w, r)
		return
	}
	h.Get(w, r)
}

// Get is the HTTP handler for the GET method.
func (h Handler) Get(w http.ResponseWriter, r *http.Request) {
	// Grab the UserID from the request fields.
	userID, ok := mux.Vars(r)["userID"]

	// Convert the bare userID value (e.g. 12345) into a qualified data.ID
	// (one which includes the database or other data source name).
	dataID := DataID(userID)

	// Get the data from the cache if possible. By this point, the HTTP middleware has already
	// handled removing expired entries from the cache.
	var u User
	var err error
	ok = scache.Get(r, dataID, &u)
	if !ok {
		// If the data isn't in the cache, we need to go back to the data source.
		u, err = h.GetUser(userID)
		if err != nil {
			http.Error(w, "failed to get user from DB", http.StatusInternalServerError)
			return
		}
	}

	e := json.NewEncoder(w)
	e.Encode(u)
}

// Post is the HTTP handler for the POST method.
func (h Handler) Post(w http.ResponseWriter, r *http.Request) {
	// Grab the UserID from the request fields.
	userID, ok := mux.Vars(r)["userID"]
	if !ok {
		http.Error(w, "userID not found in path", http.StatusBadRequest)
		return
	}

	// Convert the bare userID value (e.g. 12345) into a qualified data.ID
	// (one which includes the database or other data source name).
	dataID := DataID(userID)

	// Decode the User from the JSON body of the HTTP request.
	var u User

	d := json.NewDecoder(r.Body)
	err := d.Decode(&u)
	if err != nil {
		http.Error(w, "failed to decode JSON payload", http.StatusUnprocessableEntity)
		return
	}
	defer r.Body.Close()

	// Attempt to update the data.
	err = h.PutUser(u)
	if err != nil {
		http.Error(w, "failed to store user", http.StatusInternalServerError)
		return
	}

	scache.Invalidate(r, dataID)
	w.Write([]byte(`{ "ok": "true" }`))
}
