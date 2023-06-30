package database

import (
	"sync"

	"github.com/activecm/rita-bl/list"
)

type (
	//Handle provides an interface for using a databae to hold
	//blacklist information
	Handle interface {
		//GetRegisteredLists retrieves all of the lists registered with the database
		GetRegisteredLists() ([]list.Metadata, error)

		//RegisterList registers a new blacklist source with the database
		RegisterList(list.Metadata) error

		//RemoveList removes an existing blacklist source from the database
		RemoveList(list.Metadata) error

		//UpdateListMetadata updates the metadata of an existing blacklist
		UpdateListMetadata(list.Metadata) error

		//ClearCache clears old entries for a given list
		ClearCache(list.Metadata) error

		//InsertEntries inserts entries from a list into the database
		InsertEntries(
			entryType list.BlacklistedEntryType,
			entries <-chan list.BlacklistedEntry,
			wg *sync.WaitGroup, errorsOut chan<- error,
		)

		//FindEntries finds entries of a given type and index
		FindEntries(dataType list.BlacklistedEntryType, index string) ([]BlacklistResult, error)
	}

	//BlacklistResult is the database safe version of BlacklistedEntry.
	//A way to think about this is that entries go in the database, and
	//results come out. This structure is also used to return data from RPC calls.
	BlacklistResult struct {
		//Index is the main data held by this entry
		Index string
		//List is the source list
		List string
		//ExtraData contains extra information this blacklist source provides
		ExtraData map[string]string
	}
)
