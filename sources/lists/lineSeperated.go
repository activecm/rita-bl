package lists

import (
	"bufio"
	"io"

	"github.com/activecm/rita-bl/list"
)

//lineSeperatedList gathers blacklisted ip addresses from myip.ms
type lineSeperatedList struct {
	meta       list.Metadata
	dataSource func() (io.ReadCloser, error)
}

//NewLineSeperatedList returns a new lineSeperatedList object
func NewLineSeperatedList(entryType list.BlacklistedEntryType, name string,
	cacheTime int64, dataFactory func() (io.ReadCloser, error)) list.List {
	return &lineSeperatedList{
		meta: list.Metadata{
			Types:     []list.BlacklistedEntryType{entryType},
			Name:      name,
			CacheTime: cacheTime,
		},
		dataSource: dataFactory,
	}
}

//GetMetadata returns the Metadata associated with this blacklist
func (m *lineSeperatedList) GetMetadata() list.Metadata {
	return m.meta
}

//SetMetadata sets the Metadata associated with this blacklist
func (m *lineSeperatedList) SetMetadata(meta list.Metadata) {
	m.meta = meta
}

//FetchData fetches the BlacklistedEntries associated with this list.
//This function must run the fetch in the background and immediately
//return a map of channels to read from.
func (m *lineSeperatedList) FetchData(entryMap list.BlacklistedEntryMap, errorsOut chan<- error) {
	entryType := m.GetMetadata().Types[0]
	defer close(entryMap[entryType])
	reader, err := m.dataSource()
	if err != nil {
		errorsOut <- err
		return
	}
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
