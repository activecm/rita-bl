package lists

import (
	"bytes"
	"io"
	"testing"

	blacklist "github.com/ocmdev/rita-blacklist2"
	"github.com/ocmdev/rita-blacklist2/database"
	"github.com/ocmdev/rita-blacklist2/list"
)

type nopCloser struct{ io.Reader }

func (nopCloser) Close() error { return nil }

func TestCustomBL(t *testing.T) {
	b := blacklist.NewBlacklist(database.NewMongoDB,
		"localhost:27017", "rita-blacklist-TEST-Custom",
		func(err error) { panic(err) })

	//clear the db
	b.SetLists()
	b.Update()
	getData := func() (io.ReadCloser, error) {
		buf := new(bytes.Buffer)
		buf.WriteString(`
192.168.0.1
192.168.0.2
192.168.0.3
10.10.10.10
`)
		return nopCloser{buf}, nil
	}

	//get the data
	customBL := NewLineSeperatedList(list.BlacklistedIPType, "test", 86400, getData)
	b.SetLists(customBL)
	b.Update()

	blIP := "10.10.10.10"
	if len(b.CheckEntries(list.BlacklistedIPType, blIP)[blIP]) < 1 {
		t.Fail()
	}
}
