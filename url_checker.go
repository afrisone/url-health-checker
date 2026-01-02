package main

import (
	"bufio"
	"flag"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

func main() {
	maxParallel := flag.Int("max", 3, "Max parallel goroutines at any given time")
	filePath := flag.String("file", "urls.txt", "The relative file path to the url file to scan")

	flag.Parse()

	timeout := 5 * time.Second
	file, err := os.Open(*filePath)

	if err != nil {
		log.Println("Error reading file:", err)
		os.Exit(1)
	}
	defer file.Close()

	wg := new(sync.WaitGroup)
	sem := make(chan struct{}, *maxParallel)

	scanner := bufio.NewScanner(file)

	client := &http.Client{Timeout: timeout}
	lineNum := 1

	for scanner.Scan() {
		url := scanner.Text()

		if url == "" {
			continue
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(url string, lineNum int) {
			defer wg.Done()
			defer func() { <-sem }()

			checkUrl(client, url, lineNum)
		}(url, lineNum)

		lineNum++
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading file, %v\n", err)
	}

	wg.Wait()
}

func checkUrl(client *http.Client, url string, lineNum int) {
	resp, httpErr := client.Head(url)

	if httpErr != nil {
		printResult(lineNum, url, httpErr)
		return
	}

	defer resp.Body.Close()

	printResult(lineNum, url, resp.Status)
}

func printResult(lineNum int, url string, result any) {
	log.Printf("    | %d %s  Result: %v\n", lineNum, url, result)
}
