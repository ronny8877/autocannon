package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/ttacon/chalk"
)

// BenchmarkConfig holds all configuration options for the benchmark
type BenchmarkConfig struct {
	URI              string
	Connections      int
	Duration         int
	Timeout          int
	Method           string
	Headers          map[string]string
	Body             string
	ExpectStatusCode int
	Debug            bool
	OutputFile       string
}

// BenchmarkResult holds the results of the benchmark
type BenchmarkResult struct {
	Connections      int           `json:"connections"`
	Duration         int           `json:"durationSeconds"`
	TotalRequests    int64         `json:"totalRequests"`
	SuccessfulReqs   int64         `json:"successfulRequests"`
	FailedReqs       int64         `json:"failedRequests"`
	Timeouts         int64         `json:"timeouts"`
	RequestsPerSec   float64       `json:"requestsPerSecond"`
	AverageLatency   float64       `json:"averageLatencyMs"`
	MinLatency       float64       `json:"minLatencyMs"`
	MaxLatency       float64       `json:"maxLatencyMs"`
	BytesRead        int64         `json:"bytesRead"`
	BytesWritten     int64         `json:"bytesWritten"`
	ErrorRate        float64       `json:"errorRate"`
	StatusCodeCounts map[int]int64 `json:"statusCodes"`
	Timestamp        time.Time     `json:"timestamp"`
}

func main() {
	// Parse command-line arguments
	uri := flag.String("uri", "", "The uri to benchmark against. (Required)")
	clients := flag.Int("clients", 10, "The number of connections to open to the server.")
	runtime := flag.Int("duration", 10, "The number of seconds to run the autocannnon.")
	timeout := flag.Int("timeout", 10, "The number of seconds before timing out on a request.")
	method := flag.String("method", "GET", "HTTP method to use")
	body := flag.String("body", "", "Request body to send")
	expectStatus := flag.Int("expect", 200, "Expected status code")
	output := flag.String("output", "", "Output file to write results as JSON")
	debug := flag.Bool("debug", false, "A utility debug flag.")
	flag.Parse()

	if *uri == "" {
		fmt.Println("You must provide a uri to benchmark against.")
		flag.Usage()
		os.Exit(1)
	}

	// Print parameters
	fmt.Print(chalk.Green, "Starting autocannon with the following parameters:\n", chalk.Reset)
	fmt.Printf("URI: %s\n", *uri)
	fmt.Printf("Connections: %d\n", *clients)
	fmt.Printf("Duration: %d seconds\n", *runtime)
	fmt.Printf("Timeout: %d seconds\n", *timeout)
	fmt.Printf("Method: %s\n", *method)
	fmt.Printf("Expected status: %d\n", *expectStatus)
	if *output != "" {
		fmt.Printf("Output file: %s\n", *output)
	}
	fmt.Printf("Debug: %t\n", *debug)
	fmt.Println(chalk.Green, "Starting autocannon...", chalk.Reset)

	// Configure the benchmark
	config := BenchmarkConfig{
		URI:              *uri,
		Connections:      *clients,
		Duration:         *runtime,
		Timeout:          *timeout,
		Method:           *method,
		Headers:          map[string]string{},
		Body:             *body,
		ExpectStatusCode: *expectStatus,
		Debug:            *debug,
		OutputFile:       *output,
	}

	// Run the benchmark
	result := runBenchmark(config)

	// Display results
	displayResults(result)

	// Write results to file if specified
	if config.OutputFile != "" {
		writeResultsToFile(result, config.OutputFile)
	}
}

func runBenchmark(config BenchmarkConfig) BenchmarkResult {
	result := BenchmarkResult{
		Connections:      config.Connections,
		Duration:         config.Duration,
		StatusCodeCounts: make(map[int]int64),
		Timestamp:        time.Now(),
	}

	var wg sync.WaitGroup
	var totalRequests int64
	var successfulReqs int64
	var failedReqs int64
	var timeouts int64
	var bytesRead int64
	var bytesWritten int64
	var statusCodeMutex sync.Mutex
	// For latency tracking
	var totalLatency float64
	var minLatency float64 = float64(^uint64(0) >> 1) // Max float64 value
	var maxLatency float64

	// Channel to collect latency measurements
	latencyChan := make(chan float64, 1000)

	// Create a client with specified timeout
	client := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}

	// Create a stop channel that will signal workers to stop
	stopChan := make(chan struct{})

	// Launch worker goroutines
	for i := 0; i < config.Connections; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for {
				select {
				case <-stopChan:
					return
				default:
					startTime := time.Now()

					// Create request
					req, err := http.NewRequest(config.Method, config.URI, nil)
					if err != nil {
						atomic.AddInt64(&failedReqs, 1)
						if config.Debug {
							fmt.Printf("Error creating request: %v\n", err)
						}
						continue
					}

					// Add headers
					for key, value := range config.Headers {
						req.Header.Add(key, value)
					}

					// Send request and measure time
					resp, err := client.Do(req)
					latency := float64(time.Since(startTime).Milliseconds())

					// Send latency to channel for stats
					latencyChan <- latency

					// Increment request counter
					atomic.AddInt64(&totalRequests, 1)

					// Handle response or error
					if err != nil {
						atomic.AddInt64(&failedReqs, 1)
						if config.Debug {
							fmt.Printf("Request error: %v\n", err)
						}
						// Check if it's a timeout
						if os.IsTimeout(err) {
							atomic.AddInt64(&timeouts, 1)
						}
					} else {
						atomic.AddInt64(&successfulReqs, 1)

						// Use mutex to protect map update
						statusCodeMutex.Lock()
						result.StatusCodeCounts[resp.StatusCode]++
						statusCodeMutex.Unlock()

						// Read and discard body (important to close connections properly)
						body, _ := io.ReadAll(resp.Body)
						atomic.AddInt64(&bytesRead, int64(len(body)))
						atomic.AddInt64(&bytesWritten, int64(req.ContentLength))

						resp.Body.Close()
					}
				}
			}
		}(i)
	}

	// Start latency collector goroutine
	latencyDone := make(chan struct{})
	go func() {
		count := 0
		for latency := range latencyChan {
			count++
			totalLatency += latency

			if latency < minLatency {
				minLatency = latency
			}
			if latency > maxLatency {
				maxLatency = latency
			}
		}
		close(latencyDone)
	}()

	// Run for specified duration
	time.Sleep(time.Duration(config.Duration) * time.Second)

	// Signal workers to stop
	close(stopChan)

	// Wait for all workers to finish
	wg.Wait()

	close(latencyChan)
	<-latencyDone
	result.TotalRequests = totalRequests
	result.SuccessfulReqs = successfulReqs
	result.FailedReqs = failedReqs
	result.Timeouts = timeouts
	result.BytesRead = bytesRead
	result.BytesWritten = bytesWritten

	if totalRequests > 0 {
		result.RequestsPerSec = float64(totalRequests) / float64(config.Duration)
		result.ErrorRate = float64(failedReqs) / float64(totalRequests) * 100
	}

	if successfulReqs > 0 {
		result.AverageLatency = totalLatency / float64(successfulReqs)
		result.MinLatency = minLatency
		result.MaxLatency = maxLatency
	}

	return result
}
func displayResults(result BenchmarkResult) {
	fmt.Println(chalk.Green, "\nBenchmark Results:", chalk.Reset)

	// Main results table
	mainTable := tablewriter.NewTable(os.Stdout,
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Formatting: tw.CellFormatting{
					Alignment: tw.AlignLeft,
				},
				ColumnAligns: []tw.Align{tw.AlignLeft, tw.AlignRight},
			},
			Header: tw.CellConfig{
				Formatting: tw.CellFormatting{
					Alignment: tw.AlignCenter,
				},
			},
		}),
	)

	mainTable.Header("Metric", "Value")

	mainTable.Append([]string{"Total Requests", fmt.Sprintf("%d", result.TotalRequests)})
	mainTable.Append([]string{"Successful Requests", fmt.Sprintf("%d", result.SuccessfulReqs)})
	mainTable.Append([]string{"Failed Requests", fmt.Sprintf("%d", result.FailedReqs)})
	mainTable.Append([]string{"Timeouts", fmt.Sprintf("%d", result.Timeouts)})
	mainTable.Append([]string{"Requests/sec", fmt.Sprintf("%.2f", result.RequestsPerSec)})
	mainTable.Append([]string{"Average Latency", fmt.Sprintf("%.2f ms", result.AverageLatency)})
	mainTable.Append([]string{"Min Latency", fmt.Sprintf("%.2f ms", result.MinLatency)})
	mainTable.Append([]string{"Max Latency", fmt.Sprintf("%.2f ms", result.MaxLatency)})
	mainTable.Append([]string{"Total Data Received", fmt.Sprintf("%d bytes", result.BytesRead)})
	mainTable.Append([]string{"Error Rate", fmt.Sprintf("%.2f%%", result.ErrorRate)})

	mainTable.Render()

	// Status code distribution table
	fmt.Println(chalk.Green, "\nStatus Code Distribution:", chalk.Reset)

	statusTable := tablewriter.NewTable(os.Stdout,
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Formatting: tw.CellFormatting{
					Alignment: tw.AlignLeft,
				},
				ColumnAligns: []tw.Align{tw.AlignCenter, tw.AlignRight, tw.AlignRight},
			},
			Header: tw.CellConfig{
				Formatting: tw.CellFormatting{
					Alignment: tw.AlignCenter,
				},
			},
		}),
	)

	statusTable.Header("Status Code", "Count", "Percentage")

	for code, count := range result.StatusCodeCounts {
		percentage := float64(count) / float64(result.TotalRequests) * 100
		statusTable.Append([]string{
			fmt.Sprintf("%d", code),
			fmt.Sprintf("%d", count),
			fmt.Sprintf("%.2f%%", percentage),
		})
	}

	statusTable.Render()
}

func writeResultsToFile(result BenchmarkResult, filename string) {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling results to JSON: %v\n", err)
		return
	}

	err = ioutil.WriteFile(filename, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing results to file: %v\n", err)
		return
	}

	fmt.Printf("Results written to %s\n", filename)
}
