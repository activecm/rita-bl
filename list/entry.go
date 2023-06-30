package list

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type (
	//BlacklistedEntry is an entry of a List with a given BlacklistedType
	BlacklistedEntry struct {
		//Index is the main data held by this entry
		Index string
		//List is the source list
		List List
		//ExtraData contains extra information this blacklist source provides
		ExtraData map[string]string
	}

	//BlacklistedEntryType is a string representing which type of data
	//is held in the Index field of a BlacklistedEntry
	BlacklistedEntryType string
)

// NewBlacklistedEntry creates a new BlacklistedEntry
func NewBlacklistedEntry(index string, source List) BlacklistedEntry {
	return BlacklistedEntry{
		Index:     index,
		List:      source,
		ExtraData: make(map[string]string),
	}
}

// entryTypeValidators is a map of entry types to functions which validate them
var entryTypeValidators map[BlacklistedEntryType]func(string) error

// BlacklistedHostnameType should be added to the metadata types array
// in order to return hostnames from a list
const BlacklistedHostnameType BlacklistedEntryType = "hostname"

func validateHostname(hostname string) error {
	if len(hostname) > 253 || len(hostname) < 1 {
		return errors.New("hostnames must be less than 254 characters long")
	}
	specialCharIdx := strings.IndexFunc(hostname, func(r rune) bool {
		return !((r >= 'A' && r <= 'Z') ||
			(r >= 'a' && r <= 'z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '.' || r == '_')
	})
	if specialCharIdx != -1 {
		return fmt.Errorf("invalid char %c found in hostname label", hostname[specialCharIdx])
	}

	labels := strings.Split(hostname, ".")
	for _, l := range labels {
		if len(l) > 63 || len(l) < 1 {
			return errors.New("hostname labels must be between 1 and 63 characters long")
		}
		if l[0] == '-' {
			return errors.New("hostnames labels must not start with a minus sign")
		}
		if l[len(l)-1] == '-' {
			return errors.New("hostname labels must not end with a minus sign")
		}
	}

	return nil
}

// BlacklistedIPType should be added to the metadata types array
// in order to return ips from a list
const BlacklistedIPType BlacklistedEntryType = "ip"

func validateIP(ip string) error {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return errors.New("failed to parse ip address")
	}
	return nil
}

// BlacklistedURLType should be added to the metadata types array
// in order to return hostnames from a list
const BlacklistedURLType BlacklistedEntryType = "url"

func validateURL(str string) error {
	u, err := url.ParseRequestURI(str)
	if err != nil {
		return err
	}
	portSplit := strings.Split(u.Host, ":")
	if len(portSplit) > 1 {
		portNumber, err := strconv.ParseInt(portSplit[1], 10, 64)
		if err != nil {
			return errors.New("port specifier found but unable to parse port in url")
		}
		if portNumber < 0 || portNumber > 65536 {
			return errors.New("invalid port number specified in url")
		}
		u.Host = portSplit[0]
	}
	isValidHost := validateHostname(u.Host) == nil || validateIP(u.Host) == nil
	if !isValidHost {
		return errors.New("invalid host for url")
	}
	return nil
}

func init() {
	entryTypeValidators = make(map[BlacklistedEntryType]func(string) error)
	entryTypeValidators[BlacklistedHostnameType] = validateHostname
	entryTypeValidators[BlacklistedIPType] = validateIP
	entryTypeValidators[BlacklistedURLType] = validateURL
}
