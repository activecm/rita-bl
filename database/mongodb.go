package database

import (
	"sync"

	"github.com/ocmdev/rita-blacklist2/list"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//MongoDB provides a MongoDB backend for rita-blacklist
type MongoDB struct {
	session *mgo.Session
}

const database string = "rita-blacklist2"
const listsCollection string = "lists"

//Init opens the connection to the backing database
func (m *MongoDB) Init(connectionString string) error {
	ssn, err := mgo.Dial(connectionString)
	if err != nil {
		return err
	}
	m.session = ssn
	return nil
}

//GetRegisteredLists retrieves all of the lists registered with the database
func (m *MongoDB) GetRegisteredLists() ([]list.Metadata, error) {
	var lists []list.Metadata
	ssn := m.session.Copy()
	defer ssn.Close()

	err := ssn.DB(database).C(listsCollection).Find(nil).All(&lists)
	return lists, err
}

//RegisterList registers a new blacklist source with the database
func (m *MongoDB) RegisterList(l list.Metadata) error {
	ssn := m.session.Copy()
	defer ssn.Close()
	err := ssn.DB(database).C(listsCollection).Insert(l)
	if err != nil {
		return err
	}

	collectionNames, err := ssn.DB(database).CollectionNames()
	if err != nil {
		return err
	}

	for _, entryType := range l.Types {

		found := false
		for _, existingColl := range collectionNames {
			if existingColl == string(entryType) {
				found = true
				break
			}
		}

		if !found {
			ssn.DB(database).C(string(entryType)).Create(&mgo.CollectionInfo{
				DisableIdIndex: true,
			})
			ssn.DB(database).C(string(entryType)).EnsureIndex(mgo.Index{
				Key:    []string{"$hashed:index"},
				Unique: false,
			})
			ssn.DB(database).C(string(entryType)).EnsureIndex(mgo.Index{
				Key:    []string{"index", "list"},
				Unique: true,
			})
		}
	}
	return nil
}

//RemoveList removes an existing blaclist source from the database
func (m *MongoDB) RemoveList(l list.Metadata) error {
	err := m.ClearCache(l)
	if err != nil {
		return err
	}
	ssn := m.session.Copy()
	defer ssn.Close()
	err = ssn.DB(database).C(listsCollection).Remove(bson.M{"name": l.Name})
	if err != nil {
		return err
	}
	return nil
}

//UpdateListMetadata updates the metadata of an existing blacklist
func (m *MongoDB) UpdateListMetadata(l list.Metadata) error {
	ssn := m.session.Copy()
	defer ssn.Close()
	return ssn.DB(database).C(listsCollection).Update(bson.M{"name": l.Name}, l)
}

//ClearCache clears old entries for a given list
func (m *MongoDB) ClearCache(l list.Metadata) error {
	ssn := m.session.Copy()
	defer ssn.Close()
	for _, entryType := range l.Types {
		_, err := ssn.DB(database).C(string(entryType)).RemoveAll(bson.M{"list": l.Name})
		if err != nil {
			return err
		}
	}
	return nil
}

//InsertEntries inserts entries from a list into the database
func (m *MongoDB) InsertEntries(entryType list.BlacklistedEntryType,
	entries <-chan list.BlacklistedEntry, wg *sync.WaitGroup, errorsOut chan<- error) {
	ssn := m.session.Copy()
	defer ssn.Close()
	for entry := range entries {
		dbSafe := DBEntry{
			Index:     entry.Index,
			List:      entry.List.GetMetadata().Name,
			ExtraData: entry.ExtraData,
		}
		err := ssn.DB(database).C(string(entryType)).Insert(dbSafe)
		if err != nil {
			errorsOut <- err
		}
	}
	wg.Done()
}

//FindEntries finds entries of a given type and index
func (m *MongoDB) FindEntries(dataType list.BlacklistedEntryType, index string) ([]DBEntry, error) {
	ssn := m.session.Copy()
	defer ssn.Close()
	var entries []DBEntry
	err := ssn.DB(database).C(string(dataType)).Find(bson.M{"index": index}).All(&entries)
	return entries, err
}
