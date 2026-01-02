package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	maxParallel := flag.Int("max", 3, "Max parallel goroutines at any given time")
	filePath := flag.String("file", "urls.txt", "The relative file path to the url file to scan")

	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, *filePath, *maxParallel); err != nil {
		log.Fatalf("App error: %v\n", err)
	}
}

func run(ctx context.Context, filePath string, maxParallel int) error {
	file, err := os.Open(filePath)

	if err != nil {
		log.Println("Error opening file:", err)
		return err
	}
	defer file.Close()

	wg := new(sync.WaitGroup)
	sem := make(chan struct{}, maxParallel)

	scanner := bufio.NewScanner(file)

	client := &http.Client{Timeout: 5 * time.Second}
	lineNum := 1

	for scanner.Scan() {
		URL := scanner.Text()

		if URL == "" {
			<-sem
			continue
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case sem <- struct{}{}:
			// semaphore aquired, pause if full
		}

		wg.Add(1)

		go func(URL string, lineNum int) {
			defer wg.Done()
			defer func() { <-sem }()

			checkURL(ctx, client, URL, lineNum)
		}(URL, lineNum)

		lineNum++
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error reading file", err)
	}

	wg.Wait()

	return scanner.Err()
}

func checkURL(ctx context.Context, client *http.Client, URL string, lineNum int) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, URL, nil)

	if err != nil {
		printResult(lineNum, URL, err)
		return
	}

	resp, err := client.Do(req)

	if err != nil {
		printResult(lineNum, URL, err)
		return
	}

	defer resp.Body.Close()

	printResult(lineNum, URL, resp.Status)
}

func printResult(lineNum int, URL string, result any) {
	log.Printf("    | %d %s  Result: %v\n", lineNum, URL, result)
}
