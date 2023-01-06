package netbox

type Address struct {
	Address string
}

type Device struct {
	ID          int64
	Name        string  `json:"name"`
	PrimaryIPv4 Address `json:"primary_ip4"`
}

type Interface struct {
	ID         int64
	Name       string
	MACAddress string `json:"mac_address"`
}
