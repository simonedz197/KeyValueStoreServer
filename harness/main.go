package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

var failures []error

func main() {
	welcome()

	t, err := startup()

	if err != nil {
		fmt.Println("Failed to start the test harness!")
		fmt.Println(err)
		os.Exit(-1)
	}

	t.SanityCheck()

	fmt.Print("\nRunning...\n\n")

	Handle(t.Ping())

	if len(failures) > 0 {
		fmt.Print("\nYour store is not running!\n\n")
		os.Exit(-1)
	}

	Handle(t.Login()...)
	Handle(t.InvalidLogin())
	Handle(t.InvalidUser())
	Handle(t.CRUD()...)
	Handle(t.Eclipsed()...)
	Handle(t.Override()...)
	Handle(t.LRU()...)
	Handle(t.Stress()...)
	Handle(t.List()...)
	Handle(t.ListAll()...)
	Handle(t.AdvancedList()...)
	Handle(t.Shutdown())

	switch len(failures) {
	case 0:
		fmt.Print("\nAll tests passed, well done!\n\n")
		os.Exit(0)
	case 1:
		fmt.Print("\nOne test failure left to fix:\n\n")
		fmt.Printf("Failure: %v\n", failures[0])
		os.Exit(1)
	}

	if len(failures) > maxFailures {
		failures = failures[0:maxFailures]
		fmt.Printf("\nA few things still need to be fixed (only the first %d are shown):\n\n",
			maxFailures)
	} else {
		fmt.Print("\nA few things still need to be fixed:\n\n")
	}

	if len(failures) > 0 {
		for i, failure := range failures {
			fmt.Printf("%2d) Failure: %v\n", i+1, failure)
		}
	}

	os.Exit(len(failures))
}

func startup() (Tester, error) {
	scanner := bufio.NewScanner(os.Stdin)
	tester := Tester{
		URL:          defaultURL,
		Users:        users,
		Logins:       logins,
		Data:         data,
		Capabilities: defaultCapabilities,
	}

	fmt.Printf("URL for your store (default: %s): ", tester.URL)

	scanner.Scan()
	text := scanner.Text()

	if text != "" {
		tester.URL = text
	}

	if tester.Capabilities == "" {
		fmt.Print("Store capabilities (default: none): ")
	} else {
		fmt.Printf("Store capabilities (default: %s): ", tester.Capabilities)
	}

	scanner.Scan()
	text = scanner.Text()

	if text != "" {
		tester.Capabilities = text
	}

	if tester.Capability("lru") {
		fmt.Printf("LRU depth (default: %d): ", defaultLRUDepth)

		scanner.Scan()
		text = scanner.Text()

		if text == "" {
			text = strconv.Itoa(defaultLRUDepth)
		}

		v, err := strconv.Atoi(text)

		if err != nil {
			return tester, fmt.Errorf("failed to read LRU depth: %w", err)
		}

		tester.Depth = v
	}

	return tester, nil
}

func welcome() {
	fmt.Println()
	fmt.Println("Go Academy (Week 2) Test Harness")
	fmt.Println("================================")
	fmt.Println()
	fmt.Println("Please make sure your key/value store is running and")
	fmt.Println("accessible from this machine.")
	fmt.Println()
	fmt.Println("If you have attempted one or more of the stretch goals")
	fmt.Println("the test harness can test them for you. List the ")
	fmt.Println("capabilities you have added as a space or comma separated")
	fmt.Println("list. Valid options are: login, override, lru and list.")
	fmt.Println("You can override the defaults used in variables.go")
	fmt.Println()
}
