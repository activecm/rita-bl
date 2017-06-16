package blacklist

import (
	"fmt"
	"time"

	"github.com/ocmdev/rita-blacklist2/database"
	"github.com/ocmdev/rita-blacklist2/list"
	"github.com/ocmdev/rita-blacklist2/sources"
)

type (
	//Blacklist is the main controller for rita-blacklist
	Blacklist struct {
		DB           database.Handle
		errorHandler func(error)
	}
)

//NewBlacklist creates a new blacklist controller and connects to the
//backing database
func NewBlacklist(connectionString string, errorHandler func(error)) *Blacklist {
	b := &Blacklist{
		DB:           nil,
		errorHandler: errorHandler,
	}
	err := b.DB.Init(connectionString)
	if err != nil {
		errorHandler(err)
		return nil
	}
	return b
}

//Update updates the blacklist database with the latest information pulled
//from the registered sources
func (b *Blacklist) Update() {
	//handle errors
	errorChannel := createErrorChannel(b.errorHandler)
	defer close(errorChannel)

	//get the existing lists from the db
	registeredMetas, err := b.DB.GetRegisteredLists()
	if err != nil {
		errorChannel <- err
		return
	}

	//obtain functional list objects
	registeredLists := getListsFromMetas(registeredMetas, errorChannel)

	//find lists to add to the database
	listsToAdd := getListsToAdd(registeredLists)

	//update existing blacklists
	updateRegisteredLists(registeredLists, b.DB, errorChannel)

	//insert new blacklists
	createNewLists(listsToAdd, b.DB, errorChannel)
}

//CheckEntries checks entries of different types against the blacklist database
func (b *Blacklist) CheckEntries(entryType list.BlacklistedType, indexes ...string) map[string][]database.DBEntry {
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
		rpcs := sources.GetRPCs(entryType)
		for _, rpc := range rpcs {
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

func getListsFromMetas(metas []list.Metadata, errorsOut chan<- error) []list.List {
	lists := make([]list.List, len(metas))
	for _, meta := range metas {
		l := sources.CreateList(meta.Name)
		if l != nil {
			lists = append(lists, l)
		} else {
			errorsOut <- fmt.Errorf("Could not find a List named %s", meta.Name)
		}
	}
	return lists
}

func getListsToAdd(registeredLists []list.List) []list.List {
	var listsToAdd []list.List
	//check for new list sources
	for _, availableList := range sources.GetAvailableLists() {
		found := false
		for _, registeredList := range registeredLists {
			if availableList == registeredList.GetMetadata().Name {
				found = true
				break
			}
		}
		if !found {
			listsToAdd = append(listsToAdd, sources.CreateList(availableList))
		}
	}
	return listsToAdd
}

func updateRegisteredLists(registeredLists []list.List,
	dbHandle database.Handle, errorsOut chan<- error) {
	for _, registeredList := range registeredLists {
		if list.ShouldFetch(registeredList.GetMetadata()) {
			//kick off fetching in a new thread
			entriesChannel := list.FetchAndValidateEntries(registeredList, errorsOut)
			//delete all existing entries and re-add the list
			err := dbHandle.RemoveList(registeredList)
			if err != nil {
				errorsOut <- err
				continue
			}
			registeredList.GetMetadata().LastUpdate = time.Now().Unix()
			err = dbHandle.RegisterList(registeredList)
			if err != nil {
				errorsOut <- err
				continue
			}

			//in the current thread do the inserts
			dbHandle.InsertEntries(entriesChannel, errorsOut)
		}
	}
}

func createNewLists(listsToAdd []list.List,
	dbHandle database.Handle, errorsOut chan<- error) {
	for _, listToAdd := range listsToAdd {
		if list.ShouldFetch(listToAdd.GetMetadata()) {
			//kick off fetching in a new thread
			entriesChannel := list.FetchAndValidateEntries(listToAdd, errorsOut)
			listToAdd.GetMetadata().LastUpdate = time.Now().Unix()
			err := dbHandle.RegisterList(listToAdd)
			if err != nil {
				errorsOut <- err
				continue
			}

			//in the current thread do the inserts
			dbHandle.InsertEntries(entriesChannel, errorsOut)
		}
	}
}
