package changes

import (
	"errors"
	"reflect"
	"testing"

	"github.com/a-h/scache/data"
	"github.com/a-h/scache/expiry"
)

type MockStreamGetter struct {
	GetFuncs     []func(from expiry.StreamPosition) (keys []string, to expiry.StreamPosition, err error)
	GetCallCount int
}

func (msg *MockStreamGetter) Get(from expiry.StreamPosition) (keys []string, to expiry.StreamPosition, err error) {
	defer func() { msg.GetCallCount++ }()
	return msg.GetFuncs[msg.GetCallCount](from)
}

func TestThatTheStreamPositionIsUpdated(t *testing.T) {
	id1 := data.NewID("db1.table1.id", "1")
	id2 := data.NewID("db1.table1.id", "2")
	id3 := data.NewID("db1.table1.id", "3")
	id4 := data.NewID("db1.table1.id", "4")
	id5 := data.NewID("db1.table1.id", "5")
	id6 := data.NewID("db1.table1.id", "6")

	getter := &MockStreamGetter{
		GetFuncs: []func(from expiry.StreamPosition) (keys []string, to expiry.StreamPosition, err error){
			func(from expiry.StreamPosition) (keys []string, to expiry.StreamPosition, err error) {
				if len(from) != 0 {
					t.Error("shouldn't have a stream position on the first call to the stream")
				}
				keys = []string{id1.String(), id2.String(), id3.String()}
				to = map[expiry.ShardID]expiry.SequenceNumber{
					"shard_1": "2",
					"shard_2": "1",
				}
				return
			},
			func(from expiry.StreamPosition) (keys []string, to expiry.StreamPosition, err error) {
				if len(from) != 2 {
					t.Fatalf("should have positions for two shards, but have: %v", len(from))
				}
				if from["shard_1"] != "2" {
					t.Errorf("on second stream read, expected shard_1 to start at position '2', but got '%v'", from["shard_1"])
				}
				if from["shard_2"] != "1" {
					t.Errorf("on second stream read, expected shard_2 to start at position '1', but got '%v'", from["shard_2"])
				}
				keys = []string{id4.String(), id5.String(), id6.String()}
				to = map[expiry.ShardID]expiry.SequenceNumber{
					"shard_1": "7",
					"shard_2": "8",
				}
				return
			},
			func(from expiry.StreamPosition) (keys []string, to expiry.StreamPosition, err error) {
				if len(from) != 0 {
					t.Fatalf("after calling reset, the stream position should have been reset, but got: %v", from)
				}
				to = map[expiry.ShardID]expiry.SequenceNumber{
					"shard_1": "9",
					"shard_2": "10",
				}
				return
			},
		},
	}

	o := NewObserver(getter)
	// Read from the latest position.
	ids, err := o.Observe()
	if err != nil {
		t.Fatalf("unexpected error observing stream (#1): %v", err)
	}
	expected := []data.ID{id1, id2, id3}
	if !reflect.DeepEqual(ids, expected) {
		t.Errorf("after first observation, expected %v, got: %v", expected, ids)
	}
	// Continue reading.
	ids, err = o.Observe()
	if err != nil {
		t.Fatalf("unexpected error observing stream (#2): %v", err)
	}
	expected = []data.ID{id4, id5, id6}
	if !reflect.DeepEqual(ids, expected) {
		t.Errorf("after second observation, expected %v, got: %v", expected, ids)
	}
	// Reset the stream to the latest data and read again.
	o.Reset()
	_, err = o.Observe()
	if err != nil {
		t.Fatalf("unexpected error observing stream (#3): %v", err)
	}
}

func TestThatObservervationMergesParseErrors(t *testing.T) {
	id1 := data.NewID("db1.table1.id", "1")
	id2 := data.NewID("db1.table1.id", "2")
	id3 := data.NewID("db1.table1.id", "3")

	getter := &MockStreamGetter{
		GetFuncs: []func(from expiry.StreamPosition) (keys []string, to expiry.StreamPosition, err error){
			func(from expiry.StreamPosition) (keys []string, to expiry.StreamPosition, err error) {
				keys = []string{id1.String(), "something we can't parse", id2.String(), "some more bad stuff", id3.String()}
				return
			},
		},
	}

	o := NewObserver(getter)
	// Read from the latest position.
	ids, err := o.Observe()
	expectedErr := "observer: data.ID: value is not a data ID 'something we can't parse', data.ID: value is not a data ID 'some more bad stuff'"
	if err == nil || err.Error() != expectedErr {
		t.Fatalf("unexpected error observing stream: %v", err)
	}
	expected := []data.ID{id1, id2, id3}
	if !reflect.DeepEqual(ids, expected) {
		t.Errorf("after first observation, expected %v, got: %v", expected, ids)
	}
}

func TestObservationErrors(t *testing.T) {
	getter := &MockStreamGetter{
		GetFuncs: []func(from expiry.StreamPosition) (keys []string, to expiry.StreamPosition, err error){
			func(from expiry.StreamPosition) (keys []string, to expiry.StreamPosition, err error) {
				err = errors.New("network error")
				return
			},
		},
	}

	o := NewObserver(getter)
	_, err := o.Observe()
	expectedErr := "observer: could not get from stream: network error"
	if err == nil || err.Error() != expectedErr {
		t.Fatalf("unexpected error observing stream: %v", err)
	}
}
