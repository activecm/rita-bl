package lists

import (
	"testing"

	blacklist "github.com/ocmdev/rita-bl"
	"github.com/ocmdev/rita-bl/database"
	"github.com/ocmdev/rita-bl/list"
)

func TestMDL(t *testing.T) {
	b := blacklist.NewBlacklist(database.NewMongoDB,
		"localhost:27017", "rita-blacklist-TEST-MDL",
		func(err error) { panic(err) })
	//clear the db
	b.SetLists()
	b.Update()
	//get the data
	mdl := NewMdlList()
	b.SetLists(mdl)
	b.Update()
	blIP := "54.236.134.245"
	if len(b.CheckEntries(list.BlacklistedIPType, blIP)[blIP]) < 1 {
		t.Fail()
	}
}
