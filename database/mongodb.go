package database

import (
	"crypto/tls"
	"sync"

	"github.com/ocmdev/mgosec"
	"github.com/ocmdev/rita-bl/list"
	mgo "github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

//mongoDB provides a MongoDB backend for rita-blacklist
type mongoDB struct {
	session  *mgo.Session
	database string
}

const listsCollection string = "lists"

//NewMongoDB returns a new mongoDB Handle
func NewMongoDB(conn string, authMech mgosec.AuthMechanism,
	db string) (Handle, error) {
	m := new(mongoDB)
	m.database = db

	ssn, err := mgosec.DialInsecure(conn, authMech)
	if err != nil {
		return nil, err
	}
	m.session = ssn
	return m, nil
}

//NewSecureMongoDB returns a new mongoDB Handle encrypted with TLS
func NewSecureMongoDB(conn string, authMech mgosec.AuthMechanism,
	db string, tlsConf *tls.Config) (Handle, error) {
	m := new(mongoDB)
	m.database = db

	ssn, err := mgosec.Dial(conn, authMech, tlsConf)
	if err != nil {
		return nil, err
	}
	m.session = ssn
	return m, nil
}

//GetRegisteredLists retrieves all of the lists registered with the database
func (m *mongoDB) GetRegisteredLists() ([]list.Metadata, error) {
	var lists []list.Metadata
	ssn := m.session.Copy()
	defer ssn.Close()

	err := ssn.DB(m.database).C(listsCollection).Find(nil).All(&lists)
	return lists, err
}

//RegisterList registers a new blacklist source with the database
func (m *mongoDB) RegisterList(l list.Metadata) error {
	ssn := m.session.Copy()
	defer ssn.Close()

	//get the existing collections
	collectionNames, err := ssn.DB(m.database).CollectionNames()
	if err != nil {
		return err
	}

	//check if lists collection is in collection names
	found := false
	for _, existingColl := range collectionNames {
		if existingColl == listsCollection {
			found = true
			break
		}
	}

	//create listsCollection if it doesn't exist
	if !found {
		err = ssn.DB(m.database).C(listsCollection).Create(&mgo.CollectionInfo{
			DisableIdIndex: true,
		})
		if err != nil {
			return err
		}

		err = ssn.DB(m.database).C(listsCollection).EnsureIndex(mgo.Index{
			Key:    []string{"name"},
			Unique: true,
		})
		if err != nil {
			return err
		}
	}
	//insert the new list
	err = ssn.DB(m.database).C(listsCollection).Insert(l)
	if err != nil {
		return err
	}

	//create the collections for the types of entries this list produces
	for _, entryType := range l.Types {
		//see if the collection exists
		found = false
		for _, existingColl := range collectionNames {
			if existingColl == string(entryType) {
				found = true
				break
			}
		}

		//create the collection if it doesn't exist
		if !found {
			ssn.DB(m.database).C(string(entryType)).Create(&mgo.CollectionInfo{
				DisableIdIndex: true,
			})
			ssn.DB(m.database).C(string(entryType)).EnsureIndex(mgo.Index{
				Key:    []string{"$hashed:index"},
				Unique: false,
			})
			ssn.DB(m.database).C(string(entryType)).EnsureIndex(mgo.Index{
				Key:    []string{"index", "list"},
				Unique: true,
			})
		}
	}
	return nil
}

//RemoveList removes an existing blaclist source from the database
func (m *mongoDB) RemoveList(l list.Metadata) error {
	err := m.ClearCache(l)
	if err != nil {
		return err
	}
	ssn := m.session.Copy()
	defer ssn.Close()
	err = ssn.DB(m.database).C(listsCollection).Remove(bson.M{"name": l.Name})
	if err != nil {
		return err
	}
	return nil
}

//UpdateListMetadata updates the metadata of an existing blacklist
func (m *mongoDB) UpdateListMetadata(l list.Metadata) error {
	ssn := m.session.Copy()
	defer ssn.Close()
	return ssn.DB(m.database).C(listsCollection).Update(bson.M{"name": l.Name}, l)
}

//ClearCache clears old entries for a given list
func (m *mongoDB) ClearCache(l list.Metadata) error {
	ssn := m.session.Copy()
	defer ssn.Close()
	for _, entryType := range l.Types {
		_, err := ssn.DB(m.database).C(string(entryType)).RemoveAll(bson.M{"list": l.Name})
		if err != nil {
			return err
		}
	}
	return nil
}

//InsertEntries inserts entries from a list into the database
func (m *mongoDB) InsertEntries(entryType list.BlacklistedEntryType,
	entries <-chan list.BlacklistedEntry, wg *sync.WaitGroup, errorsOut chan<- error) {
	ssn := m.session.Copy()
	defer ssn.Close()

	i := 0
	bulk := ssn.DB(m.database).C(string(entryType)).Bulk()
	buffSize := 100000
	for entry := range entries {
		bulk.Insert(BlacklistResult{
			Index:     entry.Index,
			List:      entry.List.GetMetadata().Name,
			ExtraData: entry.ExtraData,
		})
		i++
		if i == buffSize {
			_, err := bulk.Run()
			if err != nil {
				errorsOut <- err
			}
			i = 0
			bulk = ssn.DB(m.database).C(string(entryType)).Bulk()
		}
	}
	if i != 0 {
		_, err := bulk.Run()
		if err != nil {
			errorsOut <- err
		}
	}
	wg.Done()
}

//FindEntries finds entries of a given type and index
func (m *mongoDB) FindEntries(dataType list.BlacklistedEntryType, index string) ([]BlacklistResult, error) {
	ssn := m.session.Copy()
	defer ssn.Close()
	var entries []BlacklistResult
	err := ssn.DB(m.database).C(string(dataType)).Find(bson.M{"index": index}).All(&entries)
	return entries, err
}
