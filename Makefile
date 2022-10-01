BINARY=scrape

build:
	go build -o ./$(BINARY) cmd/scrape.go

clean:
	rm -f $(BINARY)