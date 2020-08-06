package lists

import (
	"io"
	"net/http"

	"github.com/activecm/rita-bl/list"
)

func NewFeodoList() list.List {
	return NewLineSeparatedList(
		list.BlacklistedIPType,
		"feodo tracker",
		86400,
		func() (io.ReadCloser, error) {
			url := "https://feodotracker.abuse.ch/downloads/ipblocklist.txt"
			resp, err := http.Get(url)
			if err != nil {
				return nil, err
			}
			return resp.Body, nil
		},
	)
}
