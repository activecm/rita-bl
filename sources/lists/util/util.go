package util

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
)

//ReadZippedFileFromWeb reads a .zip archive containing a single .
//This format is common for blacklist distribution.
func ReadZippedFileFromWeb(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	//open the file inside
	fileHandle, err := zipReader.File[0].Open()
	if err != nil {
		return nil, err
	}
	return fileHandle, nil
}
