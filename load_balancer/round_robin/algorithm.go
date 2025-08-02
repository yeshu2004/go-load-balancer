package main

import (
	"net/url"
	"sync"
)

type LoadBalancer struct {
	count int
	mu    sync.Mutex
}

type Server struct {
	URL       *url.URL
	isHealthy bool
	mu        sync.Mutex
}


func NewServer(url *url.URL, isHealthy bool) *Server {
	return &Server{URL: url, isHealthy: isHealthy}
}

func NewLoadBalancer(count int) *LoadBalancer {
    return &LoadBalancer{count: count}
}


func (lb *LoadBalancer) GetNextServerByRoundRobin(servers []*Server) *Server {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for i := 0; i < len(servers); i++ {
		index := lb.count % len(servers)
		nextServer := servers[index] // server selected
		lb.count++

		// check if server is healthy or not!
		nextServer.mu.Lock()
		isHealthy := nextServer.isHealthy
		if isHealthy {
			return nextServer
		}
		defer nextServer.mu.Unlock()

	}
	return nil
}