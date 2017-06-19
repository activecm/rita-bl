package lists

import (
	"bufio"
	"io"

	"github.com/ocmdev/rita-blacklist2/list"
)

//customList gathers blacklisted ip addresses from myip.ms
type customList struct {
	meta       list.Metadata
	dataSource func() io.ReadCloser
}

//NewCustomList returns a new customList object
func NewCustomList(entryType list.BlacklistedEntryType, name string,
	dataFactory func() io.ReadCloser) list.List {
	return &customList{
		meta: list.Metadata{
			Types:     []list.BlacklistedEntryType{entryType},
			Name:      name,
			CacheTime: 86400,
		},
		dataSource: dataFactory,
	}
}

//GetMetadata returns the Metadata associated with this blacklist
func (m *customList) GetMetadata() list.Metadata {
	return m.meta
}

//SetMetadata sets the Metadata associated with this blacklist
func (m *customList) SetMetadata(meta list.Metadata) {
	m.meta = meta
}

//FetchData fetches the BlacklistedEntries associated with this list.
//This function must run the fetch in the background and immediately
//return a map of channels to read from.
func (m *customList) FetchData(entryMap list.BlacklistedEntryMap, errorsOut chan<- error) {
	entryType := m.GetMetadata().Types[0]
	defer close(entryMap[entryType])
	reader := m.dataSource()
	defer reader.Close()
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if scanner.Err() != nil {
			errorsOut <- scanner.Err()
			return
		}
		line := scanner.Text()

		//skip empty lines
		if len(line) == 0 {
			continue
		}
		entryMap[entryType] <- list.NewBlacklistedEntry(line, m)
	}
}
