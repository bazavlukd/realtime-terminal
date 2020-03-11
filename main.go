package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"

	"github.com/joho/godotenv"
	pusher "github.com/pusher/pusher-http-go"
)

const (
	channelName = "realtimeTerminal"
	eventName   = "cmdLog"
)

func main() {
	info, err := os.Stdin.Stat()
	if err != nil {
		log.Fatal(err)
	}

	if info.Mode()&os.ModeCharDevice != 0 {
		log.Println("This command is intended to be used as a pipe such as yourprogram | thisprogram")
		os.Exit(0)
	}

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	appID := os.Getenv("PUSHER_APP_ID")
	appKey := os.Getenv("PUSHER_APP_KEY")
	appSecret := os.Getenv("PUSHER_APP_SECRET")
	appCluster := os.Getenv("PUSHER_APP_CLUSTER")
	appIsSecure := os.Getenv("PUSHER_APP_SECURE")

	var isSecure bool

	if appIsSecure == "1" {
		isSecure = true
	}

	client := &pusher.Client{
		AppID:   appID,
		Key:     appKey,
		Secret:  appSecret,
		Cluster: appCluster,
		Secure:  isSecure,
	}

	reader := bufio.NewReader(os.Stdin)

	var writer io.Writer
	writer = pusherChannelWriter{client: client}

	for {
		in, _, err := reader.ReadLine()
		if err != nil && err == io.EOF {
			break
		}

		in = append(in, []byte("\n")...)
		if _, err := writer.Write(in); err != nil {
			log.Fatalln(err)
		}
	}
}

type pusherChannelWriter struct {
	client *pusher.Client
}

func (pusher pusherChannelWriter) Write(p []byte) (int, error) {
	s := string(p)
	dd := bytes.Split(p, []byte("\n"))

	var data = make([]string, 0, len(dd))

	for _, v := range dd {
		data = append(data, string(v))
	}

	err := pusher.client.Trigger(channelName, eventName, s)
	return len(p), err
}
