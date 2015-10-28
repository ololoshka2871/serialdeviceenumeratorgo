package serialdeviceenumeratorgo

import (
	"testing"
)

const (
	FTDI_VID = 0x0403
	FT232RL_PID = 0x6001
)

func TestBackendSelector(t *testing.T) {
	en, err := Enumerate()
	if err != nil {
		t.Error(err.Error())
	} else {
		for n, v := range en {
			t.Log(n, " ", v)
		}
	}
}


func TestFTDIPresent(t *testing.T) {
	en, err := Enumerate()
	if err != nil {
		t.Error(err.Error())
	} else {
		for _, v := range en {
			if v.VendorID == FTDI_VID && v.ProductID == FT232RL_PID {
				t.Logf("FTDI chip FT232RL found on %s", v.Name)
			} else {
				t.Logf("Not FTDI chip on %s (%v, %v)", v.Name, v.VendorID, v.ProductID)
			}
		}
	}
}
