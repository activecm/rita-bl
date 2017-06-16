package sources

import (
	"github.com/ocmdev/rita-blacklist2/database"
	"github.com/ocmdev/rita-blacklist2/list"
)

var listFactories map[string]func() list.List

//rpcs is a map of BlacklistedTypes to arrays of remote procedure calls
//which can be used to check a given index. Note: no caching is done on RPCs
//inside of rita-blacklist
var rpcs map[list.BlacklistedType][]func(string) database.DBEntry

func init() {
	listFactories = make(map[string]func() list.List)
	rpcs = make(map[list.BlacklistedType][]func(string) database.DBEntry)
}

//TODO: Write sources

//BootstrapList adds a list factory to the available factories
func BootstrapList(name string, factory func() list.List) {
	listFactories[name] = factory
}

//BootstrapRPC adds a remote procedure call to the list of available RPCs
func BootstrapRPC(entryType list.BlacklistedType, rpc func(string) database.DBEntry) {
	rpcs[entryType] = append(rpcs[entryType], rpc)
}

//GetAvailableLists returns the available blacklist names
func GetAvailableLists() []string {
	lists := make([]string, len(listFactories))
	for list := range listFactories {
		lists = append(lists, list)
	}
	return lists
}

//CreateList uses the list of available factories to create a list.List
func CreateList(name string) list.List {
	factory, ok := listFactories[name]
	if !ok {
		return nil
	}
	return factory()
}

//GetRPCs returns the remote procedure calls for the given entryType
func GetRPCs(entryType list.BlacklistedType) []func(string) database.DBEntry {
	return rpcs[entryType]
}
