package lists

import (
	"bufio"
	"errors"
	"strconv"
	"strings"
	"unicode"

	"github.com/ocmdev/rita-blacklist2/list"
	"github.com/ocmdev/rita-blacklist2/sources/lists/util"
)

//myIPmsList gathers blacklisted ip addresses from myip.ms
type myIPmsList struct {
	meta list.Metadata
}

//NewMyIPmsList returns a new MyIPmsList object
func NewMyIPmsList() list.List {
	return &myIPmsList{
		meta: list.Metadata{
			Types:     []list.BlacklistedEntryType{list.BlacklistedIPType},
			Name:      "myip.ms",
			CacheTime: 86400,
		},
	}
}

//GetMetadata returns the Metadata associated with this blacklist
func (m *myIPmsList) GetMetadata() list.Metadata {
	return m.meta
}

//SetMetadata sets the Metadata associated with this blacklist
func (m *myIPmsList) SetMetadata(meta list.Metadata) {
	m.meta = meta
}

//FetchData fetches the BlacklistedEntries associated with this list.
//This function must run the fetch in the background and immediately
//return a map of channels to read from.
func (m *myIPmsList) FetchData(entryMap list.BlacklistedEntryMap, errorsOut chan<- error) {
	defer close(entryMap[list.BlacklistedIPType])
	myIPmsURL := "https://myip.ms/files/blacklist/general/full_blacklist_database.zip"
	fileHandle, err := util.ReadZippedFileFromWeb(myIPmsURL)
	if err != nil {
		errorsOut <- err
		return
	}

	lineReader := bufio.NewScanner(fileHandle)
	for lineReader.Scan() {
		if lineReader.Err() != nil {
			errorsOut <- lineReader.Err()
			return
		}

		line := lineReader.Text()

		//remove whitespace
		line = strings.Map(func(r rune) rune {
			if unicode.IsSpace(r) {
				return -1
			}
			return r
		}, line)

		//skip comments and empty lines
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		//replace comment hash with comma to make csv
		line = strings.Replace(line, "#", ",", -1)
		lineSplit := strings.Split(line, ",")

		if len(lineSplit) < 5 {
			errorsOut <- errors.New("malformed line from myip.ms; missing field")
			continue
		}

		ret := list.NewBlacklistedEntry(lineSplit[0], m)
		ret.ExtraData["date"] = lineSplit[1]
		ret.ExtraData["host"] = lineSplit[2]
		ret.ExtraData["country"] = lineSplit[3]
		id, err := strconv.Atoi(lineSplit[4])
		if err != nil {
			id = -1
		}
		ret.ExtraData["id"] = id

		entryMap[list.BlacklistedIPType] <- ret
	}
}
