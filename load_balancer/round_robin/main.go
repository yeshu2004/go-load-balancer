package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"net/url"

	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
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
	// redis setup
	redisClient := redis.NewClient(&redis.Options{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           0,
		Protocol:     2,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		PoolSize:     10,
		MinIdleConns: 2,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v. Ensure Redis is running on localhost:6379.", err)
	}
	log.Println("Redis connection successful")
	
	var servers []*Server
	interval, err := time.ParseDuration(config.healthCheckInterval)
	if err != nil {
		log.Printf("Error parsing health check interval: %v", err)
	}
	for _, serverUrl := range config.servers_url{
		parsedUrl, err := url.Parse(serverUrl)
		if err != nil {
			log.Printf("Error parsing server url: %v", err)
		}
		server := NewServer(parsedUrl, true); // server -> urls, heathCheck , mu
		servers = append(servers, server);
		go HealthCheck(server, interval)
	}
	
	lb := NewLoadBalancer(0);
	go StartHealthChecks(servers, interval)

	http.HandleFunc("/", func(rw http.ResponseWriter,r *http.Request){
		ip, port, err := getUserIP(r)
		if err != nil {
				http.Error(rw, "Error getting IP: "+err.Error(), http.StatusBadRequest)
				return
		}
		log.Printf("Request from IP: %s, Port: %s", ip, port)


		//Check for session cookie
		cookie, err := r.Cookie("session_id")
		var sessionID  string
		if err == http.ErrNoCookie{
			//set cookie 
			sessionID = setCookieHandler(rw)
			log.Printf("New session set: %s", sessionID)
		} else if err != nil {
			http.Error(rw, "Error reading cookie: "+err.Error(), http.StatusInternalServerError)
			return
		} else {
			sessionID  = cookie.Value
			log.Printf("Existing session: %s", sessionID)
		}

		var forwardedServer *Server
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		val, err := redisClient.Get(ctx, sessionID).Result()
		if err == redis.Nil {
			// key does not exist
			forwardedServer = lb.GetNextServerByRoundRobin(servers); 
			if forwardedServer == nil {
				http.Error(rw, "Healthy server not available", http.StatusServiceUnavailable)
				return
			}

			err := redisClient.Set(ctx, sessionID, forwardedServer.URL.String(), 1*time.Hour).Err()
			if err != nil{
				log.Printf("Failed to store session in Redis: %v", err)
			}
			log.Printf("New session %s mapped to server %s", sessionID, forwardedServer.URL.String())
		} else if err != nil {
			// some other error
			log.Printf("Redis error: %v", err)
			forwardedServer = lb.GetNextServerByRoundRobin(servers)
			if forwardedServer == nil {
				http.Error(rw, "Healthy server not available", http.StatusServiceUnavailable)
				return
			}
			err = redisClient.Set(ctx, sessionID, forwardedServer.URL.String(), 1*time.Hour).Err()
			if err != nil {
				log.Printf("Failed to store session in Redis: %v", err)
			}
			log.Printf("New session %s mapped to server %s", sessionID, forwardedServer.URL.String())
		} else {
			for _, server := range servers{
				if server.URL.String() == val && server.IsHealthy() {
					forwardedServer = server
					break
				}
			}

			if forwardedServer == nil{
				forwardedServer = lb.GetNextServerByRoundRobin(servers); 
				if forwardedServer == nil {
					http.Error(rw, "Healthy server not available", http.StatusServiceUnavailable)
					return
				}

				err := redisClient.Set(ctx, sessionID, forwardedServer.URL.String(), 1*time.Hour).Err()
				if err != nil{
					log.Printf("Failed to store session in Redis: %v", err)
				}
				log.Printf("Session %s remapped to server %s", sessionID, forwardedServer.URL.String())
			} else {
				log.Printf("Session %s routed to mapped server %s", sessionID, val)
			}
		}

		rw.Header().Add("X-Forwarded-Server", forwardedServer.URL.String());
		forwardedServer.ReverseProxy().ServeHTTP(rw, r)
	})


	log.Printf("load balancer server starting on port %v", config.port)
	if err := http.ListenAndServe(config.port, nil); err != nil{
		log.Fatalf("Error starting load balancer: %s\n", err.Error())
	}

}


func getUserIP(r *http.Request) (string, string, error){
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		// Fall back to RemoteAddr if no proxy header
		ip = r.RemoteAddr
	}

	host, port, err := net.SplitHostPort(ip)
	if err != nil {
		return "", "", errors.New("invalid IP:port format")
	}

	userIP := net.ParseIP(host)
	if userIP == nil {
		return "", "", errors.New("invalid IP address")
	}

	return host, port, nil
}

func setCookieHandler(w http.ResponseWriter) string{
	id := uuid.New().String()
	cookie := http.Cookie{
		Name: "session_id",
		Value: id,
		Path: "/",
		MaxAge: 3600,
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)
	return id;
}