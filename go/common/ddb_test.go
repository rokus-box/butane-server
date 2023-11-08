package common

import "testing"

func TestNewDDB(t *testing.T) {
	client := NewDDB()

	if nil == client {
		t.Error("client is nil")
	}
}
