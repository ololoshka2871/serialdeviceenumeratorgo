package serialdeviceenumeratorgo

import (
	"testing"
)

func TestGUIDLen(t *testing.T) {
	if len(guidArray) != 3 {
		t.Error("en(guidArray) != 3")
	}
}
