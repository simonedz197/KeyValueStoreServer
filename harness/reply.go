package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"bitbucket.org/idomdavis/gohttp/session"
)

// Reply is a convenience struct to handle responses from the server.
type Reply struct {
	Raw    []byte
	String string
	Auth   string
	Code   int
}

// Turn non 200 codes into errors.
var (
	ErrUnauthorised = fmt.Errorf("%d: %s", http.StatusUnauthorized,
		http.StatusText(http.StatusUnauthorized))
	ErrForbidden = fmt.Errorf("%d: %s", http.StatusForbidden,
		http.StatusText(http.StatusForbidden))
	ErrNotFound = fmt.Errorf("%d: %s", http.StatusNotFound,
		http.StatusText(http.StatusNotFound))
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
)

// NewReply will build a reply and handle errors from a call to the server.
func NewReply(r *http.Response, err error) (Reply, error) {
	if err != nil {
		return Reply{}, fmt.Errorf("failed to make request: %w", err)
	}

	b, err := ioutil.ReadAll(r.Body)

	if err != nil {
		return Reply{}, fmt.Errorf("failed to read response: %w", err)
	}

	_ = r.Body.Close()

	switch r.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized:
		err = ErrUnauthorised
	case http.StatusForbidden:
		err = ErrForbidden
	case http.StatusNotFound:
		err = ErrNotFound
	default:
		err = fmt.Errorf("%w: %d: %s", ErrUnexpectedStatusCode, r.StatusCode,
			http.StatusText(r.StatusCode))
	}

	return Reply{
		Raw:    b,
		String: string(b),
		Auth:   r.Header.Get(session.AuthHeader),
		Code:   r.StatusCode,
	}, err
}
