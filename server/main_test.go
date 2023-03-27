package main

import (
	"testing"
)

func TestMainCimmandLineFlagParse(t *testing.T) {
	t.Run("NoPortParameterSupplied", func(t *testing.T) {
		_, ok := ParseFlags(make([]string, 0), "")
		if ok {
			t.Error("parseFlag should return false if no port flag specified")
		}
	})
	t.Run("NoPortParameterSupplied", func(t *testing.T) {
		_, ok := ParseFlags([]string{"--help 8000"}, "--port")
		if ok {
			t.Error("parseFlag should return false if no port flag specified")
		}
	})
	t.Run("PortParameterSupplied", func(t *testing.T) {
		_, ok := ParseFlags([]string{"--port 8000"}, "--port")
		if !ok {
			t.Error("parseFlag should return true for this test")
		}
	})
}
