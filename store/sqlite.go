package store

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"scrape/api"
	"scrape/pkg/promql"
	"sync"
	"time"
)

const tableTimeseries = `
create table if not exists 'timeseries' (
    id		INTEGER PRIMARY KEY AUTOINCREMENT,
    hash    TEXT,
    UNIQUE(hash)                                 
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

type SqliteColumn struct {
	Name  string
	Value string
}

type SqliteRow struct {
	Columns []SqliteColumn
}

type SqliteResult struct {
	Rows    []SqliteRow
	Success bool
}

type SqliteStore struct {
	db *sql.DB
	wg *sync.WaitGroup
}

func getTimeseries(labels []api.Label) string {
	hf := sha256.New()
	for _, label := range labels {
		hf.Write([]byte(label.Name))
		hf.Write([]byte(label.Value))
	}
	return fmt.Sprintf("%x", hf.Sum(nil))
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

func runQuery(db *sql.DB, query promql.PromQlASTElement) *SqliteResult {
	result := &SqliteResult{}
	result.Success = true

	query.Eval(db)

	return result
}

func insertSample(db *sql.DB, sample *api.Sample) error {
	hash := getTimeseries(sample.Labels)
	stmt, err := db.Prepare(`insert into timeseries(hash) values(?) on conflict(hash) do update set hash = hash returning id;`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	var timeseriesId = 0
	row := stmt.QueryRow(hash)
	row.Scan(&timeseriesId)

	labelIds := map[string]int{}
	for _, label := range sample.Labels {
		stmt, err = db.Prepare(`
insert into labels(name) values(?) on conflict(name) do update set name = name returning id
`)
		if err != nil {
			return err
		}
		var labelId = 0
		row := stmt.QueryRow(label.Name)
		row.Scan(&labelId)
		labelIds[label.Name] = labelId
		stmt.Close()
	}

	stmt, err = db.Prepare(`
insert or ignore into samples(timestamp, timeseries_id, value) values(?, ?, ?)
`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	result, err := stmt.Exec(time.Now().Unix(), timeseriesId, sample.Value)
	if err != nil {
		return err
	}

	// no new metric was inserted, exit early
	// we can only capture metrics at 1s resolution
	metricId, err := result.LastInsertId()
	if err != nil {
		return err
	}
	if metricId <= 0 {
		return nil
	}

	for _, label := range sample.Labels {
		stmt, err = db.Prepare(`
insert into timeseries_labels(timeseries_id, label_id, label_value) values(?,?,?)
`)
		if err != nil {
			return err
		}
		_, err = stmt.Exec(timeseriesId, labelIds[label.Name], label.Value)
		if err != nil {
			return err
		}
		stmt.Close()
	}

	return nil
}

func NewSqliteStore(filename string, wg *sync.WaitGroup) (*SqliteStore, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%v?cache=shared&mode=rwc&_journal_mode=WAL", filename))
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

func (s *SqliteStore) Run(samples <-chan api.Sample, quit <-chan bool, queries <-chan promql.PromQlASTElement) {
	go func() {
		for true {
			select {
			case <-quit:
				log.Print("[sqlite] quit signal received")
				s.db.Close()
				s.wg.Done()
				break
			case sample := <-samples:
				err := insertSample(s.db, &sample)
				if err != nil {
					log.Printf("[sqlite] error adding sample: %v", err)
				}
			case query := <-queries:
				_ = runQuery(s.db, query)
			}
		}
	}()
}
