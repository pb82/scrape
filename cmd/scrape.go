package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"scrape/api"
	"scrape/pkg/common"
	"scrape/pkg/scrape"
	"scrape/store"
	"scrape/version"
	"strings"
	"sync"
	"syscall"
)

func main() {
	printVersion := flag.Bool("version", false, "print version and exit")
	scrapeUrls := flag.String("scrape.urls", "", "list of urls to scrape")
	scrapeInterval := flag.String("scrape.interval", "10s", "scrape interval")
	sqliteFilename := flag.String("sqlite.file", "metrics.db", "sqlite database file")
	flag.Parse()

	if *printVersion {
		fmt.Println(fmt.Sprintf("scrape version %v", version.Version))
		os.Exit(0)
	}

	scrapeStatus := make(chan common.OperationResult)
	dbStatus := make(chan common.OperationResult)
	quitScrape := make(chan bool, 1)
	quitDb := make(chan bool, 1)
	sigs := make(chan os.Signal)
	samples := make(chan api.Sample, 512)
	wg := &sync.WaitGroup{}

	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGINT)
	go func() {
		sig := <-sigs
		switch sig {
		case syscall.SIGTERM, syscall.SIGABRT, syscall.SIGINT:
			quitScrape <- true
			quitDb <- true
		}
	}()

	sqlite, err := store.NewSqliteStore(*sqliteFilename, wg)
	if err != nil {
		panic(err)
	}

	sqlite.Run(samples, dbStatus, quitDb)

	if scrapeUrls != nil && *scrapeUrls != "" {
		var scrapeTargets []*url.URL
		urls := strings.Split(*scrapeUrls, ",")
		for _, u := range urls {
			scrapeTarget, err := url.Parse(strings.TrimSpace(u))
			if err != nil {
				panic(err)
			}
			scrapeTargets = append(scrapeTargets, scrapeTarget)
		}

		for _, scrapeTarget := range scrapeTargets {
			scraper, err := scrape.NewUrlScaper(scrapeTarget, *scrapeInterval, wg)
			if err != nil {
				panic(err)
			}
			scraper.Scrape(scrapeStatus, samples, quitScrape)
		}
	}

	wg.Wait()
}
