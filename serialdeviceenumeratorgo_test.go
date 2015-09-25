package serialdeviceenumeratorgo

import (
	"testing"
)

func TestBackendSelector(t *testing.T) {
	_, err := Enumerate()
	if err != nil {
		t.Error(err.Error())
	}
}
