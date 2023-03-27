package main

// You can update this if your URL isn't localhost:8000.
const defaultURL = "http://localhost:8000"

// You can update this if you want to default the capabilities you have set.
const defaultCapabilities = ""

// Default depth of the LRU cache (if implemented).
const defaultLRUDepth = 100

// These two are used when stress testing. It'll launch `goroutine` go routines
// and each will fire off `requests` requests. The stored values are also used
// to test /list. If the tests are a bit slow you can change these to 10/100.
const (
	goroutines = 100
	requests   = 1000
)

// The maximum number of failures that will get reported - when the stress test
// blows up there can be a few hundred thousand ;).
const maxFailures = 10

// The usernames the harness will use.
const (
	a = "user_a"
	b = "user_b"
	c = "user_c"

	admin = "admin"
)

// The set of users (excluding the admin user).
var users = []string{a, b, c}

// The logins for all the users. And yes, I'm aware it's using plain text.
// passwords.
var logins = map[string]string{
	a:     "passwordA",
	b:     "passwordB",
	c:     "passwordC",
	admin: "Password1",
}

// The data set used when testing.
var data = map[string]string{"a": "A", "b": "B", "c": "C"}
