package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"bitbucket.org/idomdavis/gohttp/conversation"
	"bitbucket.org/idomdavis/gohttp/session"
)

// Tester handles running the various tests on the key-value store.
type Tester struct {
	Depth        int
	URL          string
	Users        []string
	Logins       map[string]string
	Data         map[string]string
	Capabilities string
}

// List output.
type List struct {
	Key    string
	Owner  string
	Writes int
	Reads  int
	Age    int64
}

// Errors we might care about.
var (
	ErrUnexpectedResult = errors.New("unexpected result")
	ErrSkipped          = errors.New("skipped")
)

var last int

// SanityCheck the data we have.
func (t *Tester) SanityCheck() {
	switch {
	case t.URL == "":
		exit("no target URL defined")
	case len(t.Users) <= 1:
		exit("Harness requires at least 2 users")
	case len(t.Logins)-len(t.Users) != 1:
		exit("Harness needs logins for all users plus the admin user")
	case t.Logins[admin] == "":
		exit("Harness required an admin user")
	case len(t.Data) == 0:
		exit("No test data defined")
	case t.Capability("lru") && t.Depth < len(t.Logins):
		exit(fmt.Sprintf("LRU Depth must be at least %d",
			len(t.Logins)))
	}
}

// Ping test.
func (t *Tester) Ping() error {
	const (
		endpoint = "/ping"
		expected = "pong"
	)

	url := t.URL + endpoint

	Format("Testing /ping")

	r, err := NewReply(conversation.Request{}.Get(url))

	switch {
	case err != nil:
		err = fmt.Errorf("error calling %s: %w", url, err)
	case r.String != expected:
		err = fmt.Errorf("%w: %s: Expected %q, got %q",
			ErrUnexpectedResult, url, expected, r.String)
	}

	return err
}

// Login tests.
func (t *Tester) Login() []error {
	const endpoint = "/login"

	var errs []error

	Format("Getting logins")

	url := t.URL + endpoint

	if !t.Capability("login") {
		for k := range t.Logins {
			t.Logins[k] = k
		}

		return []error{ErrSkipped}
	}

	for user := range t.Logins {
		r, err := NewReply(conversation.Request{
			Username: user,
			Password: t.Logins[user],
		}.Get(url))

		switch {
		case err != nil:
			errs = append(errs, fmt.Errorf("error calling %s: %w", url, err))
		case strings.HasPrefix(r.Auth, "Bearer"):
			t.Logins[user] = r.Auth
		case strings.HasPrefix(r.String, "Bearer"):
			t.Logins[user] = r.String
		default:
			errs = append(errs, fmt.Errorf(
				"%w, couldn't find bearer token in Authorization header or response body",
				ErrUnexpectedResult))
		}
	}

	return errs
}

// InvalidLogin checks invalid password for a valid user are rejected.
func (t *Tester) InvalidLogin() error {
	Format("Checking /login rejects invalid passwords")

	return t.invalidLogin(t.Users[0], "invalid")
}

// InvalidUser checks unknown users are rejected.
func (t *Tester) InvalidUser() error {
	Format("Checking /login rejects invalid users")

	return t.invalidLogin("invalid", t.Logins[t.Users[0]])
}

// CRUD tests, including checking DELETE and GET return a 404 for non-existent
// entries.
func (t *Tester) CRUD() []error {
	var errs []error

	Format("Testing PUT/GET/DELETE /store/<key>")

	for _, user := range t.Users {
		for key, data := range t.Data {
			t.checkNotFound(http.MethodGet, user, key, errs)

			for _, value := range []string{data, data + "2"} {
				errs = t.write(user, key, value, errs)
				errs = t.checkSet(user, key, value, errs)
			}

			errs = t.delete(user, key, errs)
			errs = t.checkNotFound(http.MethodGet, user, key, errs)
			errs = t.checkNotFound(http.MethodDelete, user, key, errs)
		}
	}

	return errs
}

// Eclipsed tests that we can't delete/update other peoples entries.
func (t *Tester) Eclipsed() []error {
	var errs []error

	Format("Checking user can alter another users data")

	for key, value := range t.Data {
		errs = t.write(t.Users[0], key, value, errs)
		errs = t.checkForbidden(http.MethodPut, t.Users[0], t.Users[1], key, value, errs)
		errs = t.checkForbidden(http.MethodDelete, t.Users[0], t.Users[1], key, "", errs)
		errs = t.delete(t.Users[0], key, errs)
	}

	return errs
}

// Stress tests.
func (t *Tester) Stress() []error {
	var (
		wg   sync.WaitGroup
		errs []error
	)

	Format("Stress testing store")

	for i := 0; i < goroutines; i++ {
		wg.Add(1)

		go func(i int) {
			for j := 0; j < requests; j++ {
				key := strconv.Itoa(i) + "_" + strconv.Itoa(j)
				errs = t.write(admin, key, strconv.Itoa(j), errs)
			}

			last = i

			wg.Done()
		}(i)

		wg.Wait()
	}

	return errs
}

// List tests the /list/<key> endpoint.
func (t *Tester) List() []error {
	Format("Testing /list/<key>")

	key := strconv.Itoa(last) + "_" + strconv.Itoa(requests-1)
	list, err := t.listItem(admin, key)

	switch {
	case err != nil:
		return []error{err}
	case list.Owner != admin:
		return []error{fmt.Errorf("%w: owner for list item %s should be %q, not %q",
			ErrUnexpectedResult, key, admin, list.Owner)}
	case list.Key != key:
		return []error{fmt.Errorf("%w: key  for list item %s should be %q, not %q",
			ErrUnexpectedResult, key, key, list.Key)}
	}

	return t.checkNotFound(http.MethodGet, admin, "unset", []error{})
}

// ListAll tests the list/ endpoint.
//nolint:gocyclo,cyclop
func (t *Tester) ListAll() []error {
	var (
		list []List
		errs []error
	)

	Format("Testing /list")

	list, err := t.listAll(admin)

	switch {
	case err != nil:
		return []error{err}
	case len(list) == 0:
		errs = append(errs, fmt.Errorf("%w: /list returned empty list",
			ErrUnexpectedResult))
	case len(list) > t.Depth && t.Depth != 0:
		errs = append(errs, fmt.Errorf(
			"%w: LRU depth is %d but got %d items returned from /list",
			ErrUnexpectedResult, t.Depth, len(list)))
	case len(list) != goroutines*requests && t.Depth == 0:
		errs = append(errs, fmt.Errorf(
			"%w: Expected %d items from /list, got %d",
			ErrUnexpectedResult, goroutines*requests, len(list)))
	case t.Capability("list") && list[len(list)-1].Age == 0:
		errs = append(errs, fmt.Errorf(
			"%w: /list doesn't seem to be returning age",
			ErrUnexpectedResult))
	case t.Capability("list") && list[len(list)-1].Writes != 1:
		errs = append(errs, fmt.Errorf(
			"%w: /list returning incorrect write counts",
			ErrUnexpectedResult))
	}

	for _, e := range list {
		errs = t.optionalDelete(e.Owner, e.Key, errs)
	}

	list, err = t.listAll(admin)

	switch {
	case err != nil:
		errs = append(errs, err)
	case len(list) != 0:
		errs = append(errs, fmt.Errorf(
			"%w: /list returning %d items, expected 0",
			ErrUnexpectedResult, len(list)))
	}

	return errs
}

// AdvancedList tests the format of the extended list requirement.
func (t *Tester) AdvancedList() []error {
	var errs []error

	const reads, writes, maxAge = 3, 2, 1000

	Format("Testing extended list format")

	if !t.Capability("list") {
		return []error{ErrSkipped}
	}

	key, value := "key", "value"

	for i := 0; i < writes; i++ {
		errs = t.write(admin, key, value, errs)
	}

	for i := 0; i < reads; i++ {
		errs = t.checkSet(admin, key, value, errs)
	}

	// Just want to wait a few milliseconds here.
	time.Sleep(time.Millisecond * writes)

	list, err := t.listItem(admin, key)

	switch {
	case err != nil:
		errs = append(errs, err)
	case list.Writes != writes:
		errs = append(errs, fmt.Errorf("%w: expected %d writes, got %d",
			ErrUnexpectedResult, writes, list.Writes))
	case list.Reads != reads:
		errs = append(errs, fmt.Errorf("%w: expected %d reads, got %d",
			ErrUnexpectedResult, reads, list.Reads))
	case list.Age == 0:
		errs = append(errs, fmt.Errorf("%w: item has no age",
			ErrUnexpectedResult))
	case list.Age > maxAge:
		errs = append(errs, fmt.Errorf("%w: age of %dms looks old",
			ErrUnexpectedResult, list.Age))
	}

	return errs
}

// Override checks the admin override function.
func (t *Tester) Override() []error {
	var errs []error

	Format("Testing admin override")

	if !t.Capability("override") {
		return []error{ErrSkipped}
	}

	for key, value := range t.Data {
		errs = t.write(t.Users[0], key, value, errs)
		errs = t.write(admin, key, value, errs)
		errs = t.delete(admin, key, errs)
		errs = t.optionalDelete(t.Users[0], key, errs)
	}

	return errs
}

// LRU checks the LRU cache functionality.
func (t *Tester) LRU() []error {
	const (
		key     = "keep"
		value   = "value"
		discard = "discard"
	)

	var errs []error

	Format("Testing LRU cache")

	if !t.Capability("lru") {
		return []error{ErrSkipped}
	}

	t.write(admin, key, value, errs)
	t.write(admin, discard, value, errs)

	for depth := 0; depth <= t.Depth; depth += len(t.Users) {
		for i, user := range t.Users {
			t.write(user, strconv.Itoa(depth)+user, strconv.Itoa(depth+i), errs)
		}

		errs = t.checkSet(admin, key, value, errs)
	}

	return t.checkNotFound(http.MethodGet, admin, discard, errs)
}

// Shutdown tests we can shut down the server. It'll wait for a second before
// checking.
func (t *Tester) Shutdown() error {
	const endpoint = "/shutdown"

	url := t.URL + endpoint

	Format("Testing /shutdown")

	r, err := NewReply(conversation.Request{
		Headers: map[string]string{session.AuthHeader: t.Logins[t.Users[0]]},
	}.Make(http.MethodGet, url, nil))

	switch {
	case errors.Is(err, ErrForbidden):
	// passed
	case err != nil:
		return fmt.Errorf("%s should return 403 for non admin access, got :%w",
			endpoint, err)
	default:
		return fmt.Errorf("%w: %s should return 403 for non admin access, got :%s",
			ErrUnexpectedResult, endpoint, r.String)
	}

	_, err = NewReply(conversation.Request{
		Headers: map[string]string{session.AuthHeader: t.Logins[admin]},
	}.Make(http.MethodGet, url, nil))

	if err != nil {
		return fmt.Errorf("%s for admin returned: %w", endpoint, err)
	}

	time.Sleep(time.Second)

	_, err = NewReply(conversation.Request{
		Headers: map[string]string{session.AuthHeader: t.Logins[admin]},
	}.Make(http.MethodGet, url, nil))

	if err == nil {
		return fmt.Errorf("%w: server has not shutdown", ErrUnexpectedResult)
	}

	return nil
}

// Capability returns true if the Tester has a capability set.
func (t *Tester) Capability(c string) bool {
	return strings.Contains(t.Capabilities, c)
}

func (t *Tester) write(user, key, value string, errs []error) []error {
	const endpoint = "/store/"

	url := t.URL + endpoint

	_, err := NewReply(conversation.Request{
		Headers: map[string]string{session.AuthHeader: t.Logins[user]},
	}.Make(http.MethodPut, url+key, []byte(value)))

	if err != nil {
		errs = append(errs, fmt.Errorf(
			"failed to PUT %s via %s%s for user %s: %w",
			value, endpoint, key, user, err))
	}

	return errs
}

func (t *Tester) invalidLogin(username, password string) error {
	const endpoint = "/login"

	url := t.URL + endpoint

	if !t.Capability("login") {
		return ErrSkipped
	}

	r, err := NewReply(conversation.Request{
		Username: username,
		Password: password,
	}.Get(url))

	switch {
	case errors.Is(err, ErrUnauthorised):
		return nil
	case err != nil:
		return fmt.Errorf("error calling %s: %w", url, err)
	default:
		return fmt.Errorf(
			"%w, expected Unauthorised, got %s", ErrUnexpectedResult, r.String)
	}
}

func (t *Tester) delete(user, key string, errs []error) []error {
	const endpoint = "/store/"

	url := t.URL + endpoint

	_, err := NewReply(conversation.Request{
		Headers: map[string]string{session.AuthHeader: t.Logins[user]},
	}.Make(http.MethodDelete, url+key, nil))

	if err != nil {
		errs = append(errs, fmt.Errorf(
			"failed to DELETE value using %s%s for user %s: %w",
			endpoint, key, user, err))
	}

	return errs
}

func (t *Tester) optionalDelete(user, key string, errs []error) []error {
	e := t.delete(user, key, []error{})

	if len(e) == 1 && errors.Is(e[0], ErrNotFound) {
		return errs
	}

	return append(errs, e...)
}

func (t *Tester) checkSet(user, key, value string, errs []error) []error {
	const endpoint = "/store/"

	url := t.URL + endpoint

	r, err := NewReply(conversation.Request{
		Headers: map[string]string{session.AuthHeader: t.Logins[user]},
	}.Get(url + key))

	switch {
	case err != nil:
		errs = append(errs, fmt.Errorf(
			"failed to GET value from %s%s for user %s: %w",
			endpoint, key, user, err))
	case r.String != value:
		errs = append(errs, fmt.Errorf(
			"%w: expected %s from %s%s for user %s, got %s",
			ErrUnexpectedResult, value, endpoint, key, user, r.String))
	}

	return errs
}

func (t *Tester) checkNotFound(method, user, key string, errs []error) []error {
	const endpoint = "/store/"

	url := t.URL + endpoint

	r, err := NewReply(conversation.Request{
		Headers: map[string]string{session.AuthHeader: t.Logins[user]},
	}.Make(method, url+key, nil))

	switch {
	case err == nil:
		errs = append(errs, fmt.Errorf(
			"%w: got %s, expected 404 (unset value) from %s %s%s for user %s",
			ErrUnexpectedResult, r.String, method, endpoint, key, user))
	case !errors.Is(err, ErrNotFound):
		errs = append(errs, fmt.Errorf(
			"error trying to %s unset value from %s%s for user %s: %w",
			method, endpoint, key, user, err))
	}

	return errs
}

func (t *Tester) checkForbidden(method, owner, user, key, value string, errs []error) []error {
	const endpoint = "/store/"

	url := t.URL + endpoint

	r, err := NewReply(conversation.Request{
		Headers: map[string]string{session.AuthHeader: t.Logins[user]},
	}.Make(method, url+key, []byte(value)))

	switch {
	case errors.Is(err, ErrForbidden):
	// passed
	case err != nil:
		errs = append(errs, fmt.Errorf(
			"failed to %s %s via %s%s (owned by %s) for user %s: %w",
			method, value, endpoint, key, owner, user, err))
	default:
		errs = append(errs, fmt.Errorf(
			"%w: %s %s via %s%s (owned by %s) for user %s should return 403, got: %s",
			ErrUnexpectedResult, method, value, endpoint, key, owner, user, r.String))
	}

	return errs
}

func (t *Tester) listItem(user, key string) (List, error) {
	const endpoint = "/list/"

	var list List

	url := t.URL + endpoint

	r, err := NewReply(conversation.Request{
		Headers: map[string]string{session.AuthHeader: t.Logins[user]},
	}.Get(url + key))

	if err != nil {
		return list, fmt.Errorf(
			"failed to list item using %s%s for user %s: %w",
			endpoint, key, user, err)
	}

	if err = json.Unmarshal(r.Raw, &list); err != nil {
		return list, fmt.Errorf(
			"failed to parse response from %s%s for user %s: %w\nResponse: %s",
			endpoint, key, user, err, r.String)
	}

	return list, nil
}

func (t *Tester) listAll(user string) ([]List, error) {
	const endpoint = "/list"

	var list []List

	url := t.URL + endpoint

	r, err := NewReply(conversation.Request{
		Headers: map[string]string{session.AuthHeader: t.Logins[user]},
	}.Get(url))

	if err != nil {
		return list, fmt.Errorf(
			"failed to list all items using %s for user %s: %w",
			endpoint, user, err)
	}

	if err = json.Unmarshal(r.Raw, &list); err != nil {
		return list, fmt.Errorf(
			"failed to parse response from %s for user %s: %w\nResponse: %s",
			endpoint, user, err, r.String)
	}

	return list, nil
}

// Format the test title.
func Format(a ...string) {
	fmt.Printf("%-50s ... ", strings.Join(a, " "))
}

func exit(msg string) {
	fmt.Println(msg)
	os.Exit(-1)
}
