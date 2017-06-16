package list

import "time"

type (
	//List provides an interface for fetching a list of blacklisted items
	List interface {
		//GetType returns the BlacklistedType associated with this blacklist
		GetType() BlacklistedType
		//GetMetadata returns the Metadata associated with this blacklist
		GetMetadata() *Metadata
		//FetchData fetches the BlacklistedEntry associated with this blacklist.
		//This function must close entriesOut when it is finished.
		//This function should not close errorsOut as it is part of a larger
		//pipeline.
		FetchData(entriesOut chan<- BlacklistedEntry, errorsOut chan<- error)
	}

	//Metadata stores the name of the blacklist source as well as other
	//pieces of metadata
	Metadata struct {
		Name string
		//LastUpdate is the unix timestamp corresponding to the latest fetch of this
		//blacklist
		LastUpdate int64
		//CacheTime is the time in seconds the data from this list should be cached
		CacheTime int64
	}
)

//ShouldFetch returns true if the CacheTiem is up on a given list
func ShouldFetch(m *Metadata) bool {
	return time.Now().Unix() > m.LastUpdate+m.CacheTime
}

//FetchAndValidateEntries fetches the entries from a given List,
//validates the entries coming from the list, and returns a channel
//consisting of the validated entries. errorHandler is used to handle any
//errors that arrise in the processing of the entries
func FetchAndValidateEntries(l List, errorsOut chan<- error) <-chan BlacklistedEntry {
	//fetch the data
	fetchEntriesOutput := make(chan BlacklistedEntry)
	go l.FetchData(fetchEntriesOutput, errorsOut)

	//validate the data
	validatedOuput := make(chan BlacklistedEntry)
	go validateHelper(l.GetType(), fetchEntriesOutput, validatedOuput, errorsOut)
	return validatedOuput
}

func validateHelper(listType BlacklistedType, entriesIn <-chan BlacklistedEntry,
	entriesOut chan<- BlacklistedEntry, errorsOut chan<- error) {

	//loop over the input channel
	for entry := range entriesIn {
		//validate the entry's index
		err := listType.ValidateIndex(entry.Index)
		if err != nil {
			entriesOut <- entry
		} else {
			errorsOut <- err
		}
	}
	close(entriesOut)
}
