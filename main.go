package main

import (
	"github.com/wojtekolesinski/battleships/app"
	"github.com/wojtekolesinski/battleships/client"
	"time"
)

const (
	serverAddress     = "https://go-pjatk-server.fly.dev/api"
	httpClientTimeout = 30 * time.Second
)

func main() {
	c := client.NewClient(serverAddress, httpClientTimeout)
	a := app.New(c)

	a.Run()
}
