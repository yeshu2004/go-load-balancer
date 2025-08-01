package main

import (
	"log"
	"net/http"
	"net/url"
	"go-load-balancer/load_balancer"
	"time"
)

type Config struct{
	port string
	healthCheckInterval string
	servers_url []string
}


func main(){
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

	var servers []*load_balancer.Server
	interval, err := time.ParseDuration(config.healthCheckInterval)
	if err != nil {
		log.Printf("Error parsing health check interval: %v", err)
	}
	for _, serverUrl := range config.servers_url{
		parsedUrl, err := url.Parse(serverUrl)
		if err != nil {
			log.Printf("Error parsing server url: %v", err)
		}
		server := load_balancer.NewServer(parsedUrl, true); // server -> urls, heathCheck , mu
		servers = append(servers, server);

		go load_balancer.HeathCheck(server, interval)
	}
	
	lb := load_balancer.NewLoadBalancer(0);
	go load_balancer.StartHealthChecks(servers, interval)

	http.HandleFunc("/", func(rw http.ResponseWriter,r *http.Request){
		forwardedServer := lb.GetNextServerByRoundRobin(servers); 
		if forwardedServer == nil {
			http.Error(rw, "Healthy server not available", http.StatusServiceUnavailable)
			return
		}

		forwardedServer.ReverseProxy().ServeHTTP(rw, r)
		rw.Header().Add("X-Forwarded-Server", forwardedServer.URL.String());
	})


	log.Printf("load balancer server starting on port %v", config.port)
	if err := http.ListenAndServe(config.port, nil); err != nil{
		log.Fatalf("Error starting load balancer: %s\n", err.Error())
	}

}