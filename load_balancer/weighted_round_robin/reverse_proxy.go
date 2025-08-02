package main

import (
	"net/http/httputil"
)

func (s *WeightedServer) ReverseProxy() *httputil.ReverseProxy{
	return httputil.NewSingleHostReverseProxy(s.URL);
}