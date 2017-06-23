package lists

import (
	"io"

	"github.com/ocmdev/rita-blacklist2/list"
	"github.com/ocmdev/rita-blacklist2/sources/lists/util"
)

//NewDNSBHList returns a new DNSBHList object
func NewDNSBHList() list.List {
	return NewLineSeperatedList(
		list.BlacklistedHostnameType,
		"dns-bh",
		86400,
		func() (io.ReadCloser, error) {
			url := "http://www.malware-domains.com/files/justdomains.zip"
			fileHandle, err := util.ReadZippedFileFromWeb(url)
			if err != nil {
				return nil, err
			}
			return fileHandle, nil
		},
	)
}
