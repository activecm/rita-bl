package list

type (
	//BlacklistedType provides an interface for different types of blacklists
	//such as ip addresses, hostnames, urls, emails, etc. Additionally it includes
	//a method for validating data
	BlacklistedType interface {
		//Type returns the name of this type of data e.g. "email"
		Type() string
		//Validate checks whether or not a given piece of data fits the schema for
		//this type of data
		ValidateIndex(data string) error
	}

	//BlacklistedEntry is an entry of a List with a given BlacklistedType
	BlacklistedEntry struct {
		//Index is the main data held by this entry
		Index string
		//List is the source list
		List List
		//ExtraData contains extra information this blacklist source provides
		ExtraData map[string]interface{}
	}
)

//NewBlacklistedEntry creates a new BlacklistedEntry
func NewBlacklistedEntry(index string, source List) *BlacklistedEntry {
	return &BlacklistedEntry{
		Index:     index,
		List:      source,
		ExtraData: make(map[string]interface{}),
	}
}
