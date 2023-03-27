package store

import "time"

// Entry in the store.
type Entry struct {
	Owner string `json:"owner"`
	Key   string `json:"key"`
	Value string `json:"-"`

	Reads  int   `json:"reads,omitempty"`
	Writes int   `json:"writes,omitempty"`
	Age    int64 `json:"age,omitempty"`

	Timestamp time.Time `json:"-"`
}

// Clone an entry for use with /list. This will calculate the age and blank any
// items not required for the current mode.
func (e Entry) Clone(detailed bool) Entry {
	entry := Entry{
		Owner: e.Owner,
		Key:   e.Key,
	}

	if detailed {
		entry.Reads = e.Reads
		entry.Writes = e.Writes
		entry.Age = time.Since(e.Timestamp).Milliseconds()
	}

	return entry
}
