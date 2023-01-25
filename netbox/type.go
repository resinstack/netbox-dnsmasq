package netbox

// Address is an embedded type for json deserialization from the
// netbox API.
type Address struct {
	Address string
}

// Tag is a small associated label that is used to group machines.
// This is used by the shoelaces mapper to determine what target
// should be mapped for the particular machine.
type Tag struct {
	Name string
	Slug string
}

// Device represents the minimum information we wish to retreive from
// the netbox devices API as opposed to using the full fat OpenAPI
// client.
type Device struct {
	ID          int64
	Name        string  `json:"name"`
	PrimaryIPv4 Address `json:"primary_ip4"`
	Tags        []Tag   `json:"tags"`
}

// Interface represents the minimum informatino we wish to retreive
// from the netbox interfaces API as opposed to using the full fat
// OpenAPI client.
type Interface struct {
	ID         int64
	Name       string
	MACAddress string `json:"mac_address"`
}
