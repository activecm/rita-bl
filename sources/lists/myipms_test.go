package lists

import (
	"testing"

	blacklist "github.com/ocmdev/rita-blacklist2"
	"github.com/ocmdev/rita-blacklist2/database"
	"github.com/ocmdev/rita-blacklist2/list"
)

func TestMyIPms(t *testing.T) {
	b := blacklist.NewBlacklist(database.NewMongoDB,
		"localhost:27017", "rita-blacklist-TEST-MyIPms",
		func(err error) { panic(err) })
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
