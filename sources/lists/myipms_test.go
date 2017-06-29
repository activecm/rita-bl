package lists

import (
	"testing"

	"github.com/ocmdev/mgosec"
	blacklist "github.com/ocmdev/rita-bl"
	"github.com/ocmdev/rita-bl/database"
	"github.com/ocmdev/rita-bl/list"
)

func TestMyIPms(t *testing.T) {
	db, err := database.NewMongoDB("localhost:27017", mgosec.None, "rita-blacklist-TEST-MyIPms")
	if err != nil {
		t.FailNow()
	}
	b := blacklist.NewBlacklist(db, func(err error) { panic(err) })

	//clear the db
	b.SetLists()
	b.Update()
	//get the data
	myIPms := NewMyIPmsList()
	b.SetLists(myIPms)
	b.Update()
	blIP := "1.0.146.162"
	if len(b.CheckEntries(list.BlacklistedIPType, blIP)[blIP]) < 1 {
		t.Fail()
	}
}
