package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
)

var (
	serverAddr   string
	numClients   int
	testDuration time.Duration
)

func init() {
	flag.StringVar(&serverAddr, "addr", "", "Server address to test")
	flag.IntVar(&numClients, "clients", 1, "Number of concurrent clients")
	flag.DurationVar(&testDuration, "duration", 10*time.Second, "Test duration")
}

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

type Stats struct {
	mu          sync.Mutex
	total       int
	success     int
	errors      int
	statusCodes map[int]int
	latencies   []time.Duration
}

func (s *Stats) Record(status int, err error, latency time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.total++
	if err != nil {
		s.errors++
	} else {
		s.statusCodes[status]++
		if status == http.StatusOK {
			s.success++
		}
	}
	s.latencies = append(s.latencies, latency)
}

func (s *Stats) PrintReport(b *testing.B) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b.Logf("\nTest Report:")
	b.Logf("Total Requests: %d", s.total)
	b.Logf("Successful: %d (%.2f%%)", s.success, 100*float64(s.success)/float64(s.total))
	b.Logf("Errors: %d", s.errors)
	b.Logf("Status Code Distribution:")
	for code, count := range s.statusCodes {
		b.Logf("  %d: %d", code, count)
	}

	if len(s.latencies) > 0 {
		var totalLatency time.Duration
		for _, l := range s.latencies {
			totalLatency += l
		}
		avgLatency := totalLatency / time.Duration(len(s.latencies))
		b.Logf("Average Latency: %v", avgLatency)
	}
}

func BenchmarkServer(b *testing.B) {
	if serverAddr == "" {
		b.Fatal("Server address is required (use -addr flag)")
	}

	stats := &Stats{
		statusCodes: make(map[int]int),
	}

	var wg sync.WaitGroup
	done := make(chan struct{})

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			ip := fmt.Sprintf("192.168.0.%d", clientID)
			client := &http.Client{Timeout: 5 * time.Second}

			// генерируем разный интервал между запросами для каждого клиента (100-1000ms)
			interval := time.Duration(100+(clientID%10)*100) * time.Millisecond
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					start := time.Now()
					req, _ := http.NewRequest("GET", serverAddr, nil)
					req.Header.Set("X-Forwarded-For", ip)

					resp, err := client.Do(req)
					latency := time.Since(start)

					var status int
					if resp != nil {
						status = resp.StatusCode
						resp.Body.Close()
					}

					stats.Record(status, err, latency)
				}
			}
		}(i)
	}

	b.Logf("Running test for %v with %d clients...", testDuration, numClients)
	time.Sleep(testDuration)
	close(done)
	wg.Wait()

	stats.PrintReport(b)
}
