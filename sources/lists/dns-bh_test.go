package lists

import (
	"testing"

	blacklist "github.com/ocmdev/rita-blacklist2"
	"github.com/ocmdev/rita-blacklist2/database"
	"github.com/ocmdev/rita-blacklist2/list"
)

func TestDNSBH(t *testing.T) {
	b := blacklist.NewBlacklist(database.NewMongoDB,
		"localhost:27017", "rita-blacklist-TEST-DNS-BH",
		func(err error) { panic(err) })
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
