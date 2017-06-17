package blacklist

import (
	"os"
	"testing"

	"github.com/ocmdev/rita-blacklist2/list"
	"github.com/ocmdev/rita-blacklist2/sources/mock"
)

//nolint: golint
var __blacklistTestHandle *Blacklist

func TestMain(m *testing.M) {
	__blacklistTestHandle = NewBlacklist("localhost:27017", func(err error) { panic(err) })
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
