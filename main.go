package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var serverConfig *ServerConfig

func init() {
	serverConfig = new(ServerConfig)
	serverConfig.Version = "0.0.1.0"
	serverConfig.LogFileName = "q50w.log"
	serverConfig.Host = "127.0.0.1"
	serverConfig.Port = "8080"

	flag.StringVar(&serverConfig.Host, "host", "127.0.0.1", "-host 127.0.0.1")
	flag.StringVar(&serverConfig.Port, "port", "8080", "-port 80")
	flag.Parse()
}

func main() {
	f, err := os.OpenFile(serverConfig.LogFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		fmt.Printf("Error opening file: %v", err)
	}
	defer f.Close()

	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	s := NewHttpServer(serverConfig)

	go func() {
		log.Printf("Q50 web server v%s started on address: %v\n", serverConfig.Version, serverConfig.Addr())
		log.Fatal(s.ListenAndServe())
	}()

	waitingShutdown(s, 10*time.Second)
}

func waitingShutdown(s *http.Server, timeout time.Duration) {
	defer func() {
		log.Println("Q50W server stopped")
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	log.Println("Shutdown...")

	if err := s.Shutdown(ctx); err != nil {
		log.Printf("Error: %v\n", err)
	}
}
