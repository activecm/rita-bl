package list

import (
	"time"
)

type (
	//List provides an interface for fetching a list of blacklisted items
	List interface {
		//GetMetadata returns the Metadata associated with this blacklist
		GetMetadata() *Metadata
		//FetchData fetches the BlacklistedEntrys associated with this blacklist.
		//This function must close the channels supplied in the entryMap.
		//This function should not close errorsOut as it is part of a larger
		//pipeline.
		FetchData(entryMap BlacklistedEntryMap, errorsOut chan<- error)
	}

	//Metadata stores the name of the blacklist source as well as other
	//pieces of metadata
	Metadata struct {
		//Name is the unique identifying name for the associated List
		Name string
		//Types is a list of the BlacklistedEntryTypes the associated List will produce
		Types []BlacklistedEntryType
		//LastUpdate is the unix timestamp corresponding to the latest fetch of this
		//blacklist
		LastUpdate int64
		//CacheTime is the time in seconds the data from this list should be cached
		CacheTime int64
	}

	//BlacklistedEntryMap is a map of BlacklistedEntryTypes to go channels.
	//This datatype is used for sending different types of BlacklistedEntry together.
	BlacklistedEntryMap map[BlacklistedEntryType]chan BlacklistedEntry
)

//NewBlacklistedEntryMap creates a new BlacklistedEntryMap with a given set
//of BlacklistedEntryTypes
func NewBlacklistedEntryMap(types ...BlacklistedEntryType) BlacklistedEntryMap {
	entryMap := make(BlacklistedEntryMap)
	for _, entryType := range types {
		entryMap[entryType] = make(chan BlacklistedEntry)
	}
	return entryMap
}

//ShouldFetch returns true if the CacheTiem is up on a given list
func ShouldFetch(m *Metadata) bool {
	return time.Now().Unix() > m.LastUpdate+m.CacheTime
}

//FetchAndValidateEntries fetches the entries from a given List,
//validates the entries coming from the list, and returns a channel
//consisting of the validated entries. errorHandler is used to handle any
//errors that arrise in the processing of the entries
func FetchAndValidateEntries(l List, errorsOut chan<- error) BlacklistedEntryMap {
	//fetch the data
	rawOutput := NewBlacklistedEntryMap(l.GetMetadata().Types...)
	go l.FetchData(rawOutput, errorsOut)

	//validate the data
	validatedOutput := NewBlacklistedEntryMap(l.GetMetadata().Types...)
	go validateHelper(rawOutput, validatedOutput, errorsOut)
	return validatedOutput
}

func validateHelper(inputEntryMap BlacklistedEntryMap,
	outputEntryMap BlacklistedEntryMap, errorsOut chan<- error) {
	for inputEntryType, inputEntryChannel := range inputEntryMap {

		go func(
			entryType BlacklistedEntryType,
			inputChannel <-chan BlacklistedEntry,
			outputChannel chan<- BlacklistedEntry,
			errorsChannel chan<- error) {

			for entry := range inputChannel {
				err := entryTypeValidators[entryType](entry.Index)
				if err == nil {
					outputChannel <- entry
				} else {
					errorsChannel <- err
				}
			}
			close(outputChannel)
		}(inputEntryType, inputEntryChannel, outputEntryMap[inputEntryType], errorsOut)

	}

}
