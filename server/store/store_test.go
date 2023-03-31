// Package store_test.
package store_test

import (
	"KeyValueStoreServer/server/store"
	"testing"
)

func TestValidatLogin(t *testing.T) {
	t.Run("Valid Login", func(t *testing.T) {
		ok := store.ValidateLogin("admin", "Password1")
		if !ok {
			t.Error("Expected a true result for admin Password1 but got false")
		}
	})
	t.Run("Invalid Password", func(t *testing.T) {
		ok := store.ValidateLogin("admin", "password1")
		if ok {
			t.Error("Expected a false result for admin password1 but got true")
		}
	})

	t.Run("Invalid User", func(t *testing.T) {
		ok := store.ValidateLogin("simon", "password1")
		if ok {
			t.Error("Expected a false result for simon password1 but got true")
		}
	})
}

func TestInsert(t *testing.T) {
	go store.PublicAccess.Monitor()
	t.Run("Insert", func(t *testing.T) {
		go store.PublicAccess.Monitor()
		response := <-store.PublicAccess.Upsert("key0", "admin", "should be pushed off")
		if response != nil {
			t.Errorf("expected no error but got %v", response)
		}
		response = <-store.PublicAccess.Upsert("key1", "user1", "value1")
		if response != nil {
			t.Errorf("expected no error but got %v", response)
		}
		response = <-store.PublicAccess.Upsert("key2", "user2", "value2")
		if response != nil {
			t.Errorf("expected no error but got %v", response)
		}
		response = <-store.PublicAccess.Upsert("key3", "user1", "value3")
		if response != nil {
			t.Errorf("expected no error but got %v", response)
		}
		response = <-store.PublicAccess.Upsert("key4", "user2", "value4")
		if response != nil {
			t.Errorf("expected no error but got %v", response)
		}
		response = <-store.PublicAccess.Upsert("key5", "user1", "value5")
		if response != nil {
			t.Errorf("expected no error but got %v", response)
		}
		response = <-store.PublicAccess.Upsert("key6", "user3", "value6")
		if response != nil {
			t.Errorf("expected no error but got %v", response)
		}
	})
}

func TestUpdateAndFetchReadAndWrites(t *testing.T) {
	t.Run("Update", func(t *testing.T) {
		response := <-store.PublicAccess.Upsert("key1", "user1", "value1-amended By User")
		if response != nil {
			t.Errorf("Expected nil on insert but got %v", response)
		}

		response = <-store.PublicAccess.Fetch("key1")
		val, ok := response.(store.DataValue)
		if !ok {
			t.Errorf("Expected a DataValue structure back from fetch but got %v", response)
		} else if val.Value != "value1-amended By User" {
			t.Errorf("Expected value to be value1-amended By User but got %s", val.Value)
		}

		response = <-store.PublicAccess.Upsert("key1", "admin", "value1-amended By Admin")
		if response != nil {
			t.Errorf("Expected nil on insert but got %v", response)
		}

		response = <-store.PublicAccess.Fetch("key1")
		val, ok = response.(store.DataValue)
		if !ok {
			t.Errorf("Expected a DataValue structure back from fetch but got %v", response)
		} else if val.Value != "value1-amended By Admin" {
			t.Errorf("Expected value to be value1-amended By Admin but got %s", val.Value)
		} else {
			if val.Writes != 3 {
				t.Errorf("Expected writes to be 3 but got %d", val.Writes)
			} else {
				if val.Reads != 2 {
					t.Errorf("Expected reads to be 2 but got %d", val.Reads)
				}
			}
		}
	})
}

func TestCheckDelete(t *testing.T) {
	t.Run("Delete", func(t *testing.T) {

		// now attempt to delete a key we do not own - not admin
		response := <-store.PublicAccess.Delete("key6", "user1")
		if response == nil {
			t.Errorf("Expected response to be forbidden but got %v", response)
		}

		// attempt to delete a key we do not own - admin
		response = <-store.PublicAccess.Delete("key6", "admin")
		if response != nil {
			t.Errorf("Expected response to be nil but got %v", response)
		}

		// attempt to delete a key we do own
		response = <-store.PublicAccess.Delete("key5", "user1")
		if response != nil {
			t.Errorf("Expected response to be nil but got %v", response)
		}
	})
}

func TestLRUCapability(t *testing.T) {
	var maxentries = 5
	store.StoreDepth = maxentries
	t.Run("lru", func(t *testing.T) {
		response := <-store.PublicAccess.Upsert("key7", "admin", "value7")
		if response != nil {
			t.Errorf("expected no error but got %v", response)
		}

		// now try and get key0 - should be gone
		response = <-store.PublicAccess.Fetch("key0")
		if response != nil {
			t.Errorf("Expected response to be nil but got %v", response)
		}
	})
}

func TestList(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		response := <-store.PublicAccess.List("user1")
		if len(response) != 2 {
			t.Errorf("Expected 2 entries for user1 but got %d", len(response))
		}
	})

	t.Run("Admin", func(t *testing.T) {
		response := <-store.PublicAccess.List("admin")
		if len(response) != 5 {
			t.Errorf("Expected 5 entries for admin but got %d", len(response))
		}
	})
}

func TestListKey(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		response := <-store.PublicAccess.ListForKey("key3", "user1")
		if len(response) != 1 {
			t.Errorf("Expected 1 entries for user1 key3 but got %d", len(response))
		}
	})

	t.Run("BasicMissing", func(t *testing.T) {
		response := <-store.PublicAccess.ListForKey("key4", "user1")
		if len(response) != 0 {
			t.Errorf("Expected 0 entries for user1 key4 but got %d", len(response))
		}
	})

	t.Run("Admin", func(t *testing.T) {
		response := <-store.PublicAccess.ListForKey("key3", "admin")
		if len(response) != 1 {
			t.Errorf("Expected 1 entries for admin but got %d", len(response))
		}
	})

	store.Done <- store.DoneRequest{}

}
