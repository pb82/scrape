package scrape

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"scrape/api"
	"scrape/pkg/common"
	"scrape/pkg/ingest"
	"sync"
	"time"
)

type UrlScaper struct {
	scrapeUrl      *url.URL
	scrapeInterval time.Duration
	client         *http.Client
	wg             *sync.WaitGroup
}

func NewUrlScaper(u *url.URL, interval string, wg *sync.WaitGroup) (*UrlScaper, error) {
	duration, err := time.ParseDuration(interval)
	if err != nil {
		return nil, err
	}

	client := http.Client{
		Timeout: duration,
	}

	wg.Add(1)
	return &UrlScaper{
		scrapeUrl:      u,
		scrapeInterval: duration,
		client:         &client,
		wg:             wg,
	}, nil
}

func (s *UrlScaper) parseResponse(response []byte, samples chan<- api.Sample) error {
	scanner := ingest.NewScanner()
	tokens, err := scanner.Scan(string(response))
	if err != nil {
		return err
	}

	parser := ingest.NewParser(tokens)
	parsedSamples, err := parser.Parse()
	if err != nil {
		return err
	}

	for _, sample := range parsedSamples {
		samples <- sample
	}

	return nil
}

func (s *UrlScaper) scrapeInternal(samples chan<- api.Sample) error {
	req, err := http.NewRequest(http.MethodGet, s.scrapeUrl.String(), nil)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("unexpected status, expected 200 got %v", resp.StatusCode))
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return s.parseResponse(bytes, samples)
}

func (s *UrlScaper) Scrape(status chan<- common.OperationResult, samples chan<- api.Sample, quit <-chan bool) {
	go func() {
		for true {
			select {
			case <-quit:
				log.Print("[scrape] quit signal received")
				close(status)
				s.wg.Done()
				break
			default:
				start := time.Now()
				err := s.scrapeInternal(samples)
				elapsed := time.Since(start)
				if err != nil {
					status <- common.OperationResult{
						Status:  common.OperationStatusFailed,
						Message: err.Error(),
					}
				} else {
					status <- common.OperationResult{
						Status:  common.OperationStatusSuccess,
						Message: fmt.Sprintf("scrape of %v finished in %vms", s.scrapeUrl, elapsed.Milliseconds()),
					}
				}
				time.Sleep(s.scrapeInterval)
			}
		}
	}()
}
