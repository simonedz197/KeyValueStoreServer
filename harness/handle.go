package main

import (
	"errors"
	"fmt"
)

// Handle the results from a test.
func Handle(errs ...error) {
	switch {
	case len(errs) == 1 && errors.Is(errs[0], ErrSkipped):
		fmt.Println("SKIPPED")
	case len(errs) == 1 && errs[0] == nil:
		fmt.Println("PASSED")
	case len(errs) > 0:
		fmt.Println("FAILED")

		failures = append(failures, errs...)
	default:
		fmt.Println("PASSED")
	}
}
