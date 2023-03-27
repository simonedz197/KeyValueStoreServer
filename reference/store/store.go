package store

import (
	"container/list"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// LRU store.
type LRU struct {
	Override bool

	Depth int
	Data  map[string]*list.Element
	Order *list.List

	sync.Mutex
}

// ErrNotOwner is returned if an attempt is made to access a key the user does
// not own.
var ErrNotOwner = errors.New("not owner")

// ErrNotFound is returned if a key was not found.
var ErrNotFound = errors.New("not found")

const defaultDepth = 100

// NewLRU returns an LRU store of the given depth. A depth of 0 means unbounded.
func NewLRU(depth int) *LRU {
	initial := defaultDepth

	if depth > 0 {
		initial = depth
	}

	return &LRU{
		Depth: depth,
		Data:  make(map[string]*list.Element, initial),
		Order: list.New(),
	}
}

// Get a value from the store, returning an error if the key is not found.
func (l *LRU) Get(key string) (string, error) {
	l.Lock()
	defer l.Unlock()

	element, ok := l.Data[key]

	if !ok {
		return "", fmt.Errorf("get: key %q: %w", key, ErrNotFound)
	}

	entry, _ := element.Value.(*Entry)
	entry.Reads++
	entry.Timestamp = time.Now()

	l.Order.MoveToFront(element)

	return entry.Value, nil
}

// Put a value into the store. If the owner doesn't have permission to update an
// entry then an error is returned. Successful updates will trim the store to
// Size.
func (l *LRU) Put(key, value, user string) error {
	var entry *Entry

	l.Lock()
	defer l.Unlock()

	element, ok := l.Data[key]

	if ok {
		entry, _ = element.Value.(*Entry)

		if !Authorised(user, entry.Owner, l.Override) {
			return fmt.Errorf("put: %q %w of %q", user, ErrNotOwner, key)
		}

		entry.Value = value
		entry.Writes++
		entry.Timestamp = time.Now()

		l.Order.MoveToFront(element)
	} else {
		entry = &Entry{Key: key, Value: value, Owner: user,
			Writes: 1, Timestamp: time.Now()}
		element = l.Order.PushFront(entry)
		l.Data[key] = element
	}

	if l.Depth == 0 {
		return nil
	}

	for len(l.Data) > l.Depth {
		last := l.Order.Back()
		remove, _ := last.Value.(*Entry)

		logrus.WithField("key", key).Warn("Dropping key")

		l.Order.Remove(last)
		delete(l.Data, remove.Key)
	}

	return nil
}

// Delete the given key, returning an error if the key doesn't exist or the
// user doesn't have permission to remove the key.
func (l *LRU) Delete(key, user string) error {
	l.Lock()
	defer l.Unlock()

	element, ok := l.Data[key]

	if !ok {
		return fmt.Errorf("delete: key %q: %w", key, ErrNotFound)
	}

	entry, _ := element.Value.(*Entry)

	if !Authorised(user, entry.Owner, l.Override) {
		return fmt.Errorf("delete: %q %w of %q", user, ErrNotOwner, key)
	}

	l.Order.Remove(element)
	delete(l.Data, entry.Key)

	return nil
}

// ListAll items in the store.
func (l *LRU) ListAll() []Entry {
	l.Lock()
	defer l.Unlock()

	entries := make([]Entry, l.Order.Len())

	for i := 0; i < l.Order.Len(); i++ {
		element := l.Order.Front()
		l.Order.MoveToBack(element)

		entry, _ := element.Value.(*Entry)

		entries[i] = entry.Clone(l.Depth > 0)
	}

	return entries
}

// List the given item, returning an error if it can't be found.
func (l *LRU) List(key string) (Entry, error) {
	l.Lock()
	defer l.Unlock()

	element, ok := l.Data[key]

	if !ok {
		return Entry{}, fmt.Errorf("list: key %q: %w", key, ErrNotFound)
	}

	entry, _ := element.Value.(*Entry)

	return entry.Clone(l.Depth > 0), nil
}

// Authorised return true if the user is authorised to update an entry created
// by owner.
func Authorised(user, owner string, overwrite bool) bool {
	if overwrite && user == "admin" {
		return true
	}

	return user == owner
}
