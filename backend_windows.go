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
	"strings"
	"unsafe"
)

var guidArray = [...]C.GUID{
	/* Windows Ports Class GUID */
	C.GUID{0x4D36E978, 0xE325, 0x11CE, [...]C.uchar{0xBF, 0xC1, 0x08, 0x00, 0x2B, 0xE1, 0x03, 0x18}},
	/* Virtual Ports Class GUIG (i.e. com0com, nmea and etc) */
	C.GUID{0xDF799E12, 0x3C56, 0x421B, [...]C.uchar{0xB2, 0x98, 0xB6, 0xD3, 0x64, 0x2B, 0xC8, 0x78}},
	/* Windows Modems Class GUID */
	C.GUID{0x4D36E96D, 0xE325, 0x11CE, [...]C.uchar{0xBF, 0xC1, 0x08, 0x00, 0x2B, 0xE1, 0x03, 0x18}},
}

func is_INVALID_HANDLE_VALUE(p interface{}) bool {
	if v, ok := p.(uintptr); ok {
		return v+1 == uintptr(0)
	}
	return false
}

func getNativeName(DeviceInfoSet C.HDEVINFO, DeviceInfoData C.PSP_DEVINFO_DATA) (string, error) {
	key := C.SetupDiOpenDevRegKey(DeviceInfoSet, DeviceInfoData, C.DICS_FLAG_GLOBAL, 0, C.DIREG_DEV, C.KEY_READ)
	if is_INVALID_HANDLE_VALUE(key) {
		return "", errors.New(fmt.Sprintf("Reg error: %d", int(C.GetLastError())))
	}

	var i C.DWORD = 0
	var keyType C.DWORD = 0
	buffKeyName := make([]C.CHAR, 16384)
	buffKeyVal := make([]C.BYTE, 16384)
	for {
		var lenKeyName C.DWORD = C.DWORD(cap(buffKeyName))
		var lenKeyValue C.DWORD = C.DWORD(cap(buffKeyVal))
		ret := C.RegEnumValue(key, i, &buffKeyName[0], &lenKeyName, (*C.DWORD)(nil), &keyType, &buffKeyVal[0], &lenKeyValue)
		if ret == C.ERROR_SUCCESS && keyType == C.REG_SZ {
			itemName := C.GoString((*C.char)(&buffKeyName[0]))
			itemValue := C.GoString((*C.char)(unsafe.Pointer((&buffKeyVal[0]))))

			if strings.Contains(itemName, "PortName") {
				return itemValue, nil
			}
		}
	}

	return "", errors.New("Empty response")
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

		if is_INVALID_HANDLE_VALUE(DeviceInfoSet) { // #define INVALID_HANDLE_VALUE ((HANDLE)(LONG_PTR)-1) => -1
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
