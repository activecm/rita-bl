package mock

import (
	"encoding/binary"
	"net"
	"strings"

	"github.com/ocmdev/rita-blacklist2/list"
)

//DummyList provides mock data for rita-blacklist
type DummyList struct {
	meta list.Metadata
}

//NewDummyList returns a new DummyList object
func NewDummyList() *DummyList {
	return &DummyList{
		meta: list.Metadata{
			Types: []list.BlacklistedEntryType{
				list.BlacklistedIPType, list.BlacklistedHostnameType,
			},
			Name:      "Dummy",
			CacheTime: 0,
		},
	}
}

//GetMetadata returns the Metadata associated with this blacklist
func (d *DummyList) GetMetadata() list.Metadata {
	return d.meta
}

//SetMetadata sets the Metadata associated with this blacklist
func (d *DummyList) SetMetadata(m list.Metadata) {
	d.meta = m
}

//FetchData fetches the BlacklistedEntries associated with this list.
//This function must run the fetch in the background and immediately
//return a map of channels to read from.
func (d *DummyList) FetchData(entryMap list.BlacklistedEntryMap, errorsOut chan<- error) {
	var i uint32
	for i = 0; i < 100; i++ {
		bs := make([]byte, 4)
		binary.LittleEndian.PutUint32(bs, i)
		ipAddr := net.IPv4(bs[0], bs[1], bs[2], bs[3]).String()
		entry := list.NewBlacklistedEntry(ipAddr, d)
		entryMap[list.BlacklistedIPType] <- entry
	}
	close(entryMap[list.BlacklistedIPType])

	for _, line := range strings.Split(hostnames, "\n") {
		entryMap[list.BlacklistedHostnameType] <- list.NewBlacklistedEntry(line, d)
	}
	close(entryMap[list.BlacklistedHostnameType])
}

const hostnames string = `163.com
1688.com
accuweather.com
alexa.com
angelfire.com
blog.com
bloomberg.com
booking.com
boston.com
cdc.gov
cnbc.com
com.com
craigslist.org
creativecommons.org
dagondesign.com
dailymail.co.uk
dedecms.com
delicious.com
deliciousdays.com
dell.com
diigo.com
disqus.com
domainmarket.com
ebay.co.uk
ed.gov
eventbrite.com
fastcompany.com
fc2.com
feedburner.com
geocities.jp
github.io
google.cn
google.it
goo.ne.jp
hexun.com
home.pl
hostgator.com
hp.com
huffingtonpost.com
icio.us
ifeng.com
ihg.com
infoseek.co.jp
irs.gov
jigsy.com
live.com
livejournal.com
mapy.cz
marketwatch.com
mayoclinic.com
mit.edu
mozilla.org
mtv.com
narod.ru
nationalgeographic.com
netvibes.com
newyorker.com
noaa.gov
nydailynews.com
nyu.edu
oakley.com
ovh.net
plala.or.jp
prweb.com
salon.com
sfgate.com
springer.com
stanford.edu
state.tx.us
taobao.com
technorati.com
tinypic.com
topsy.com
typepad.com
ucla.edu
umn.edu
ustream.tv
va.gov
vistaprint.com
vkontakte.ru
weibo.com
xing.com
xrea.com
yahoo.co.jp
youtu.be`
