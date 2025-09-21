# Autocannon - HTTP Benchmarking Tool

A fast, lightweight HTTP benchmarking tool written in Go that helps you measure the performance of your web services and APIs.

## Features

- 🚀 **High Performance**: Concurrent HTTP requests using goroutines
- 📊 **Detailed Statistics**: Comprehensive metrics including latency, throughput, and error rates
- 🎯 **Flexible Configuration**: Customizable connections, duration, timeouts, and HTTP methods
- 📈 **Beautiful Output**: Color-coded console output with formatted tables
- 💾 **JSON Export**: Save results to JSON files for further analysis
- 🔍 **Status Code Tracking**: Detailed breakdown of HTTP response codes
- ⏱️ **Latency Metrics**: Min, max, and average response times
- 🛠️ **Debug Mode**: Verbose logging for troubleshooting

## Installation

### From Source

```bash
git clone <repository-url>
cd autocannon
go build -o autocannon main.go
```

### Using Go Install

```bash
go install github.com/your-username/autocannon@latest
```

## Usage

### Basic Usage

```bash
# Simple GET request benchmark
./autocannon -uri http://localhost:3000

# Benchmark with custom parameters
./autocannon -uri http://localhost:3000 -clients 50 -duration 30
```

### Command Line Options

| Flag | Default | Description |
|------|---------|-------------|
| `-uri` | *required* | The URI to benchmark against |
| `-clients` | 10 | Number of concurrent connections |
| `-duration` | 10 | Duration of the test in seconds |
| `-timeout` | 10 | Request timeout in seconds |
| `-method` | GET | HTTP method to use |
| `-body` | "" | Request body to send |
| `-expect` | 200 | Expected HTTP status code |
| `-output` | "" | Output file for JSON results |
| `-debug` | false | Enable debug logging |

### Examples

#### Basic GET Request
```bash
./autocannon -uri https://httpbin.org/get
```

#### High Load Test
```bash
./autocannon -uri http://localhost:8080/api/health -clients 100 -duration 60
```

#### POST Request with Body
```bash
./autocannon -uri http://localhost:8080/api/users -method POST -body '{"name":"test"}'
```

#### Save Results to File
```bash
./autocannon -uri http://localhost:3000 -output results.json
```

#### Debug Mode
```bash
./autocannon -uri http://localhost:3000 -debug
```

## Output

The tool provides two main types of output:

### Console Output

```
Starting autocannon with the following parameters:
URI: http://localhost:3000
Connections: 10
Duration: 10 seconds
Timeout: 10 seconds
Method: GET
Expected status: 200
Debug: false
Starting autocannon...

Benchmark Results:
┌─────────────────────┬────────────┐
│       METRIC        │   VALUE    │
├─────────────────────┼────────────┤
│ Total Requests      │      15420 │
│ Successful Requests │      15420 │
│ Failed Requests     │          0 │
│ Timeouts           │          0 │
│ Requests/sec       │    1542.00 │
│ Average Latency    │       6.48 │
│ Min Latency        │       1.23 │
│ Max Latency        │      45.67 │
│ Total Data Received │    1234567 │
│ Error Rate         │       0.00 │
└─────────────────────┴────────────┘

Status Code Distribution:
┌─────────────┬───────┬────────────┐
│ STATUS CODE │ COUNT │ PERCENTAGE │
├─────────────┼───────┼────────────┤
│     200     │ 15420 │   100.00%  │
└─────────────┴───────┴────────────┘
```

### JSON Output

When using the `-output` flag, results are saved in JSON format:

```json
{
  "connections": 10,
  "durationSeconds": 10,
  "totalRequests": 15420,
  "successfulRequests": 15420,
  "failedRequests": 0,
  "timeouts": 0,
  "requestsPerSecond": 1542.00,
  "averageLatencyMs": 6.48,
  "minLatencyMs": 1.23,
  "maxLatencyMs": 45.67,
  "bytesRead": 1234567,
  "bytesWritten": 308400,
  "errorRate": 0.00,
  "statusCodes": {
    "200": 15420
  },
  "timestamp": "2025-09-21T10:30:00Z"
}
```

## Metrics Explained

- **Total Requests**: Total number of HTTP requests sent
- **Successful Requests**: Number of requests that completed without errors
- **Failed Requests**: Number of requests that failed (network errors, etc.)
- **Timeouts**: Number of requests that exceeded the timeout duration
- **Requests/sec**: Average throughput (requests per second)
- **Average Latency**: Mean response time in milliseconds
- **Min/Max Latency**: Fastest and slowest response times
- **Total Data Received**: Total bytes received from the server
- **Error Rate**: Percentage of failed requests
- **Status Code Distribution**: Breakdown of HTTP response codes

## Use Cases

- **API Performance Testing**: Measure the performance of REST APIs
- **Load Testing**: Determine how many concurrent users your service can handle
- **Regression Testing**: Compare performance before and after code changes
- **Infrastructure Testing**: Test the limits of your server infrastructure
- **CI/CD Integration**: Automated performance testing in deployment pipelines

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Dependencies

- [tablewriter](https://github.com/olekukonko/tablewriter) - For formatted console output
- [chalk](https://github.com/ttacon/chalk) - For colored terminal output

## Acknowledgments

- Inspired by the Node.js [autocannon](https://github.com/mcollina/autocannon) tool
- Built with the power and simplicity of Go's standard library