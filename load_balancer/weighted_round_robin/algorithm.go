package main

import (
	"net/url"
	"sync"
)

type WeightedServer struct {
	URL             *url.URL
	isHealthy       bool
	weight          int
	remainingWeight int
	mu              sync.Mutex
}

type LoadBalancer struct {
	count int
	mu    sync.Mutex
}

func NewLoadBalancer(count int) *LoadBalancer{
	return &LoadBalancer{
		count: count,
	}
}

func NewWeightedServer(url *url.URL, isHealthy bool, weight int) *WeightedServer{
	return &WeightedServer{
		URL: url,
		isHealthy: isHealthy,
		weight: weight,
		remainingWeight: weight,
	}
}

func (lb *LoadBalancer) GetNextServerByWeightedRoundRobin(servers []*WeightedServer) *WeightedServer{
	lb.mu.Lock()
	defer lb.mu.Unlock()
	for{
		for i := 0; i<len(servers); i++{
			index := lb.count % len(servers);
			 nextServer := servers[index];
			 lb.count++

			 nextServer.mu.Lock()
			 if nextServer.isHealthy && nextServer.remainingWeight >0{
				nextServer.remainingWeight--;
				nextServer.mu.Unlock()
				return nextServer
			 }

			 nextServer.mu.Unlock()
		}

		healthy := false;
		for _, server := range servers{
			server.mu.Lock()
			if server.isHealthy{
				server.remainingWeight = server.weight;
				healthy = true;
			}
			server.mu.Unlock()
		} 

		if healthy{
			return lb.GetNextServerByWeightedRoundRobin(servers)
		}else{
			return nil;
		}
	}
}