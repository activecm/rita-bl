package lists

import (
	"io"

	"github.com/activecm/rita-bl/list"
	"github.com/activecm/rita-bl/sources/lists/util"
)

//NewDNSBHList returns a new DNSBHList object
func NewDNSBHList() list.List {
	return NewLineSeparatedList(
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
