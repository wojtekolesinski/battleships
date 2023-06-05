package main

import (
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/wojtekolesinski/battleships/app"
	"github.com/wojtekolesinski/battleships/client"
	"os"
	"time"
)

const (
	serverAddress     = "https://go-pjatk-server.fly.dev/api"
	httpClientTimeout = 10 * time.Second
)

func main() {
	//var board app.Board
	//for i := range board {
	//	var row [10]gui.State
	//	for j := range row {
	//		row[j] = gui.Empty
	//	}
	//	board[i] = row
	//}
	//
	//state := app.NewGenState()
	//probs := app.GenerateBoards2(board, state)
	//for i := range probs {
	//	fmt.Println(probs[i])
	//}
	//os.Exit(0)

	logPath := fmt.Sprintf("log/%s.log", time.Now().Format("02-01-2006"))
	if len(os.Args) > 1 {
		logPath = os.Args[1]
	}
	w, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	log.SetOutput(w)
	log.SetLevel(log.DebugLevel)
	c := client.NewClient(serverAddress, httpClientTimeout)
	a := app.New(c)

	err = a.Run()
	if err != nil {
		log.Error("main [main]", "err", err)
		fmt.Println("Something went wrong")
	}
	log.Info("main [main] - ENDING GAME")
	fmt.Println("Thanks for playing")
}
