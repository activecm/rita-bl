package lists

import (
	"testing"

	"github.com/ocmdev/mgosec"
	blacklist "github.com/ocmdev/rita-bl"
	"github.com/ocmdev/rita-bl/database"
	"github.com/ocmdev/rita-bl/list"
)

func TestDNSBH(t *testing.T) {
	db, err := database.NewMongoDB("localhost:27017", mgosec.None, "rita-blacklist-TEST-DNS-BH")
	if err != nil {
		t.FailNow()
	}
	b := blacklist.NewBlacklist(db, func(err error) { panic(err) })

	//clear the db
	b.SetLists()
	b.Update()
	//get the data
	dnsbh := NewDNSBHList()
	b.SetLists(dnsbh)
	b.Update()
	blHost := "hitnrun.com.my"
	if len(b.CheckEntries(list.BlacklistedHostnameType, blHost)[blHost]) < 1 {
		t.Fail()
	}
}
