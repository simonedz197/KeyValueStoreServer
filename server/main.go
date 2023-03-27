// main package
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ParseFlags takes a list of strings and tries to parse looking for flag passed in. returns value and true if found. blank and false if not
func ParseFlags(args []string, flag string) (string, bool) {

	if len(args) > 0 {
		// arg 0 is reserved so we must have more than 0 if we have a port argument
		for _, v := range args {
			if strings.HasPrefix(v, flag) {
				if elements := strings.Split(v, " "); len(elements) > 1 {
					return elements[1], true
				}
			}
		}
	}

	return "", false
}

func main() {
	value, ok := ParseFlags(os.Args, "--port")
	if !ok {
		os.Exit(-1)
	}

	serverPorttoListenOn, err := strconv.Atoi(value)

	if err != nil {
		os.Exit(-1)
	}

	fmt.Println("Will Listen on Port ", serverPorttoListenOn)
}
