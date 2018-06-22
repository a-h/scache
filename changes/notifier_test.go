package changes

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/a-h/scache/data"
)

type MockStreamPutter struct {
	PutFuncs     []func(keys []string) error
	PutCallCount int
}

func (msp *MockStreamPutter) Put(keys []string) error {
	defer func() { msp.PutCallCount++ }()
	return msp.PutFuncs[msp.PutCallCount](keys)
}

type UserRecord struct {
	UserID int
	Name   string
	Email  string
}

func (u UserRecord) ObservableID() data.ID {
	return data.NewID("db.users.userid", strconv.Itoa(u.UserID))
}

func TestNotifier(t *testing.T) {
	id1 := data.NewID("db.users.userid", "1")
	id2 := data.NewID("db.users.userid", "2")
	id3 := data.NewID("db.users.userid", "3")
	id4 := data.NewID("db.users.userid", "4")

	expected1 := []string{id1.String(), id2.String()}
	expected2 := []string{id3.String(), id4.String()}

	putter := &MockStreamPutter{
		PutFuncs: []func(keys []string) error{
			func(keys []string) (err error) {
				if !reflect.DeepEqual(keys, expected1) {
					t.Errorf("put #1: expected %v, got %v", expected1, keys)
				}
				return
			},
			func(keys []string) (err error) {
				if !reflect.DeepEqual(keys, expected2) {
					t.Errorf("put #2: expected %v, got %v", expected2, keys)
				}
				return
			},
		},
	}

	n := NewNotifier(putter)

	records := []Observable{
		UserRecord{
			UserID: 1,
			Name:   "Charles",
			Email:  "charles@example.com",
		},
		UserRecord{
			UserID: 2,
			Name:   "William",
			Email:  "william@example.com",
		},
	}

	err := n.NotifyObservablesChanged(records...)
	if err != nil {
		t.Errorf("unexpected error on NotifyObservablesChanged: %v", err)
	}
	err = n.NotifyDataChanged(id3, id4)
	if err != nil {
		t.Errorf("unexpected error on NotifyDataChanged: %v", err)
	}
}
