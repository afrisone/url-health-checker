package main

import (
	"bufio"
	"flag"
	"fmt"
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
		fmt.Println("Error reading file:", err)
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
		fmt.Printf("Error reading file, %v\n", err)
	}

	wg.Wait()
}

func checkUrl(client *http.Client, url string, lineNum int) {
	fmt.Println("Checking....", url)
	resp, httpErr := client.Head(url)

	if httpErr != nil {
		fmt.Printf("    | %d  Result: Error checking status\n", lineNum)
		fmt.Println("        |", httpErr)
		return
	}

	defer resp.Body.Close()

	fmt.Printf("    | %d %s  Result: %s\n", lineNum, url, resp.Status)
}
