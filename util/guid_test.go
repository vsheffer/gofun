package util

import (
	"testing"
)

func TestGuid(t *testing.T) {
	var guid *Guid
	guids := make(map[*Guid]bool)

	for i := 0; i < 1000; i++ {
		guid,_ = NewGuid()
		if guids[guid] {
			t.Error("Guid already used")
		}
		guids[guid] = true
		t.Log(guid)
	}
}