package blacklist

import (
	"sync"
	"time"

	"github.com/ocmdev/rita-blacklist2/database"
	"github.com/ocmdev/rita-blacklist2/list"
	"github.com/ocmdev/rita-blacklist2/sources/rpc"
)

type (
	//Blacklist is the main controller for rita-blacklist
	Blacklist struct {
		DB           database.Handle
		lists        []list.List
		rpcs         map[list.BlacklistedEntryType][]rpc.RPC
		errorHandler func(error)
	}
)

//NewBlacklist creates a new blacklist controller and connects to the
//backing database
func NewBlacklist(connectionString string, errorHandler func(error)) *Blacklist {
	b := &Blacklist{
		DB:           &database.MongoDB{},
		lists:        make([]list.List, 0),
		rpcs:         make(map[list.BlacklistedEntryType][]rpc.RPC),
		errorHandler: errorHandler,
	}
	err := b.DB.Init(connectionString)
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
func (b *Blacklist) SetRPCs(entryType list.BlacklistedEntryType, checkFuncs ...rpc.RPC) {
	b.rpcs[entryType] = checkFuncs
}

//Update updates the blacklist database with the latest information pulled
//from the registered sources
func (b *Blacklist) Update() {
	//handle errors
	errorChannel := createErrorChannel(b.errorHandler)
	defer close(errorChannel)

	//get the existing lists from the db
	remoteMetas, err := b.DB.GetRegisteredLists()
	if err != nil {
		errorChannel <- err
		return
	}

	//get the lists to remove from the db
	metasToRemove := getListsToRemove(b.lists, remoteMetas)
	for _, metaToRemove := range metasToRemove {
		b.DB.RemoveList(metaToRemove)
	}

	existingLists, listsToAdd := findExistingLists(b.lists, remoteMetas)

	updateExistingLists(existingLists, b.DB, errorChannel)

	createNewLists(listsToAdd, b.DB, errorChannel)
}

//CheckEntries checks entries of different types against the blacklist database
func (b *Blacklist) CheckEntries(entryType list.BlacklistedEntryType, indexes ...string) map[string][]database.DBEntry {
	results := make(map[string][]database.DBEntry)
	for _, index := range indexes {
		//check against cached blacklists
		entries, err := b.DB.FindEntries(entryType, index)
		if err != nil {
			b.errorHandler(err)
			continue
		}
		results[index] = entries

		//run remote procedure calls
		for _, rpc := range b.rpcs[entryType] {
			results[index] = append(results[index], rpc(index))
		}
	}
	return results
}

func createErrorChannel(errHandler func(error)) chan<- error {
	errorChannel := make(chan error)
	go func(errHandler func(error), errors <-chan error) {
		for err := range errorChannel {
			errHandler(err)
		}
	}(errHandler, errorChannel)
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
//that are not, in that order
func findExistingLists(loadedLists []list.List, remoteMetas []list.Metadata) ([]list.List, []list.List) {
	var existingLists []list.List
	var listsToAdd []list.List
	for _, loadedList := range loadedLists {
		found := false
		for _, remoteMeta := range remoteMetas {
			if loadedList.GetMetadata().Name == remoteMeta.Name {
				found = true
				break
			}
		}
		if found {
			existingLists = append(existingLists, loadedList)
		} else {
			listsToAdd = append(listsToAdd, loadedList)
		}
	}
	return existingLists, loadedLists
}

func updateExistingLists(existingLists []list.List,
	dbHandle database.Handle, errorsOut chan<- error) {
	for _, existingList := range existingLists {
		if list.ShouldFetch(existingList.GetMetadata()) {
			//kick off fetching in a new thread
			entryMap := list.FetchAndValidateEntries(existingList, errorsOut)
			//delete all existing entries and re-add the list
			err := dbHandle.ClearCache(*existingList.GetMetadata())
			if err != nil {
				errorsOut <- err
				continue
			}
			//in the current thread do the inserts
			wg := new(sync.WaitGroup)
			for entryType, entryChannel := range entryMap {
				wg.Add(1)
				go dbHandle.InsertEntries(entryType, entryChannel, wg, errorsOut)
			}
			wg.Wait()
			existingList.GetMetadata().LastUpdate = time.Now().Unix()
			err = dbHandle.UpdateListMetadata(*existingList.GetMetadata())
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
		if list.ShouldFetch(listToAdd.GetMetadata()) {
			//kick off fetching in a new thread
			entryMap := list.FetchAndValidateEntries(listToAdd, errorsOut)
			listToAdd.GetMetadata().LastUpdate = time.Now().Unix()
			err := dbHandle.RegisterList(*listToAdd.GetMetadata())
			if err != nil {
				errorsOut <- err
				continue
			}
			wg := new(sync.WaitGroup)
			//in the current thread do the inserts
			for entryType, entryChannel := range entryMap {
				wg.Add(1)
				go dbHandle.InsertEntries(entryType, entryChannel, wg, errorsOut)
			}
			wg.Wait()
		}
	}
}
