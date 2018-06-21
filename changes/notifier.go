package changes

import "github.com/a-h/scache/data"

// Notifier notifies listeners of changes.
type Notifier struct {
	s StreamPutter
}

// StreamPutter defines the requirements for informing consumers of changes.
type StreamPutter interface {
	Put(keys []string) error
}

// NewNotifier creates a way of notifying consumers of changes.
func NewNotifier(s StreamPutter) Notifier {
	return Notifier{
		s: s,
	}
}

// NotifyObservablesChanged notifies consumers of changes to data items.
func (n Notifier) NotifyObservablesChanged(changesTo ...Observable) error {
	keys := make([]string, len(changesTo))
	for i, changed := range changesTo {
		keys[i] = changed.ObservableID().String()
	}
	return n.s.Put(keys)
}

// NotifyDataChanged notifies consumers of changes to data items.
func (n Notifier) NotifyDataChanged(changed ...data.ID) error {
	keys := make([]string, len(changed))
	for i, id := range changed {
		keys[i] = id.String()
	}
	return n.s.Put(keys)
}
