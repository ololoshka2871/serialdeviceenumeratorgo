package serialdeviceenumeratorgo

type DeviceDescription struct {
	Name         string
	ShortName    string
	SystemPath   string
	SubSystem    string
	LocationInfo string
	Driver       string
	FriendlyName string
	Description  string
	HardwareID   string
	VendorID     uint16
	ProductID    uint16
	Manufacturer string
	Service      string
	Bus          string
	Revision     string
	IsExists     bool
	IsBusy       bool
}

func Enumerate() ([]DeviceDescription, error) {
	return enumerate()
}
