# Load Balancer and API Server Project

This repository contains Go-based round-robin and weighted round-robin load balancers that distribute HTTP requests across backend API servers (`server_one` to `server_four`). The **round-robin load balancer** features **sticky session** support using Redis to ensure requests from the same client are routed to the same backend server.

## Project Structure

- `load_balancer/round_robin/`
  - `algorithm.go`: Implements round-robin logic with `LoadBalancer` and `Server` structs.
  - `health_check.go`: Manages health checks for backend servers.
  - `reverse_proxy.go`: Handles reverse proxy routing with sticky session logic.
  - `main.go`: Core application logic, including HTTP handling and Redis-based session persistence.
  - `go.sum`: Dependency checksums.

- `load_balancer/weighted_round_robin/`
  - `algorithm.go`: Implements weighted round-robin logic with `WeightedServer` and `LoadBalancer` structs.
  - `health_check.go`: Manages health checks for weighted servers.
  - `reverse_proxy.go`: Handles reverse proxy routing.
  - `main.go`: Core application logic for weighted distribution.

- `api_server/`
  - `server_one/`: Backend server on port 8001.
  - `server_two/`: Backend server on port 8002.
  - `server_three/`: Backend server on port 8003.
  - `server_four/`: Backend server on port 8004.
  - Each server includes a `/health` endpoint returning HTTP 200 OK when healthy.

## Features

- **Sticky Sessions (Round-Robin Only)**:
  - Uses Redis to store session IDs, ensuring requests with the same `session_id` cookie are routed to the same backend server.
  - New sessions are assigned a unique `session_id` (via `github.com/google/uuid`) and stored in Redis.

- **Round-Robin Load Balancer**:
  - Distributes new requests cyclically across healthy servers.
  - Maintains session persistence using Redis lookups.

- **Weighted Round-Robin Load Balancer**:
  - Distributes requests based on server weights (e.g., 3:1:2:1 for servers 8001, 8002, 8003, 8004).
  - No sticky session support (requests are distributed purely by weight).

- **Health Checks**: Periodically verifies server health, skipping unhealthy servers.
- **Fault Tolerance**: Handles Redis errors (round-robin) and server unavailability gracefully.
- **API Servers**: Backend servers with health endpoints for reliability.

## Prerequisites

- Go (version 1.18 or higher recommended)
- Redis server running on `localhost:6379` (no password, DB 0) **for round-robin only**
- Backend servers running on ports 8001, 8002, 8003, and 8004 with a `/health` endpoint.

## Setup Instructions

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/yeshu2004/go-load-balancer
   ```

2. **Install Dependencies**:
   Ensure `go.mod` includes required dependencies, or run (for round-robin):
   ```bash
   go get github.com/google/uuid
   go get github.com/redis/go-redis/v9
   ```

3. **Start Redis (for Round-Robin)**:
   - Run a Redis server locally:
     ```bash
     redis-server
     ```
   - Ensure it’s accessible at `localhost:6379` with persistence enabled (e.g., `appendonly yes` in `redis.conf`).

4. **Run Backend API Servers**:
   - Start each server in `api_server/` (e.g., `server_one`, `server_two`, etc.) on ports 8001–8004.
   - Ensure each server responds to `/health` with HTTP 200 OK when healthy. (To be done)

5. **Run the Round-Robin Load Balancer**:
   ```bash
   cd load_balancer/round_robin
   go run main.go
   ```
   - Access at `http://localhost:4000`.

6. **Run the Weighted Round-Robin Load Balancer**:
   ```bash
   cd load_balancer/weighted_round_robin
   go run main.go
   ```
   - Access at `http://localhost:4000`.

## Sticky Session Usage (Round-Robin Only)

- **New Requests**: Send an HTTP request without a `session_id` cookie. The round-robin load balancer:
  - Generates a unique `session_id` using `github.com/google/uuid`.
  - Stores the `session_id` and selected server in Redis.
  - Sets a `session_id` cookie in the response (expires in 1 hour, `MaxAge: 3600`).
  - Routes the request to a server based on round-robin logic.

- **Existing Sessions**: Send a request with a `session_id` cookie. The load balancer:
  - Queries Redis for the `session_id` to retrieve the assigned server.
  - Routes the request to the same server, ensuring session persistence.
  - Skips unhealthy servers and reassigns sessions if needed.

Example with `curl` (for round-robin):
```bash
curl -v http://localhost:4000
```
- Check the `Set-Cookie` header for the `session_id` and `X-Forwarded-Server` header for the backend server.
- Reuse the `session_id` cookie to verify routing to the same server:
  ```bash
  curl -v --cookie "session_id=<your-session-id>" http://localhost:4000
  ```

- **Weighted Round-Robin**: No sticky sessions; requests are distributed based on weights without session persistence.

## Configuration

- **Port**: Default `:4000` for both load balancers (edit `Config.port` in `main.go`).
- **Health Check Interval**: Default `2s` (edit `Config.healthCheckInterval` in `main.go`).
- **Backend Servers**: Defined in `Config.servers_url` in `main.go`.
- **Weights**: Defined in `Config.weight` for weighted round-robin (e.g., `[3, 1, 2, 1]`).
- **Redis (Round-Robin Only)**: Configured for `localhost:6379`, DB 0, no password (edit in `main.go` if needed).

## Notes

- Sticky sessions are exclusive to the round-robin load balancer, relying on Redis for persistence.
- Session cookies expire after 1 hour (`MaxAge: 3600`); adjust in `main.go` if needed.
- Health checks (every 2s) ensure only healthy servers receive requests.
- If a server becomes unhealthy, the round-robin load balancer skips it and may reassign sessions.
- Redis operations (round-robin) use a 2-second timeout for reliability.

## Troubleshooting

- **Session Not Sticking (Round-Robin)**: Verify Redis is running and check for `Failed to store session in Redis` logs. Ensure `appendonly yes` in `redis.conf`.
- **Routing Issues**: Confirm all servers are healthy (`/health` returns HTTP 200 OK).
- **Redis Errors (Round-Robin)**: Check connectivity to `localhost:6379` and ensure no password is set.
- **Stuck Requests**: Inspect for mutex deadlocks or Redis latency in logs (round-robin).

## Acknowledgments

- Sticky sessions in round-robin powered by Redis (`github.com/redis/go-redis/v9`) and UUID generation (`github.com/google/uuid`).
- Built with Go for robust load balancing and session persistence.