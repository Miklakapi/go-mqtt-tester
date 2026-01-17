package main

import (
	"context"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Miklakapi/go-mqtt-tester/internal/config"
	"github.com/Miklakapi/go-mqtt-tester/internal/watcher"
)

func main() {
	appCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logFile, err := os.OpenFile("logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.Fatalf("cannot open log file: %v", err)
	}
	defer logFile.Close()

	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	conf, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	w, err := watcher.New()
	if err != nil {
		log.Fatal(err)
	}
	defer w.Close()
	w.AddWatch(conf.DataDir, syscall.IN_DELETE|syscall.IN_MOVED_FROM|syscall.IN_MOVED_TO|syscall.IN_CLOSE_WRITE)

loop:
	for {
		select {
		case event := <-w.Events:
			if event.Mask&syscall.IN_ISDIR != 0 {
				continue
			}
			log.Println(event)
		case err := <-w.Errors:
			log.Println(err)
		case <-appCtx.Done():
			break loop
		}
	}

	<-appCtx.Done()
	log.Println("shutdown signal received")
}
