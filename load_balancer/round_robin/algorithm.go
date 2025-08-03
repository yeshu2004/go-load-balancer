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
	mu        sync.RWMutex
}


func NewServer(url *url.URL, isHealthy bool) *Server {
	return &Server{URL: url, isHealthy: isHealthy}
}

func NewLoadBalancer(count int) *LoadBalancer {
    return &LoadBalancer{count: count}
}


func (lb *LoadBalancer) GetNextServerByRoundRobin(servers []*Server) *Server {
	lb.mu.Lock()
    index := lb.count % len(servers)
    lb.count = (lb.count + 1) % len(servers) 
    candidateServer := servers[index]
    lb.mu.Unlock() 

    if candidateServer.IsHealthy() {
        return candidateServer
    }

    for i := 0; i < len(servers)-1; i++ {
        lb.mu.Lock()
        index = (index + 1) % len(servers)
        candidateServer = servers[index]
        lb.mu.Unlock()

        if candidateServer.IsHealthy() {
            return candidateServer
        }
    }
    return nil  // No healthy servers
}