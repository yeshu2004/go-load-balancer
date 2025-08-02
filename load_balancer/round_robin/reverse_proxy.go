package main

import (
	"net/http/httputil"
)

func (s *Server) ReverseProxy() *httputil.ReverseProxy{
	return httputil.NewSingleHostReverseProxy(s.URL);
}