package rpc

import "github.com/ocmdev/rita-blacklist2/database"

//RPC is a remote procedure call which can be used to check a given index.
//Note: no caching is done on RPCs inside of rita-blacklist
type RPC func(string) database.DBEntry
