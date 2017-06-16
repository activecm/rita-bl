package list

//blacklistedURLType is a BlacklistedType for URLs
type blacklistedURLType struct{}

var urlSingleton *blacklistedURLType

//BlacklistedURLType returns a singleton representing the BlacklistedURLType
func BlacklistedURLType() BlacklistedType {
	if urlSingleton == nil {
		urlSingleton = &blacklistedURLType{}
	}
	return urlSingleton
}

//Type returns "url"
func (b *blacklistedURLType) Type() string { return "url" }

//Validate returns whether or not the given string is a URL
func (b *blacklistedURLType) ValidateIndex(data string) error {
	return nil
}
