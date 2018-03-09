package lists

import (
	"bufio"
	"errors"
	"net/http"
	"strings"

	"github.com/activecm/rita-bl/list"
)

//mdlList gathers blacklisted ip addresses from myip.ms
type mdlList struct {
	meta list.Metadata
}

//NewMdlList returns a new mdlList object
func NewMdlList() list.List {
	return &mdlList{
		meta: list.Metadata{
			Types: []list.BlacklistedEntryType{
				list.BlacklistedIPType,
				list.BlacklistedURLType,
			},
			Name:      "mdl",
			CacheTime: 86400,
		},
	}
}

//GetMetadata returns the Metadata associated with this blacklist
func (m *mdlList) GetMetadata() list.Metadata {
	return m.meta
}

//SetMetadata sets the Metadata associated with this blacklist
func (m *mdlList) SetMetadata(meta list.Metadata) {
	m.meta = meta
}

//FetchData fetches the BlacklistedEntries associated with this list.
//This function must run the fetch in the background and immediately
//return a map of channels to read from.
func (m *mdlList) FetchData(entryMap list.BlacklistedEntryMap, errorsOut chan<- error) {
	defer close(entryMap[list.BlacklistedIPType])
	defer close(entryMap[list.BlacklistedURLType])
	mdlURL := "http://www.malwaredomainlist.com/mdlcsv.php"
	resp, err := http.Get(mdlURL)
	if err != nil {
		errorsOut <- err
		return
	}
	defer resp.Body.Close()

	alreadyBlacklisted := make(map[string]bool)

	scanner := bufio.NewScanner(resp.Body)
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

		// Can't just split on comma in case one of the fields contains a comma
		// Splitting on "," instead will insure we're splitting between fields
		// Also saves us the trouble of trimming quote marks on every field
		lineSplit := strings.Split(line, "\",\"")
		if len(lineSplit) < 9 {
			errorsOut <- errors.New("malformed line from MDL; missing field")
			continue
		}

		var blURL list.BlacklistedEntry
		//if url field is empty, the host field is a url >.<
		if lineSplit[1] == "-" {
			//if we've already blacklisted this url, skip it
			if _, ok := alreadyBlacklisted["http://"+lineSplit[2]]; ok {
				continue
			}

			//all of the url entries in this list are http urls
			blURL = list.NewBlacklistedEntry("http://"+lineSplit[2], m)
		} else {
			//if we've already blacklisted this url, skip it
			if _, ok := alreadyBlacklisted["http://"+lineSplit[1]]; ok {
				continue
			}

			//otherwise 1 is the url, and 2 is the ip
			blURL = list.NewBlacklistedEntry("http://"+lineSplit[1], m)

			//if we've haven't already blacklisted the IP, blacklist it
			if _, ok := alreadyBlacklisted[lineSplit[2]]; !ok {
				blIP := list.NewBlacklistedEntry(lineSplit[2], m)
				blIP.ExtraData["date"] = strings.TrimPrefix(lineSplit[0], "\"")
				blIP.ExtraData["type"] = lineSplit[4]
				blIP.ExtraData["country"] = strings.TrimSuffix(lineSplit[8], "\",")
				entryMap[list.BlacklistedIPType] <- blIP
				alreadyBlacklisted[lineSplit[2]] = true
			}
		}

		blURL.ExtraData["date"] = strings.TrimPrefix(lineSplit[0], "\"")
		blURL.ExtraData["type"] = lineSplit[4]
		blURL.ExtraData["country"] = strings.TrimSuffix(lineSplit[8], "\"")
		entryMap[list.BlacklistedURLType] <- blURL
		alreadyBlacklisted[blURL.Index] = true
	}

}
