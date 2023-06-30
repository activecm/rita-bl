package database

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"net"
	"sync"
	"text/template"
	"time"

	clickhouse "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/activecm/rita-bl/list"
)

const chListTable string = "Lists"

//go:embed clickhouse_sql/select/lists.sql
var chSelectRegisteredLists string

//go:embed clickhouse_sql/create_table/lists.sql
var chCreateListsTableQuery string

//go:embed clickhouse_sql/create_table/templates/entries.sql
var chCreateEntryTableQueryTemplateDef string

var chCreateEntryTableQueryTemplate *template.Template = template.Must(template.New("chCreateEntryTableQueryTemplate").Parse(chCreateEntryTableQueryTemplateDef))

func chCreateEntryTableQuery(entryType list.BlacklistedEntryType) string {
	var tmplBuff bytes.Buffer
	tmplInputs := struct{ EntryType string }{
		EntryType: string(entryType),
	}
	chCreateEntryTableQueryTemplate.Execute(&tmplBuff, tmplInputs)
	return tmplBuff.String()
}

//go:embed clickhouse_sql/select/templates/entries_by_index.sql
var chSelectEntriesByIndexQueryTemplateDef string

var chSelectEntriesByIndexQueryTemplate *template.Template = template.Must(template.New("chSelectEntriesByIndexQueryTemplate").Parse(chSelectEntriesByIndexQueryTemplateDef))

func chSelectEntriesByIndexQuery(entryType list.BlacklistedEntryType) string {
	var tmplBuff bytes.Buffer
	tmplInputs := struct{ EntryType string }{
		EntryType: string(entryType),
	}
	chCreateEntryTableQueryTemplate.Execute(&tmplBuff, tmplInputs)
	return tmplBuff.String()
}

type clickhouseDB struct {
	connection clickhouse.Conn
	database   string
	ctx        context.Context
}

func NewClickhouseDB(ctx context.Context, addr []string, auth clickhouse.Auth, clientLogger func(string), db string) (Handle, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: addr,
		Auth: auth,
		DialContext: func(ctx context.Context, addr string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "tcp", addr)
		},
		Debug: true,
		Debugf: func(format string, v ...interface{}) {
			clientLogger(fmt.Sprintf(format, v...))
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout:          1 * time.Minute,
		ReadTimeout:          60 * time.Minute,
		MaxOpenConns:         15,
		MaxIdleConns:         15,
		ConnMaxLifetime:      60 * time.Minute,
		ConnOpenStrategy:     clickhouse.ConnOpenInOrder,
		BlockBufferSize:      10,
		MaxCompressionBuffer: 10240,
		ClientInfo: clickhouse.ClientInfo{ // optional, please see Client info section in the README.md
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: "rita-bl", Version: "0.1"},
			},
		},
	})
	return &clickhouseDB{
		ctx:        ctx,
		connection: conn,
		database:   db,
	}, err
}

// GetRegisteredLists retrieves all of the lists registered with the database
func (c *clickhouseDB) GetRegisteredLists() ([]list.Metadata, error) {
	var lists []list.Metadata
	var listBuff list.Metadata

	rowIter, err := c.connection.Query(c.ctx, chSelectRegisteredLists)
	if err != nil {
		return []list.Metadata{}, nil
	}
	defer rowIter.Close()

	for rowIter.Next() {
		listBuff = list.Metadata{}
		err = rowIter.Scan(&listBuff.Name, &listBuff.Types, &listBuff.LastUpdate, &listBuff.CacheTime)
		if err != nil {
			return []list.Metadata{}, err
		}
		lists = append(lists, listBuff)
	}

	return lists, nil
}

// RegisterList registers a new threat intel source with the database
func (c *clickhouseDB) RegisterList(l list.Metadata) error {
	// ensure table which holds the registered lists exists
	err := c.connection.Exec(c.ctx, chCreateListsTableQuery)
	if err != nil {
		return err
	}

	// ensure tables exist for each of the entry types associated with the list
	for _, listType := range l.Types {
		err := c.connection.Exec(c.ctx, chCreateEntryTableQuery(listType))
		if err != nil {
			return err
		}
	}

	batch, err := c.connection.PrepareBatch(context.Background(), "INSERT INTO "+chListTable)
	if err != nil {
		return err
	}

	err = batch.Append(l.Name, l.Types, time.Unix(l.LastUpdate, 0), l.CacheTime)
	if err != nil {
		return err
	}

	err = batch.Send()
	if err != nil {
		return err
	}

	return nil
}

// RemoveList removes an existing threat intel source from the database
func (c *clickhouseDB) RemoveList(l list.Metadata) error {
	err := c.ClearCache(l)

	if err != nil {
		return err
	}

	mutationsSyncCtx := clickhouse.Context(c.ctx, clickhouse.WithSettings(clickhouse.Settings{
		"mutations_sync": 1,
	}))

	err = c.connection.Exec(mutationsSyncCtx, "DELETE FROM "+chListTable+" WHERE Name=@name", clickhouse.Named("name", l.Name))
	if err != nil {
		return err
	}
	return nil
}

// UpdateListMetadata updates the metadata of an existing threat intel list
func (c *clickhouseDB) UpdateListMetadata(l list.Metadata) error {
	//just perform an insert since this is a replacing merge tree
	batch, err := c.connection.PrepareBatch(context.Background(), "INSERT INTO "+chListTable)
	if err != nil {
		return err
	}

	err = batch.Append(l.Name, l.Types, time.Unix(l.LastUpdate, 0), l.CacheTime)
	if err != nil {
		return err
	}

	err = batch.Send()
	if err != nil {
		return err
	}
	return nil
}

// ClearCache clears old entries for a given list
func (c *clickhouseDB) ClearCache(l list.Metadata) error {
	mutationsSyncCtx := clickhouse.Context(c.ctx, clickhouse.WithSettings(clickhouse.Settings{
		"mutations_sync": 1,
	}))
	for _, entryType := range l.Types {
		err := c.connection.Exec(mutationsSyncCtx, "DELETE FROM "+string(entryType)+" WHERE List=@name", clickhouse.Named("name", l.Name))
		if err != nil {
			return err
		}
	}
	return nil
}

// InsertEntries inserts entries from a list into the database
func (c *clickhouseDB) InsertEntries(entryType list.BlacklistedEntryType,
	entries <-chan list.BlacklistedEntry, wg *sync.WaitGroup, errorsOut chan<- error) {

	defer wg.Done()

	i := 0
	buffSize := 100000
	bulk, err := c.connection.PrepareBatch(c.ctx, "INSERT INTO "+string(entryType))
	if err != nil {
		errorsOut <- err
	}
	for entry := range entries {
		err = bulk.Append(entry.List.GetMetadata().Name, entry.Index, entry.ExtraData)
		if err != nil {
			errorsOut <- err
		}
		i++
		if i == buffSize {
			err = bulk.Send()
			if err != nil {
				errorsOut <- err
			}
			i = 0
			bulk, err = c.connection.PrepareBatch(c.ctx, "INSERT INTO "+string(entryType))
			if err != nil {
				errorsOut <- err
			}
		}
	}
	if i != 0 {
		err := bulk.Send()
		if err != nil {
			errorsOut <- err
		}
	}

}

// FindEntries finds entries of a given type and index
func (c *clickhouseDB) FindEntries(dataType list.BlacklistedEntryType, index string) ([]BlacklistResult, error) {
	findEntriesQuery := chSelectEntriesByIndexQuery(dataType)

	var entryMatches []BlacklistResult
	var entryBuff BlacklistResult

	rowIter, err := c.connection.Query(c.ctx, findEntriesQuery, clickhouse.Named("Index", index))
	if err != nil {
		return []BlacklistResult{}, nil
	}
	defer rowIter.Close()

	for rowIter.Next() {
		entryBuff = BlacklistResult{}
		err = rowIter.Scan(&entryBuff.List, &entryBuff.Index, &entryBuff.ExtraData)
		if err != nil {
			return []BlacklistResult{}, err
		}
		entryMatches = append(entryMatches, entryBuff)
	}

	return entryMatches, nil
}
