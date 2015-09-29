package serialdeviceenumeratorgo

import (
	"fmt"
	"strings"
	"testing"
)

func TestGUIDLen(t *testing.T) {
	if len(guidArray) != 3 {
		t.Error("en(guidArray) != 3")
	}
}

func TestRegAccess(t *testing.T) {
	s, err := getNativeDriver("FTDIBUS")
	if err != nil {
		t.Error(fmt.Sprintf("regestr querry error: %s", err.Error()))
	}
	if !strings.Contains(s, "bus.sys") {
		t.Error(fmt.Sprintf("invalid result of regestry: %s", s))
	}
}
