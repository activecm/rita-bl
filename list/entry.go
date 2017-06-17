package list

type (
	//BlacklistedEntry is an entry of a List with a given BlacklistedType
	BlacklistedEntry struct {
		//Index is the main data held by this entry
		Index string
		//List is the source list
		List List
		//ExtraData contains extra information this blacklist source provides
		ExtraData map[string]interface{}
	}

	//BlacklistedEntryType is a string representing which type of data
	//is held in the Index field of a BlacklistedEntry
	BlacklistedEntryType string
)

//NewBlacklistedEntry creates a new BlacklistedEntry
func NewBlacklistedEntry(index string, source List) BlacklistedEntry {
	return BlacklistedEntry{
		Index:     index,
		List:      source,
		ExtraData: make(map[string]interface{}),
	}
}

//entryTypeValidators is a map of entry types to functions which validate them
var entryTypeValidators map[BlacklistedEntryType]func(string) error

//BlacklistedHostnameType should be added to the metadata types array
//in order to return hostnames from a list
const BlacklistedHostnameType BlacklistedEntryType = "hostname"

func validateHostname(hostname string) error {
	return nil
}

//BlacklistedIPType should be added to the metadata types array
//in order to return ips from a list
const BlacklistedIPType BlacklistedEntryType = "ip"

func validateIP(ip string) error {
	return nil
}

//BlacklistedURLType should be added to the metadata types array
//in order to return hostnames from a list
const BlacklistedURLType BlacklistedEntryType = "url"

func validateURL(url string) error {
	return nil
}

func init() {
	entryTypeValidators = make(map[BlacklistedEntryType]func(string) error)
	entryTypeValidators[BlacklistedHostnameType] = validateHostname
	entryTypeValidators[BlacklistedIPType] = validateIP
	entryTypeValidators[BlacklistedURLType] = validateURL
}
