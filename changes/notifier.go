package changes

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

// Notify notifies consumers of changes to data items.
func (n Notifier) Notify(changesTo ...Observable) error {
	keys := make([]string, len(changesTo))
	for i, changed := range changesTo {
		keys[i] = changed.ObservableID().String()
	}
	return n.s.Put(keys)
}
