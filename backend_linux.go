package serialdeviceenumeratorgo

/*
#cgo pkg-config: libudev
#include <libudev.h>
*/
import "C"

import (
	"errors"
	"strings"
	"fmt"
	"encoding/hex"
)

var	devNamesMask = [...]string{ "ttyS",       /* standart UART 8250 and etc. */
            "ttyUSB",     /* usb/serial converters PL2303 and etc. */
            "ttyACM",     /* CDC_ACM converters (i.e. Mobile Phones). */
            "ttyMI",      /* MOXA pci/serial converters. */
            "rfcomm" }

var eqBusDrvMap = map[string]string{ "usb" : "ID_USB_DRIVER" }

func enumerate() ([]DeviceDescription, error) {
	var result []DeviceDescription
	
	udev := C.udev_new()
	if udev == nil {
		return nil, errors.New("Udev connection failed!")
	}
	
	enumerate := C.udev_enumerate_new(udev)
	if enumerate == nil {
		return nil, errors.New("Unix: udev_enumerate_new() returned: 0")
	}
	
	var devices 		*C.struct_udev_list_entry
	
	C.udev_enumerate_add_match_subsystem(enumerate, C.CString("tty"))
	C.udev_enumerate_scan_devices(enumerate)
	
	devices = C.udev_enumerate_get_list_entry(enumerate)
	
	for devices != nil {
		syspath := C.udev_list_entry_get_name(devices)
		udev_device := C.udev_device_new_from_syspath(udev, syspath)
		
		if udev_device != nil {
			var dev DeviceDescription
			
			s := C.GoString(C.udev_device_get_devnode(udev_device))
			for _, mask := range devNamesMask {
				if strings.Contains(s, mask) {
					dev.Description = C.GoString(C.udev_device_get_property_value(
							udev_device, C.CString("ID_MODEL_FROM_DATABASE")))
					dev.Revision = C.GoString(C.udev_device_get_property_value(
							udev_device, C.CString("ID_REVISION")))
					dev.Bus = C.GoString(C.udev_device_get_property_value(
							udev_device, C.CString("ID_BUS")))
					dev.Driver = C.GoString(C.udev_device_get_property_value(
							udev_device, C.CString(eqBusDrvMap[dev.Bus])))
					dev.LocationInfo = strings.Replace(C.GoString(C.udev_device_get_property_value(
							udev_device, C.CString("ID_MODEL_ENC"))), "\\x20", " ", -1)
					dev.Manufacturer = C.GoString(C.udev_device_get_property_value(
							udev_device, C.CString("ID_VENDOR_FROM_DATABASE")))
					dev.SubSystem = C.GoString(C.udev_device_get_property_value(
							udev_device, C.CString("SUBSYSTEM")))
					dev.SystemPath = C.GoString(C.udev_device_get_syspath(udev_device))
					dev.ShortName = C.GoString(C.udev_device_get_property_value(
							udev_device, C.CString("DEVNAME")))
					dev.FriendlyName = fmt.Sprintf("%s (%s)", dev.Description, dev.ShortName)
					
					var v []byte
					var err error					
					if v, err = hex.DecodeString(C.GoString(C.udev_device_get_property_value(
								udev_device, C.CString("ID_VENDOR_ID")))); err == nil && len(v) == 2 {
						dev.VendorID = (uint16)(v[0]) << 8 + (uint16)(v[1])
					}	
					if v, err = hex.DecodeString(C.GoString(C.udev_device_get_property_value(
								udev_device, C.CString("ID_MODEL_ID")))); err == nil && len(v) == 2 {
						dev.ProductID = (uint16)(v[0]) << 8 + (uint16)(v[1])
					}
					
					result = append(result, dev)
				}
			}
		}
		devices = C.udev_list_entry_get_next(devices)
	}
		
	C.udev_enumerate_unref(enumerate);
	
	return result, nil
}
