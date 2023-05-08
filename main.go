package main

import (
	"github.com/wojtekolesinski/battleships/app"
	"github.com/wojtekolesinski/battleships/client"
	"golang.org/x/exp/slog"
	"os"
	"time"
)

const (
	serverAddress     = "https://go-pjatk-server.fly.dev/api"
	httpClientTimeout = 30 * time.Second
)

func main() {
	w, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		panic(err)
	}
	defer w.Close()
	slog.SetDefault(slog.New(slog.NewTextHandler(w)))

	c := client.NewClient(serverAddress, httpClientTimeout)
	a := app.New(c)

	a.Run()
	slog.Info("ENDING GAME")
}
