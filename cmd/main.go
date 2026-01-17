package main

import (
	"context"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Miklakapi/go-mqtt-tester/internal/config"
	"github.com/Miklakapi/go-mqtt-tester/internal/mqtt"
	"github.com/Miklakapi/go-mqtt-tester/internal/watcher"

	MQTT "github.com/eclipse/paho.mqtt.golang"
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

	o := MQTT.NewClientOptions()
	brokerURL := "tcp://" + conf.MqttBrokerHost + ":" + conf.MqttBrokerPort
	o.AddBroker(brokerURL)

	o.SetClientID(conf.MqttClientId)
	o.SetUsername(conf.MqttUsername)
	o.SetPassword(conf.MqttPassword)

	o.SetCleanSession(true)
	o.SetAutoReconnect(true)
	o.SetConnectRetry(true)

	o.SetConnectTimeout(10 * time.Second)
	o.SetKeepAlive(30 * time.Second)
	o.SetPingTimeout(10 * time.Second)

	mq, err := mqtt.New(o)
	if err != nil {
		log.Fatal(err)
	}
	defer mq.Close()

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
