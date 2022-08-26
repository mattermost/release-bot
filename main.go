package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/mattermost/release-bot/server"
	"github.com/mattermost/release-bot/version"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	version.Log()
	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := server.New()

	go func() {
		select {
		case <-signalChanel:
			log.Info("Received an interrupt, stopping...")
		case <-ctx.Done():
			log.Info("Context done, stopping...")
		}
		if err := srv.Stop(); err != nil {
			panic(err)
		}
	}()

	err := srv.Start(ctx)
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
	log.Info("Stopped!")
}
