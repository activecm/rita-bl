package rpc

import (
	"io"

	"github.com/google/safebrowsing"
	"github.com/ocmdev/rita-bl/database"
	"github.com/ocmdev/rita-bl/list"
)

type safeBrowsingURLsRPC struct {
	safebrowser *safebrowsing.SafeBrowser
}

//GetType returns the type of data that this RPC can check
func (s safeBrowsingURLsRPC) GetType() list.BlacklistedEntryType {
	return list.BlacklistedURLType
}

//Check checks a set of indexes against the rpc and returns a map
//of the indexes to their results
func (s safeBrowsingURLsRPC) Check(urls ...string) (map[string]database.BlacklistResult, error) {
	//threats is a 2d array indexed by the index of the urls and then by the
	//individual results for the url
	threats, err := s.safebrowser.LookupURLs(urls)
	if err != nil {
		return nil, err
	}

	entries := make(map[string]database.BlacklistResult)
	for urlIndex, urlLookup := range threats {
		//if there were hits for this url
		if len(urlLookup) > 0 {
			url := urls[urlIndex]
			entries[url] = database.BlacklistResult{
				Index:     url,
				List:      "google-safebrowsing",
				ExtraData: make(map[string]interface{}),
			}
			entries[url].ExtraData["ThreatTypes"] = make([]string, len(urlLookup))
			for _, threat := range urlLookup {
				entries[url].ExtraData["ThreatTypes"] = append(
					entries[url].ExtraData["ThreatTypes"].([]string),
					threat.ThreatType.String(),
				)
			}
		}
	}
	return entries, nil
}

//NewGoogleSafeBrowsingURLsRPC creates a rita-blacklist RPC for google's
//safebrowsing package for go
func NewGoogleSafeBrowsingURLsRPC(apiKey, dbPath string, logger io.Writer) (RPC, error) {
	config := safebrowsing.Config{
		APIKey: apiKey,
		DBPath: dbPath,
		Logger: logger,
	}
	sb, err := safebrowsing.NewSafeBrowser(config)
	if err != nil {
		return nil, err
	}
	return safeBrowsingURLsRPC{
		safebrowser: sb,
	}, nil
}
