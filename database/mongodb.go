package database

import (
	"github.com/ocmdev/rita-blacklist2/list"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//MongoDB provides a MongoDB backend for rita-blacklist
type MongoDB struct {
	session *mgo.Session
}

const database string = "rita-blacklist"
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
func (m *MongoDB) RegisterList(l list.List) error {
	ssn := m.session.Copy()
	defer ssn.Close()
	err := ssn.DB(database).C(listsCollection).Insert(l.GetMetadata())
	if err != nil {
		return err
	}
	ssn.DB(database).C(l.GetType().Type()).EnsureIndex(mgo.Index{
		Key:    []string{"Index"},
		Unique: false,
	})
	return err
}

//RemoveList removes an existing blaclist source from the database
func (m *MongoDB) RemoveList(l list.List) error {
	ssn := m.session.Copy()
	defer ssn.Close()
	_, err := ssn.DB(database).C(l.GetType().Type()).RemoveAll(bson.M{"List": l.GetMetadata().Name})
	if err != nil {
		return err
	}
	err = ssn.DB(database).C(listsCollection).Remove(bson.M{"Name": l.GetMetadata().Name})
	if err != nil {
		return err
	}
	return nil
}

//InsertEntries inserts entries from a list into the database
func (m *MongoDB) InsertEntries(entries <-chan list.BlacklistedEntry, errorsOut chan<- error) {
	ssn := m.session.Copy()
	defer ssn.Close()
	for entry := range entries {
		dbSafe := DBEntry{
			Index:     entry.Index,
			List:      entry.List.GetMetadata().Name,
			ExtraData: entry.ExtraData,
		}
		err := ssn.DB(database).C(entry.List.GetType().Type()).Insert(dbSafe)
		if err != nil {
			errorsOut <- err
		}
	}
}

//FindEntries finds entries of a given type and index
func (m *MongoDB) FindEntries(dataType list.BlacklistedType, index string) ([]DBEntry, error) {
	ssn := m.session.Copy()
	defer ssn.Close()
	var entries []DBEntry
	err := ssn.DB(database).C(dataType.Type()).Find(bson.M{"Index": index}).All(&entries)
	return entries, err
}
