package sources

import (
	"archive/zip"
	"bufio"
	"bytes"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"github.com/ocmdev/rita-blacklist2/list"
)

//MyIPmsList gathers blacklisted ip addresses from myip.ms
type MyIPmsList struct {
	meta list.Metadata
}

//NewMyIPmsList returns a new MyIPmsList object
func NewMyIPmsList() *MyIPmsList {
	return &MyIPmsList{
		meta: list.Metadata{
			Types:     []list.BlacklistedEntryType{list.BlacklistedIPType},
			Name:      "myip.ms",
			CacheTime: 86400,
		},
	}
}

//GetMetadata returns the Metadata associated with this blacklist
func (m *MyIPmsList) GetMetadata() list.Metadata {
	return m.meta
}

//SetMetadata sets the Metadata associated with this blacklist
func (m *MyIPmsList) SetMetadata(meta list.Metadata) {
	m.meta = meta
}

//FetchData fetches the BlacklistedEntries associated with this list.
//This function must run the fetch in the background and immediately
//return a map of channels to read from.
func (m *MyIPmsList) FetchData(entryMap list.BlacklistedEntryMap, errorsOut chan<- error) {
	defer close(entryMap[list.BlacklistedIPType])

	resp, err := http.Get("https://myip.ms/files/blacklist/general/full_blacklist_database.zip")
	if err != nil {
		errorsOut <- err
		return
	}
	//read the file into ram
	buff := new(bytes.Buffer)
	io.Copy(buff, resp.Body)
	resp.Body.Close()

	//extract the zip archive
	buffer := buff.Bytes()
	buffReader := bytes.NewReader(buffer)
	zipReader, err := zip.NewReader(buffReader, int64(len(buffer)))
	if err != nil {
		errorsOut <- err
		return
	}

	//open the file inside
	fileHandle, err := zipReader.File[0].Open()
	if err != nil {
		errorsOut <- err
		return
	}
	defer fileHandle.Close()

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
