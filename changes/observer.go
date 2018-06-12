package changes

import (
	"bytes"
	"errors"
	"strings"

	"github.com/a-h/scache/data"

	"github.com/a-h/scache/expiry"
)

// Observer reads a stream for changes, handling state.
type Observer struct {
	s   StreamGetter
	pos expiry.StreamPosition
}

// StreamGetter defines the requirements for informing consumers of changes.
type StreamGetter interface {
	Get(from expiry.StreamPosition) (keys []string, to expiry.StreamPosition, err error)
}

// NewObserver creates a way of keeping up-to-date with a stream.
func NewObserver(s StreamGetter) Observer {
	return Observer{
		s:   s,
		pos: map[expiry.ShardID]expiry.SequenceNumber{},
	}
}

// Observe gets all changes to the stream.
func (o Observer) Observe() (op []data.ID, err error) {
	si, to, err := o.s.Get(o.pos)
	if err != nil {
		return
	}
	op = make([]data.ID, len(si))
	var errs []error
	for i, s := range si {
		id, err := data.Parse(s)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		op[i] = id
	}
	o.pos = to
	return nil, join(errs)
}

func join(errs []error) error {
	if errs == nil || len(errs) == 0 {
		return nil
	}
	var b bytes.Buffer
	for _, e := range errs {
		b.WriteString(e.Error() + ":")
	}
	return errors.New(strings.TrimSuffix(b.String(), ":"))
}

// Reset sets the position of the stream to the latest message. Used when no data is cached, so being
// notified of changes isn't required.
func (o Observer) Reset() {
	o.pos = map[expiry.ShardID]expiry.SequenceNumber{}
}
