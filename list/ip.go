package list

//blacklistedIPType is a BlacklistedType for IP addresses
type blacklistedIPType struct{}

var ipSingleton *blacklistedIPType

//BlacklistedIPType returns a singleton representing the blacklistedIPType
func BlacklistedIPType() BlacklistedType {
	if ipSingleton == nil {
		ipSingleton = &blacklistedIPType{}
	}
	return ipSingleton
}

//Type returns "ip"
func (b *blacklistedIPType) Type() string { return "ip" }

//Validate returns whether or not the given string is an IP address
func (b *blacklistedIPType) ValidateIndex(data string) error {
	return nil
}
