package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type TrackEvent struct {
	EmailHash string `json:"email_hash"`
	EventHash string `json:"event_hash"`
}

var (
	backendURL string
	totalUsers int
	batchSize  int
	workers    int
	events     int // total events to generate
	requests   int
)

var client = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 1000,
		IdleConnTimeout:     90 * time.Second,
	},
}

func init() {
	flag.StringVar(&backendURL, "url", "http://localhost:8080/track", "Backend URL to send track events to")
	flag.IntVar(&totalUsers, "users", 1000000, "Total number of unique users to simulate")
	flag.IntVar(&batchSize, "batch-size", 1, "Number of events to send in each batch request")
	flag.IntVar(&workers, "workers", 100, "Number of concurrent workers to generate and send events")
	flag.IntVar(&events, "events", 1000000, "Total number of events to generate")
}

func main() {
	flag.Parse()

	// Calculate requests based on parsed flags
	requests = events / batchSize
	if requests == 0 {
		log.Fatal("Calculated 'requests' is zero. Ensure 'events' is greater than or equal to 'batch-size'.")
	}

	rand.Seed(time.Now().UnixNano())
	wg := sync.WaitGroup{}

	log.Printf("Starting load generation with parameters:\n")
	log.Printf("  Backend URL: %s\n", backendURL)
	log.Printf("  Total Users: %d\n", totalUsers)
	log.Printf("  Batch Size: %d\n", batchSize)
	log.Printf("  Workers: %d\n", workers)
	log.Printf("  Total Events: %d\n", events)
	log.Printf("  Total Requests: %d\n", requests)

	for {
		wg.Add(workers)

		for w := 0; w < workers; w++ {
			go func(workerID int) {
				defer wg.Done()
				eventsPerWorker := requests / workers
				if requests%workers != 0 && workerID == workers-1 { // Distribute remainder to the last worker
					eventsPerWorker += requests % workers
				}

				for i := 0; i < eventsPerWorker; i++ {
					batch := make([]TrackEvent, batchSize)
					for j := range batch {
						userID := rand.Intn(totalUsers)
						batch[j] = TrackEvent{
							EmailHash: fmt.Sprintf("user-%d@example.com", userID),
							EventHash: fmt.Sprintf("event-%d-%d", userID, rand.Int63()),
						}
					}
					sendBatch(batch)
				}
			}(w)
		}

		wg.Wait()
		log.Println("Load generation complete. Restarting for continuous load...")
		// Optional: Add a sleep here if you don't want it to restart immediately
		// time.Sleep(5 * time.Second)
	}
}

func sendBatch(events []TrackEvent) {
	data, err := json.Marshal(events)
	if err != nil {
		fmt.Println("marshal error:", err)
		return
	}

	req, _ := http.NewRequest("POST", backendURL, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("HTTP error:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		fmt.Println("unexpected status:", resp.Status)
	}
}
