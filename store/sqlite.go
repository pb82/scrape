package store

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"scrape/api"
	"scrape/pkg/common"
	"sync"
)

const tableTimeseries = `
create table if not exists 'timeseries' (
    id		INTEGER PRIMARY KEY AUTOINCREMENT,
    hash    TEXT,
    unique(hash)                                 
)
`

const tableSamples = `
create table if not exists 'samples' (
	timestamp		INTEGER,
	timeseries_id	INTEGER,
	value			REAL,
	UNIQUE(timestamp, timeseries_id)
)
`
const tableLabels = `
create table if not exists 'labels' (
	id 		INTEGER PRIMARY KEY,
	name 	TEXT,
	UNIQUE(name)
)
`

const tableTimeseriesLabels = `
create table if not exists 'timeseries_labels' (
	timeseries_id 		INTEGER,
	label_id 			INTEGER,
	label_value			TEXT
)
`

type SqliteStore struct {
	db *sql.DB
	wg *sync.WaitGroup
}

func createTables(db *sql.DB) error {
	_, err := db.Exec(tableTimeseries)
	if err != nil {
		return err
	}

	_, err = db.Exec(tableSamples)
	if err != nil {
		return err
	}

	_, err = db.Exec(tableLabels)
	if err != nil {
		return err
	}

	_, err = db.Exec(tableTimeseriesLabels)
	if err != nil {
		return err
	}

	return nil
}

func NewSqliteStore(filename string, wg *sync.WaitGroup) (*SqliteStore, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	err = createTables(db)
	if err != nil {
		return nil, err
	}

	wg.Add(1)
	return &SqliteStore{
		db: db,
		wg: wg,
	}, nil
}

func (s *SqliteStore) Run(samples <-chan api.Sample, status chan<- common.OperationResult, quit <-chan bool) {
	go func() {
		for true {
			select {
			case <-quit:
				log.Print("[sqlite] quit signal received")
				s.db.Close()
				close(status)
				s.wg.Done()
				break
			case _ = <-samples:
				// log.Printf("[sqlite] sample received: %v", s.Labels)
			}
		}
	}()
}
