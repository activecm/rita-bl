package blacklist

import (
	"sync"
	"time"

	"github.com/ocmdev/rita-bl/database"
	"github.com/ocmdev/rita-bl/list"
	"github.com/ocmdev/rita-bl/sources/rpc"
)

type (
	//Blacklist is the main controller for rita-blacklist
	Blacklist struct {
		db           database.Handle
		lists        []list.List
		rpcs         map[list.BlacklistedEntryType][]rpc.RPC
		errorHandler func(error)
	}
)

//NewBlacklist creates a new blacklist controller and connects to the
//backing database
func NewBlacklist(dbFactory database.Provider, connectionString string, dataset string, errorHandler func(error)) *Blacklist {
	b := &Blacklist{
		db:           dbFactory(),
		lists:        make([]list.List, 0),
		rpcs:         make(map[list.BlacklistedEntryType][]rpc.RPC),
		errorHandler: errorHandler,
	}
	err := b.db.Init(connectionString, dataset)
	if err != nil {
		errorHandler(err)
		return nil
	}
	return b
}

//SetLists loads a set of blacklist sources into the blacklist controller
func (b *Blacklist) SetLists(l ...list.List) {
	b.lists = l
}

//SetRPCs takes in a remote procedure calls for checking the index of the
//given entryType. This is meant for querying web services and outside programs.
//These functions will be ran when CheckEntries is called.
func (b *Blacklist) SetRPCs(rpcs ...rpc.RPC) {
	for _, call := range rpcs {
		b.rpcs[call.GetType()] = append(b.rpcs[call.GetType()], call)
	}
}

//Update updates the blacklist database with the latest information pulled
//from the registered sources
func (b *Blacklist) Update() {
	//handle errors
	finishedProcessingErrors := make(chan struct{})
	errorChannel := createErrorChannel(b.errorHandler, finishedProcessingErrors)
	defer func() { <-finishedProcessingErrors }()
	defer close(errorChannel)

	//get the existing lists from the db
	remoteMetas, err := b.db.GetRegisteredLists()
	if err != nil {
		errorChannel <- err
		return
	}

	//get the lists to remove from the db
	metasToRemove := getListsToRemove(b.lists, remoteMetas)
	for _, metaToRemove := range metasToRemove {
		b.db.RemoveList(metaToRemove)
	}

	existingLists, listsToAdd := findExistingLists(b.lists, remoteMetas)

	updateExistingLists(existingLists, b.db, errorChannel)

	createNewLists(listsToAdd, b.db, errorChannel)
}

//CheckEntries checks entries of different types against the blacklist database
func (b *Blacklist) CheckEntries(entryType list.BlacklistedEntryType, indexes ...string) map[string][]database.BlacklistResult {
	results := make(map[string][]database.BlacklistResult)
	for _, index := range indexes {
		//check against cached blacklists
		entries, err := b.db.FindEntries(entryType, index)
		if err != nil {
			b.errorHandler(err)
			continue
		}
		results[index] = entries
	}
	//run remote procedure calls
	for _, rpc := range b.rpcs[entryType] {
		//get the results from this check on all of the indexes
		rpcResults, err := rpc.Check(indexes...)
		if err != nil {
			b.errorHandler(err)
			continue
		}
		//add the results to the overall results
		for index, entries := range rpcResults {
			results[index] = append(results[index], entries)
		}
	}
	return results
}

func createErrorChannel(errHandler func(error), finished chan<- struct{}) chan<- error {
	errorChannel := make(chan error)
	go func(errHandler func(error), errors <-chan error, finished chan<- struct{}) {
		for err := range errorChannel {
			errHandler(err)
		}
		var fin struct{}
		finished <- fin
	}(errHandler, errorChannel, finished)
	return errorChannel
}

//getListsToRemove finds lists that are in remoteMetas that aren't in loadedLists
func getListsToRemove(loadedLists []list.List, remoteMetas []list.Metadata) []list.Metadata {
	var metasToRemove []list.Metadata
	for _, remoteMeta := range remoteMetas {
		found := false
		for _, loadedList := range loadedLists {
			if remoteMeta.Name == loadedList.GetMetadata().Name {
				found = true
				break
			}
		}
		if !found {
			metasToRemove = append(metasToRemove, remoteMeta)
		}
	}
	return metasToRemove
}

//findExistingLists returns the lists which are in remoteMetas and the lists
//that are not, in that order. Note: this function has the side effect of
//loading in the metadata from the database into the loaded lists
func findExistingLists(loadedLists []list.List, remoteMetas []list.Metadata) ([]list.List, []list.List) {
	var existingLists []list.List
	var listsToAdd []list.List
	for _, loadedList := range loadedLists {
		//load in the remote metadata into the loaded list if found
		var foundMeta *list.Metadata
		for _, remoteMeta := range remoteMetas {
			if loadedList.GetMetadata().Name == remoteMeta.Name {
				foundMeta = &remoteMeta
				break
			}
		}
		if foundMeta != nil {
			loadedList.SetMetadata(*foundMeta)
			existingLists = append(existingLists, loadedList)
		} else {
			listsToAdd = append(listsToAdd, loadedList)
		}
	}
	return existingLists, listsToAdd
}

func updateExistingLists(existingLists []list.List,
	dbHandle database.Handle, errorsOut chan<- error) {
	for _, existingList := range existingLists {
		meta := existingList.GetMetadata()
		if list.ShouldFetch(meta) {
			//kick off fetching in a new thread
			entryMap := list.FetchAndValidateEntries(existingList, errorsOut)
			//delete all existing entries and re-add the list
			err := dbHandle.ClearCache(meta)

			if err != nil {
				errorsOut <- err
				continue
			}

			wg := new(sync.WaitGroup)
			for entryType, entryChannel := range entryMap {
				wg.Add(1)
				go dbHandle.InsertEntries(entryType, entryChannel, wg, errorsOut)
			}
			wg.Wait()

			meta.LastUpdate = time.Now().Unix()
			err = dbHandle.UpdateListMetadata(meta)
			if err != nil {
				errorsOut <- err
				continue
			}
		}
	}
}

func createNewLists(listsToAdd []list.List,
	dbHandle database.Handle, errorsOut chan<- error) {
	for _, listToAdd := range listsToAdd {
		meta := listToAdd.GetMetadata()
		if list.ShouldFetch(listToAdd.GetMetadata()) {
			//kick off fetching in a new thread
			entryMap := list.FetchAndValidateEntries(listToAdd, errorsOut)

			//register the list, create, and index the new collections
			//set the cache to invalid so if the code errors,
			//the code will reimport it
			preWriteMetaCopy := meta
			preWriteMetaCopy.LastUpdate = 0
			preWriteMetaCopy.CacheTime = 0
			err := dbHandle.RegisterList(preWriteMetaCopy)

			if err != nil {
				errorsOut <- err
				continue
			}

			wg := new(sync.WaitGroup)
			for entryType, entryChannel := range entryMap {
				wg.Add(1)
				go dbHandle.InsertEntries(entryType, entryChannel, wg, errorsOut)
			}
			wg.Wait()

			//set the cache to valid
			meta.LastUpdate = time.Now().Unix()
			err = dbHandle.UpdateListMetadata(meta)
			if err != nil {
				errorsOut <- err
				continue
			}
		}
	}
}
