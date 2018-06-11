package expiry

import "time"

// StreamData is the data stored in each cache invalidation stream.
type StreamData struct {
	// The data keys which have been invalidated.
	Keys []string `json:"keys"`
	// The time that they were invalidated (client-side).
	Time time.Time `json:"ts"`
}

// NewStreamData creates a StreamData record.
func NewStreamData(keys []string) StreamData {
	return StreamData{
		Keys: keys,
		Time: time.Now().UTC(),
	}
}
