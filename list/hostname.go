package list

//blacklistedHostnameType is a BlacklistedType for hostnames
type blacklistedHostnameType struct{}

var hostnameSingleton *blacklistedHostnameType

//BlacklistedHostnameType returns a singleton representing the blacklistedHostnameType
func BlacklistedHostnameType() BlacklistedType {
	if hostnameSingleton == nil {
		hostnameSingleton = &blacklistedHostnameType{}
	}
	return hostnameSingleton
}

//Type returns "hostname"
func (b *blacklistedHostnameType) Type() string { return "hostname" }

//Validate returns whether or not the given string is a hostname
func (b *blacklistedHostnameType) ValidateIndex(data string) error {
	return nil
}
