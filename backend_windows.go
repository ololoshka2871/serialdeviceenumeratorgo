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
	"log"
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
	defer C.RegCloseKey(key)
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

func getDeviceRegistryProperty(DeviceInfoSet C.HDEVINFO, DeviceInfoData C.PSP_DEVINFO_DATA, property C.DWORD) (interface{}, error) {
	var dataType C.DWORD = 0
	var dataSize C.DWORD = 0

	C.SetupDiGetDeviceRegistryProperty(DeviceInfoSet, DeviceInfoData, property, &dataType, (*C.BYTE)(nil), (C.DWORD)(0), &dataSize)
	data := make([]C.BYTE, dataSize)

	if C.SetupDiGetDeviceRegistryProperty(DeviceInfoSet, DeviceInfoData, property, (*C.DWORD)(nil), &data[0], dataSize, (*C.DWORD)(nil)) == C.TRUE {
		switch dataType {
		case C.REG_EXPAND_SZ, C.REG_SZ:
			if dataSize > 0 {
				return C.GoString((*C.char)(unsafe.Pointer(&data[0]))), nil
			}
		case C.REG_MULTI_SZ:
			if dataSize > 0 {
				var res []string
				i := 0
				for {
					s := C.GoString((*C.char)(unsafe.Pointer(&data[i])))
					if len(s) == 0 {
						break
					} else {
						i += len(s)
						res = append(res, s)
					}
				}
				return res, nil
			}
		case C.REG_DWORD_BIG_ENDIAN, C.REG_DWORD:
			var t C.int
			if uintptr(dataSize) != unsafe.Sizeof(t) {
				panic("Reg incorrect result")
			}
			return int(*(*C.int)(unsafe.Pointer(&data[0]))), nil
		}
	}

	return nil, errors.New("Failed to get data from regestry")
}

func getNativeDriver(service string) (string, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\services\`+service, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()

	if v, _, err := k.GetStringValue("ImagePath"); err == nil {
		return v, nil
	} else {
		return "", err
	}
}

func result2string(v interface{}, e error) (s string, err error) {
	if e != nil {
		return "", e
	} else {
		if ss, ok := v.(string); ok == false {
			return "", errors.New("Not a string value")
		} else {
			return ss, nil
		}
	}
}

func enumerate() ([]DeviceDescription, error) {
	//log.Print("Calling windows backend")

	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\services`, registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}
	defer k.Close()

	var result []DeviceDescription

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
		DeviceInfoData.cbSize = C.DWORD(unsafe.Sizeof(DeviceInfoData))
		for C.SetupDiEnumDeviceInfo(DeviceInfoSet, DeviceIndex, &DeviceInfoData) == C.TRUE {
			name, err := getNativeName(DeviceInfoSet, &DeviceInfoData)
			if err != nil || len(name) == 0 || strings.Contains(name, "LPT") {
				continue
			}

			var dev DeviceDescription
			dev.Name = name
			if dev.Bus, err = result2string(getDeviceRegistryProperty(DeviceInfoSet, &DeviceInfoData, C.SPDRP_ENUMERATOR_NAME)); err != nil {
				log.Printf("Error get bus of device %s: %s", name, err.Error())
			}
			if dev.Description, err = result2string(getDeviceRegistryProperty(DeviceInfoSet, &DeviceInfoData, C.SPDRP_DEVICEDESC)); err != nil {
				log.Printf("Error get Description of device %s: %s", name, err.Error())
			}
			if dev.FriendlyName, err = result2string(getDeviceRegistryProperty(DeviceInfoSet, &DeviceInfoData, C.SPDRP_FRIENDLYNAME)); err != nil {
				log.Printf("Error get FriendlyName of device %s: %s", name, err.Error())
			}
			if dev.HardwareID, err = result2string(getDeviceRegistryProperty(DeviceInfoSet, &DeviceInfoData, C.SPDRP_HARDWAREID)); err != nil {
				log.Printf("Error get HardwareID of device %s: %s", name, err.Error())
			}
			if dev.LocationInfo, err = result2string(getDeviceRegistryProperty(DeviceInfoSet, &DeviceInfoData, C.SPDRP_LOCATION_INFORMATION)); err != nil {
				log.Printf("Error get LocationInfo of device %s: %s", name, err.Error())
			}
			if dev.Manufacturer, err = result2string(getDeviceRegistryProperty(DeviceInfoSet, &DeviceInfoData, C.SPDRP_MFG)); err != nil {
				log.Printf("Error get Manufacturer of device %s: %s", name, err.Error())
			}
			if dev.SubSystem, err = result2string(getDeviceRegistryProperty(DeviceInfoSet, &DeviceInfoData, C.SPDRP_CLASS)); err != nil {
				log.Printf("Error get SubSystem of device %s: %s", name, err.Error())
			}
			if dev.Service, err = result2string(getDeviceRegistryProperty(DeviceInfoSet, &DeviceInfoData, C.SPDRP_SERVICE)); err != nil {
				log.Printf("Error get Service of device %s: %s", name, err.Error())
			}

			result = append(result, dev)
		}
	}

	return nil, nil
}
