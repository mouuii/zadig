package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go-sample-site/pkg/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	signals := make(chan os.Signal, 1)
	defer close(signals)

	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		select {
		case <-signals:
			signal.Stop(signals)
			cancel()
		}
	}()

	if err := server.Serve(ctx); err != nil {
		log.Fatal(err)
	}
}
