package serialdeviceenumeratorgo

/*
#cgo LDFLAGS: -lsetupapi
#include <windows.h>
#include <setupapi.h>
*/
import "C"

import (
	//"log"
	"errors"
	"fmt"
	"golang.org/x/sys/windows/registry"
)

var guidArray = [...]C.GUID{
	/* Windows Ports Class GUID */
	C.GUID{0x4D36E978, 0xE325, 0x11CE, [...]C.uchar{0xBF, 0xC1, 0x08, 0x00, 0x2B, 0xE1, 0x03, 0x18}},
	/* Virtual Ports Class GUIG (i.e. com0com, nmea and etc) */
	C.GUID{0xDF799E12, 0x3C56, 0x421B, [...]C.uchar{0xB2, 0x98, 0xB6, 0xD3, 0x64, 0x2B, 0xC8, 0x78}},
	/* Windows Modems Class GUID */
	C.GUID{0x4D36E96D, 0xE325, 0x11CE, [...]C.uchar{0xBF, 0xC1, 0x08, 0x00, 0x2B, 0xE1, 0x03, 0x18}},
}

func getNativeName(DeviceInfoSet C.HDEVINFO, DeviceInfoData C.PSP_DEVINFO_DATA) string {
	return ""
}

func enumerate() ([]DeviceDescription, error) {
	//log.Print("Calling windows backend")

	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\services`, registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}
	defer k.Close()

	for i := 0; i < len(guidArray); i++ {
		DeviceInfoSet := C.SetupDiGetClassDevs(&guidArray[i], (*C.CHAR)(nil), (*C.struct_HWND__)(nil), C.DIGCF_PRESENT)

		if uintptr(DeviceInfoSet)+1 == uintptr(0) { // #define INVALID_HANDLE_VALUE ((HANDLE)(LONG_PTR)-1) => -1
			return nil, errors.New(fmt.Sprintf(
				`Windows: SerialDeviceEnumeratorPrivate::updateInfo() 
				SetupDiGetClassDevs() returned INVALID_HANDLE_VALUE, 
				last error: %d`, int(C.GetLastError())))
		}

		var DeviceIndex C.DWORD = 0
		var DeviceInfoData C.SP_DEVINFO_DATA
		for C.SetupDiEnumDeviceInfo(DeviceInfoSet, DeviceIndex, &DeviceInfoData) == C.TRUE {

		}
	}

	return nil, nil
}
