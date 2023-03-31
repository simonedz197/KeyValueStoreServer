package handlers_test

import (
	handler "KeyValueStoreServer/server/handlers"
	"testing"
)

func TestGetKeyValue(t *testing.T) {
	t.Run("No key value", func(t *testing.T) {
		val := handler.GetKeyValue("/store/", "/store/")
		if val != "" {
			t.Error("should return empty string is no key on url")
		}
	})
	t.Run("Key value found", func(t *testing.T) {
		val := handler.GetKeyValue("/store/", "/store/keyvalue")
		if val != "keyvalue" {
			t.Error("should return keyvalue")
		}
	})
	t.Run("URL path after key value found", func(t *testing.T) {
		val := handler.GetKeyValue("/store/", "/store/keyvalue/url/continues")
		if val != "" {
			t.Error("should return empty string if url continues after key value")
		}
	})
}
