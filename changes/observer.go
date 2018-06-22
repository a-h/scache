package changes

import (
	"bytes"
	"errors"
	"strings"
	"sync"

	"github.com/a-h/scache/data"

	"github.com/a-h/scache/expiry"
)

// Observer reads a stream for changes, handling state.
type Observer struct {
	s     StreamGetter
	pos   expiry.StreamPosition
	mutex sync.Mutex
}

// StreamGetter defines the requirements for informing consumers of changes.
type StreamGetter interface {
	Get(from expiry.StreamPosition) (keys []string, to expiry.StreamPosition, err error)
}

// NewObserver creates a way of keeping up-to-date with a stream.
func NewObserver(s StreamGetter) *Observer {
	return &Observer{
		s:     s,
		pos:   map[expiry.ShardID]expiry.SequenceNumber{},
		mutex: sync.Mutex{},
	}
}

// Observe gets all changes to the stream since the last call.
func (o *Observer) Observe() (op []data.ID, err error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	si, to, err := o.s.Get(o.pos)
	if err != nil {
		err = errors.New("observer: could not get from stream: " + err.Error())
		return
	}
	var errs []error
	for _, s := range si {
		id, parseErr := data.Parse(s)
		if parseErr != nil {
			e := errors.New(parseErr.Error() + " '" + s + "'")
			errs = append(errs, e)
			continue
		}
		op = append(op, id)
	}
	o.pos = to
	if errs != nil {
		err = join(errs)
	}
	return
}

func join(errs []error) (err error) {
	var b bytes.Buffer
	for _, e := range errs {
		b.WriteString(e.Error() + ", ")
	}
	if b.Len() > 0 {
		err = errors.New("observer: " + strings.TrimSuffix(b.String(), ", "))
	}
	return
}

// Reset sets the position of the stream to the latest message. Used when no data is cached, so being
// notified of changes isn't required.
func (o *Observer) Reset() {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.pos = map[expiry.ShardID]expiry.SequenceNumber{}
}
