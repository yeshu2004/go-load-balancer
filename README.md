# Go Load Balancer

## Overview
This project implements a simple load balancer in Go that distributes incoming HTTP requests across multiple backend servers using a round-robin algorithm. It includes health checking to ensure requests are only forwarded to healthy servers.

## Features
- **Round-Robin Load Balancing**: Distributes requests evenly across available servers.
- **Health Checks**: Periodically checks the health of backend servers and excludes unhealthy ones.
- **Reverse Proxy**: Forwards requests to backend servers using Go's `httputil.ReverseProxy`.
- **Configurable**: Allows configuration of port, health check interval, and backend server URLs.

## Project Structure
- `main.go`: Entry point for the load balancer, initializes configuration, and starts the HTTP server.
- `load_balancer/heath_check.go`: Implements health checking logic for backend servers.
- `load_balancer/reverse_proxy.go`: Configures the reverse proxy for forwarding requests.
- `load_balancer/round_robin.go`: Implements the round-robin algorithm and server management.

## Prerequisites
- Go 1.16 or higher
- Basic understanding of HTTP servers and load balancing

## Setup
1. **Clone the Repository**:
   ```bash
   git clone https://github.com/yeshu2004/go-load-balancer
   cd go-load-balancer
   ```

2. **Install Dependencies**:
   No external dependencies are required beyond the Go standard library.

3. **Run Mock Servers**:
   Start mock servers to simulate backend services. Open separate terminals and run the all mock server i.e `http://localhost:8001`, `http://localhost:8002`, `http://localhost:8003`, `http://localhost:80024`

4. **Run the Load Balancer**:
   ```bash
   go run main.go
   ```
   The load balancer starts on `:4000` and forwards requests to the configured backend servers (`http://localhost:8001`, `http://localhost:8002`, etc.).

## Configuration
The load balancer is configured in `main.go` via the `Config` struct:
- `port`: The port on which the load balancer listens (default: `:4000`).
- `healthCheckInterval`: Interval for health checks (default: `2s`).
- `servers_url`: List of backend server URLs (default: `http://localhost:8001`, `http://localhost:8002`, `http://localhost:8003`, `http://localhost:8004`).

Example configuration:
```go
config := &Config{
    port: ":4000",
    healthCheckInterval: "2s",
    servers_url: []string{
        "http://localhost:8001",
        "http://localhost:8002",
        "http://localhost:8003",
        "http://localhost:8004",
    },
}
```

## Usage
1. Start the mock servers on their respective ports.
2. Start the load balancer.
3. Send HTTP requests to `http://localhost:4000`. The load balancer will distribute requests to healthy backend servers.
   ```bash
   curl http://localhost:4000
   ```
   The response will indicate which backend server handled the request (e.g., `Response from server :8001`).

4. To test health checks, stop one of the mock servers (e.g., on `:8002`). The load balancer will detect it as unhealthy and stop forwarding requests to it.

## Health Checks
- Health checks are performed by sending a `GET` request to the `/health` endpoint of each server (if available) or a `HEAD` request to the root.
- A server is considered healthy if it responds with `200 OK`.
- Health checks run at the configured interval (default: every 2 seconds).

## Notes
- The load balancer skips unhealthy servers and returns a `503 Service Unavailable` if no healthy servers are available.
- The `X-Forwarded-Server` header is added to responses to indicate which backend server handled the request.
- The current implementation assumes mock servers do not have a `/health` endpoint, so health checks use `HEAD` requests. Modify `heath_check.go` if your servers provide a `/health` endpoint.

## Future Improvements
- Add support for weighted round-robin or other load balancing algorithms.
