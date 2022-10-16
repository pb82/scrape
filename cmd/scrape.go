package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"scrape/api"
	"scrape/pkg/promql"
	"scrape/pkg/scrape"
	"scrape/store"
	"scrape/version"
	"strings"
	"sync"
	"syscall"
	"time"
)

func startTicker(tick chan<- bool, duration time.Duration) {
	/*
		go func() {
			for true {
				tick <- true
				time.Sleep(duration)
			}
		}()
	*/
	go func() {
		tick <- true
	}()

}

func main() {
	printVersion := flag.Bool("version", false, "print version and exit")
	scrapeUrls := flag.String("scrape.urls", "", "list of urls to scrape")
	scrapeInterval := flag.String("scrape.interval", "10s", "scrape interval")
	sqliteFilename := flag.String("sqlite.file", "metrics.db", "sqlite database file")
	interactive := flag.Bool("interactive", false, "interactive mode")
	flag.Parse()

	if *printVersion {
		fmt.Println(fmt.Sprintf("scrape version %v", version.Version))
		os.Exit(0)
	}

	quitScrape := make(chan bool, 1)
	quitDb := make(chan bool, 1)
	sigs := make(chan os.Signal)
	samples := make(chan api.Sample, 512)
	tick := make(chan bool)
	queries := make(chan promql.PromQlASTElement)
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

	sqlite.Run(samples, quitDb, queries)

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
			scraper, err := scrape.NewUrlScaper(scrapeTarget, wg)
			if err != nil {
				panic(err)
			}
			scraper.Scrape(samples, quitScrape, tick)
		}
	}

	duration, err := time.ParseDuration(*scrapeInterval)
	if err != nil {
		panic(err)
	}
	startTicker(tick, duration)

	if *interactive {

		go func() {
			for true {
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("> ")
				text, _ := reader.ReadString('\n')
				if text == "" {
					continue
				}
				promQlScanner := promql.NewPromQlScanner(text)
				promqlTokens := promQlScanner.ScanPromQl()
				promQlParser := promql.NewPromQlParser(promqlTokens)
				ast, err := promQlParser.Parse()
				if err != nil {
					fmt.Println(fmt.Sprintf("[error] %v", err.Error()))
					continue
				}
				queries <- ast
			}
		}()
	}

	wg.Wait()
}
