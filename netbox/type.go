package netbox

// Address is an embedded type for json deserialization from the
// netbox API.
type Address struct {
	Address string
}

// Device represents the minimum information we wish to retreive from
// the netbox devices API as opposed to using the full fat OpenAPI
// client.
type Device struct {
	ID          int64
	Name        string  `json:"name"`
	PrimaryIPv4 Address `json:"primary_ip4"`
}

// Interface represents the minimum informatino we wish to retreive
// from the netbox interfaces API as opposed to using the full fat
// OpenAPI client.
type Interface struct {
	ID         int64
	Name       string
	MACAddress string `json:"mac_address"`
}
