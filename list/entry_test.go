package list

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name  string
	index string
	valid bool
}

func TestValidateHostname(t *testing.T) {
	indexType := string(BlacklistedHostnameType)
	validator := entryTypeValidators[BlacklistedHostnameType]
	testCases := []testCase{
		{"One word", "simple", true},
		{"Dot", ".", false},
		{"Dot Dot", "..", false},
		{"Google", "google.com", true},
		{"Minus End", "test-.com", false},
		{"Minus Middle", "test-thing.com", true},
		{"Minus Start", "-test.com", false},
		{"Special Characters", "test!.com", false},
		{"Subdomain", "test.test.com", true},
		{"Long label", "asdfasdfasdfasdfsadfasdfasdfasdfasdfsadfasdfsadfasds" +
			"dfasdfasdfasdfasdfasdfasdfsadfsadfasdfaasdfasdffsadfasdf.com", false},
		{"Long name",
			"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
				"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
				"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
				"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
				"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
				"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			false},
	}
	for _, c := range testCases {
		t.Run(fmt.Sprintf("%s: %s", indexType, c.name), func(test *testing.T) {
			if c.valid {
				assert.Nil(test, validator(c.index))
			} else {
				assert.NotNil(test, validator(c.index))
			}
		})
	}
}

func TestValidateIP(t *testing.T) {
	indexType := string(BlacklistedIPType)
	validator := entryTypeValidators[BlacklistedIPType]
	testCases := []testCase{
		{"0.0.0.0", "0.0.0.0", true},
		{"Missing octet", "0.0.0", false},
		{"Missing octet w/ dot", "0.0.0.", false},
		{"Missing octets", "0..0", false},
		{"Negative octet", "0.0.-123.0", false},
		{"Large octet", "0.300.0.0", false},
		{"Standard IPv4", "53.124.123.1", true},
		{"Standard IPv6", "2001:db8::68", true},
		{"Missing Sections IPv6", "2001:db8:68", false},
	}
	for _, c := range testCases {
		t.Run(fmt.Sprintf("%s: %s", indexType, c.name), func(test *testing.T) {
			if c.valid {
				assert.Nil(test, validator(c.index))
			} else {
				assert.NotNil(test, validator(c.index))
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	indexType := string(BlacklistedURLType)
	validator := entryTypeValidators[BlacklistedURLType]
	testCases := []testCase{
		{"Missing protocol", "google.com/index.htm", false},
		{"Missing path", "http://google.com", true},
		{"Missing host", "http:///index.htm", false},
		{"Valid", "http://google.com/index.html", true},
		{"IP address with port no protocol", "127.0.0.1:8080", false},
		{"IP address with port", "https://127.0.0.1:8080", true},
		{"Colon no port", "https://127.0.0.1:/asdf", false},
		{"IP address invalid port", "mongodb://127.0.0.1:70000", false},
	}
	for _, c := range testCases {
		t.Run(fmt.Sprintf("%s: %s", indexType, c.name), func(test *testing.T) {
			if c.valid {
				assert.Nil(test, validator(c.index))
			} else {
				assert.NotNil(test, validator(c.index))
			}
		})
	}
}
