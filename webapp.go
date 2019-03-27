package main

import (
	"net/http"
	"time"
)

func NewHttpServer(config *ServerConfig) *http.Server {
	handler := http.NewServeMux()
	s := &http.Server{
		Addr:           config.Addr(),
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    15 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return s
}
