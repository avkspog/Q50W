package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	settings, err := LoadConfig()
	if err != nil {
		fmt.Printf("Load config file error: %s. Settings will be applied by default.\n", err.Error())
	}

	f, err := os.OpenFile(settings.LogFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		fmt.Printf("Error opening file: %v", err)
	}
	defer f.Close()

	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	s := NewHttpServer(settings)

	go func() {
		log.Printf("Q50 web server v%s started on address: %v\n", settings.Version, settings.Addr())
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
