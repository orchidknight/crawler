package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"crawler/crawler"
)

func main() {
	targetURL := flag.String("url", "", "Input target URL")
	outputFile := flag.String("out", "", "Input file for results")

	flag.Parse()
	if *targetURL == "" {
		fmt.Println("Please, specify target URL with -url flag")
		return
	}

	if *outputFile == "" {
		fmt.Println("Please, specify correct filename for results")
		return
	}

	parser, err := crawler.New(&crawler.CrawlerOpts{
		IsRecursive: true,
		TargetURL:   *targetURL,
	})
	if err != nil {
		fmt.Println("Please, specify correct target URL, error: ", err)
		return
	}

	start := time.Now()
	parser.ParseParallel()

	fmt.Printf("Got %d results\n", len(parser.Results()))

	err = saveURLsToFile(parser.Results(), *outputFile)
	if err != nil {
		fmt.Printf("Failed to save URLs to file: %v\n", err)
		return
	}

	fmt.Println("URLs successfully saved to results.json")
	elapsed := time.Since(start)
	fmt.Printf("Processing time: %s\n", elapsed)
}

// saveURLsToFile get urls and save it to file
func saveURLsToFile(urls []string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(urls)
	if err != nil {
		return fmt.Errorf("error encoding JSON: %v", err)
	}

	return nil
}
