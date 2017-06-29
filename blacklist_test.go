package blacklist

import (
	"os"
	"testing"

	"github.com/ocmdev/mgosec"
	"github.com/ocmdev/rita-bl/database"
	"github.com/ocmdev/rita-bl/list"
	"github.com/ocmdev/rita-bl/sources/mock"
)

//nolint: golint
var __blacklistTestHandle *Blacklist

func TestMain(m *testing.M) {
	db, err := database.NewMongoDB("localhost:27017", mgosec.None, "rita-blacklist-TEST")
	if err != nil {
		os.Exit(-1)
	}
	__blacklistTestHandle = NewBlacklist(db, func(err error) { panic(err) })
	os.Exit(m.Run())
}

func TestDummyList(t *testing.T) {
	t.Run("Update Indeterminate", UpdateDummyList)
	t.Run("Delete", DeleteDummyList)
	t.Run("Empty IP Search", EmptyIPSearch)
	t.Run("Empty Hostname Search", EmptyHostnameSearch)
	t.Run("Update From Scratch", UpdateDummyList)
	t.Run("Update Existing", UpdateDummyList)
	t.Run("IP Search", DummyIPSearch)
	t.Run("Hostname Search", DummyHostnameSearch)
	t.Run("Delete", DeleteDummyList)
}

func UpdateDummyList(t *testing.T) {
	__blacklistTestHandle.SetLists(mock.NewDummyList())
	__blacklistTestHandle.Update()
}

func DeleteDummyList(t *testing.T) {
	__blacklistTestHandle.SetLists()
	__blacklistTestHandle.Update()
}

func EmptyIPSearch(t *testing.T) {
	results := __blacklistTestHandle.CheckEntries(list.BlacklistedIPType, "50.0.0.0")
	if len(results["50.0.0.0"]) != 0 {
		t.Fail()
	}
}

func EmptyHostnameSearch(t *testing.T) {
	results := __blacklistTestHandle.CheckEntries(list.BlacklistedHostnameType, "booking.com")
	if len(results["booking.com"]) != 0 {
		t.Fail()
	}
}

func DummyIPSearch(t *testing.T) {
	results := __blacklistTestHandle.CheckEntries(list.BlacklistedIPType, "50.0.0.0")
	if len(results["50.0.0.0"]) != 1 {
		t.Fail()
	}
}

func DummyHostnameSearch(t *testing.T) {
	results := __blacklistTestHandle.CheckEntries(list.BlacklistedHostnameType, "booking.com")
	if len(results["booking.com"]) != 1 {
		t.Fail()
	}
}

/* NOTE: This test only runs if you supply it an api key and a db path
func TestGoogleRPC(t *testing.T) {
	google, err := rpc.NewGoogleSafeBrowsingURLsRPC("API_KEY", "DB_PATH", os.Stdout)
	if err != nil {
		panic(err)
	}
	url := "http://testsafebrowsing.appspot.com/apiv4/ANY_PLATFORM/MALWARE/URL/"
	__blacklistTestHandle.SetRPCs(google)
	results := __blacklistTestHandle.CheckEntries(list.BlacklistedURLType, url)
	if len(results) < 1 {
		t.Fail()
	}
	if len(results[url]) < 1 {
		t.Fail()
	}
}*/
