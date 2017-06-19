package rpc

import (
	"github.com/ocmdev/rita-blacklist2/database"
	"github.com/ocmdev/rita-blacklist2/list"
)

//RPC is a remote procedure call which can be used to check a given index.
//Note: no caching is done on RPCs inside of rita-blacklist
type RPC interface {
	//GetType returns the type of data that this RPC can check
	GetType() list.BlacklistedEntryType
	//Check checks a set of indexes against the rpc and returns a map
	//of the indexes to their results
	Check(...string) (map[string]database.BlacklistResult, error)
}
