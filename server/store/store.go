// Package store provides a key value store.
package store

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// ErrNotFounf is Key Not Found
var ErrNotFound = errors.New("key not found")

// ErrForbidden is no access to key
var ErrForbidden = errors.New("Forbidden")

const admin = "admin"

// StoreDepth max number of values to retain in cache.
var StoreDepth = 100

var userList = map[string]string{
	"user_a": "$2a$10$tHRy2eY8a5oOjbHPnZ52X.ME5rX5L9DgoGT8cK7s8jImrv2GMqZXy",
	"user_b": "$2a$10$A0lv9mPC50j5u/r/KbtTAOkXRP8BbpioXz9ef1xGg2dtaOk5Kmo9u",
	"user_c": "$2a$10$yjAy3NgsP2KJ3yk.cAwOjOtK7P4V2jHlNfkoA3bCvp6pvdYBl52Vu",
	admin:    "$2a$10$qkKN5QitJNEZKExOFFk7BeLFh.DV4asusJ51niFLWGeD7g/W6XJWC",
}

// DataValue struct stored in store.
type DataValue struct {
	Owner     string `json:"owner"`
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
	Writes    int
	Reads     int
}

// ListValue struct for returning key info.
type ListValue struct {
	Key    string `json:"key"`
	Owner  string `json:"owner"`
	Writes int    `json:"writes"`
	Reads  int    `json:"reads"`
	Age    int64  `json:"age"`
}

// ListRequest struct for returning channel of list objects.
type ListRequest struct {
	Key      string
	Owner    string
	Response chan []ListValue
}

// FetchRequest response of a fetch.
type FetchRequest struct {
	Key      string
	Response chan interface{}
}

// UpsertRequest message to send to get update.
type UpsertRequest struct {
	Key      string
	Value    string
	Owner    string
	Response chan interface{}
}

// DeleteRequest to signal delete.
type DeleteRequest struct {
	Key      string
	Owner    string
	Response chan interface{}
}

// DoneRequest to signal done.
type DoneRequest struct {
}

// Store struct to hold the map and access channels.
type Store struct {
	value         map[string]DataValue
	upsertChannel chan UpsertRequest
	deleteChannel chan DeleteRequest
	fetchChannel  chan FetchRequest
	listChannel   chan ListRequest
}

// Upsert amend an entry in the store.
func (s *Store) Upsert(key, owner, value string) chan interface{} {
	responseChannel := make(chan interface{})
	s.upsertChannel <- UpsertRequest{Key: key, Owner: owner, Value: value, Response: responseChannel}

	return responseChannel
}

// Delete remove an entry from the store.
func (s *Store) Delete(key, owner string) chan interface{} {
	responseChannel := make(chan interface{})
	s.deleteChannel <- DeleteRequest{Key: key, Owner: owner, Response: responseChannel}

	return responseChannel
}

// Fetch gets an entry from the store using the give key.
func (s *Store) Fetch(key string) chan interface{} {
	responseChannel := make(chan interface{})
	s.fetchChannel <- FetchRequest{Key: key, Response: responseChannel}

	return responseChannel
}

// List gets key/owner for all keys.
func (s *Store) List(owner string) chan []ListValue {
	responseChannel := make(chan []ListValue)
	s.listChannel <- ListRequest{Owner: owner, Key: "", Response: responseChannel}

	return responseChannel
}

// ListForKey gets key/owner for specific key.
func (s *Store) ListForKey(key, owner string) chan []ListValue {
	responseChannel := make(chan []ListValue)
	s.listChannel <- ListRequest{Owner: owner, Key: key, Response: responseChannel}

	return responseChannel
}

func lru(m *map[string]DataValue) string {
	var oldestKey string

	now := time.Now().UnixNano()

	for k, v := range *m {
		if v.Timestamp < now {
			oldestKey = k
			now = v.Timestamp
		}
	}

	return oldestKey
}

// Done channel.
var Done = make(chan DoneRequest)

// Monitor checks for store transaction messages and acts accordingly.
func (s *Store) Monitor() {
	go transactionMonitor()

	loop := true

	for loop {
		select {
		case value := <-s.upsertChannel:
			transactionChannel <- value
		case listreq := <-s.listChannel:
			transactionChannel <- listreq
		case freq := <-s.fetchChannel:
			transactionChannel <- freq
		case dreq := <-s.deleteChannel:
			transactionChannel <- dreq
		case req := <-Done:
			transactionChannel <- req

			loop = false
		}
	}
}

var transactionChannel = make(chan interface{})

func transactionMonitor() {
	for {
		transaction := <-transactionChannel

		if _, ok := transaction.(DoneRequest); ok {
			break
		}
		// upsert transaction
		if msg, ok := transaction.(UpsertRequest); ok {
			transactionUpsert(msg)
			continue
		}
		// delete transaction
		if msg, ok := transaction.(DeleteRequest); ok {
			transactionDelete(msg)
			continue
		}
		// fetcj transaction
		if msg, ok := transaction.(FetchRequest); ok {
			transactionFetch(msg)
			continue
		}
		// list transactiom
		if msg, ok := transaction.(ListRequest); ok {
			transactionList(msg)
			continue
		}
	}
}

func transactionDelete(msg DeleteRequest) {
	if entry, ok := internalStore[msg.Key]; ok {
		if entry.Owner == msg.Owner || msg.Owner == admin {
			delete(internalStore, msg.Key)
			msg.Response <- nil
		} else {
			msg.Response <- ErrForbidden
		}
	} else {
		msg.Response <- ErrNotFound
	}
}

func transactionUpsert(msg UpsertRequest) {
	storeFull := len(internalStore) >= StoreDepth

	if current, ok := internalStore[msg.Key]; ok {
		// trying to update
		if msg.Owner != current.Owner && msg.Owner != admin {
			msg.Response <- ErrForbidden
		} else {
			internalStore[msg.Key] = DataValue{
				Owner:     current.Owner,
				Value:     msg.Value,
				Timestamp: time.Now().UnixNano(),
				Writes:    current.Writes + 1,
				Reads:     current.Reads,
			}
			msg.Response <- nil
		}
	} else {
		// inserting
		if storeFull {
			oldestKey := lru(&internalStore)
			delete(internalStore, oldestKey)
		}

		internalStore[msg.Key] = DataValue{
			Owner:     msg.Owner,
			Value:     msg.Value,
			Timestamp: time.Now().UnixNano(),
			Writes:    1,
			Reads:     0,
		}
		msg.Response <- nil
	}
}

func transactionFetch(msg FetchRequest) {
	val, ok := internalStore[msg.Key]
	if ok {
		msg.Response <- val
		val.Reads++
		val.Timestamp = time.Now().UnixNano()
		internalStore[msg.Key] = val
	} else {
		msg.Response <- nil
	}
}

func transactionList(msg ListRequest) {
	var responseList []ListValue

	if msg.Key == "" {
		// look at all keys and add them if they belong to owner or if owner = admin
		for key, element := range internalStore {
			if element.Owner == msg.Owner || msg.Owner == admin {
				responseList = append(responseList, ListValue{
					Key:    key,
					Owner:  element.Owner,
					Writes: element.Writes,
					Reads:  element.Reads,
					Age:    age(element.Timestamp)})
			}
		}
		msg.Response <- responseList

		return
	}

	val, ok := internalStore[msg.Key]
	if ok {
		// found the specific key, add it if belongs to owner or owner is admin
		if val.Owner == msg.Owner || msg.Owner == admin {
			responseList = append(responseList, ListValue{
				Key:    msg.Key,
				Owner:  val.Owner,
				Writes: val.Writes,
				Reads:  val.Reads,
				Age:    age(val.Timestamp)})
		}
	}

	msg.Response <- responseList
}

var internalStore = make(map[string]DataValue)
var upsertChannel = make(chan UpsertRequest)
var deleteChannel = make(chan DeleteRequest)
var fetchChannel = make(chan FetchRequest)
var listChannel = make(chan ListRequest)

// PublicAccess creates a new store to be used.
var PublicAccess = Store{
	value:         internalStore,
	upsertChannel: upsertChannel,
	deleteChannel: deleteChannel,
	fetchChannel:  fetchChannel,
	listChannel:   listChannel,
}

func age(since int64) int64 {
	return time.Now().UnixNano()/int64(time.Millisecond) - since
}

// ValidateLogin check usename and password is valid in list of users.
func ValidateLogin(username, password string) bool {
	value, ok := userList[username]
	if !ok {
		return false
	}

	// create byte from password string
	hash := []byte(value)
	pwd := []byte(password)
	compare := bcrypt.CompareHashAndPassword(hash, pwd)

	if compare != nil {
		return false
	}

	return true
}

// func HashFromPassword(password []byte) string {
// 	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return string(hash)
// }
