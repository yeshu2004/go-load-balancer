package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)


func StartHealthChecks(servers []*WeightedServer, interval time.Duration) {
    ticker := time.NewTicker(interval)
    go func() {
        for range ticker.C {
            for _, server := range servers {
                checkServerHealth(server)
                server.mu.Lock()
                log.Printf("Checked %s: %v", server.URL.String(), server.isHealthy)
                server.mu.Unlock()
            }
        }
    }()
}

func checkServerHealth(server *WeightedServer){
	client := &http.Client{Timeout: 2 * time.Second}
    resp, err := client.Get(server.URL.String() + "/health")
    server.mu.Lock()
    defer server.mu.Unlock()
    server.isHealthy = err == nil && resp.StatusCode == http.StatusOK
}

func HealthCheck(server *WeightedServer, healthCheckInterval time.Duration){
	for range time.Tick(healthCheckInterval){
		res , err := http.Head(server.URL.String())

		server.mu.Lock()
		if err != nil  || res.StatusCode != http.StatusOK {
			fmt.Printf("%s is down\n", server.URL)
			server.isHealthy = false;
		}else{
			server.isHealthy = true;
		}

		server.mu.Unlock()

	}
}