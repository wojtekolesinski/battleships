package app

import (
	"context"
	"fmt"
	gui "github.com/grupawp/warships-gui"
	"github.com/wojtekolesinski/battleships/client"
	"log"
	"strconv"
	"time"
)

type App struct {
	client        *client.Client
	playerBoard   [10][10]gui.State
	opponentBoard [10][10]gui.State
	state         client.StatusData
}

func New(c *client.Client) *App {
	return &App{
		client: c,
	}
}

func (a *App) Run() error {
	err := a.client.InitGame("testtt", "test", "", true)
	if err != nil {
		return err
	}

	status, err := a.client.GetStatus()
	if err != nil {
		return err
	}
	for status.GameStatus == "waiting_wpbot" {
		time.Sleep(time.Second)
		status, err = a.client.GetStatus()
		if err != nil {
			return err
		}
	}
	board, err := a.client.GetBoard()
	a.parseBoard(board)
	if err != nil {
		return err
	}
	a.Render(board)

	return nil
}

func parseCoords(coords string) (int, int, error) {
	x := int(coords[0] - 'A')
	y, err := strconv.Atoi(coords[1:])
	y -= 1
	if err != nil {
		return -1, -1, err
	}
	return x, y, nil
}

func (a *App) parseBoard(b client.Board) error {
	a.playerBoard = [10][10]gui.State{}
	a.opponentBoard = [10][10]gui.State{}
	for i := range a.playerBoard {
		a.playerBoard[i] = [10]gui.State{}
		a.opponentBoard[i] = [10]gui.State{}
	}

	for _, coords := range b.Board {
		x, y, err := parseCoords(coords)
		if err != nil {
			return err
		}
		a.playerBoard[x][y] = gui.Ship
	}
	return nil
}

func (a *App) Render(bo client.Board) {
	ctx := context.TODO()

	d := gui.NewDrawer(&gui.Config{})
	b, err := d.NewBoard(2, 4, &gui.BoardConfig{})
	b2, _ := d.NewBoard(50, 4, &gui.BoardConfig{})
	if err != nil {
		log.Fatal(err)
	}
	defer d.RemoveBoard(ctx, b)

	//coords := d.DrawBoardAndCatchCoords(ctx, b, states) // draw empty board at position (2,4)
	d.DrawBoard(ctx, b, a.playerBoard) // draw empty board at position (2,4)

	d.DrawBoard(ctx, b2, a.opponentBoard) // draw empty board at position (2,4)

	t, err := d.NewText(2, 2, nil) // initialize some text object
	if err != nil {
		log.Fatal(err)
	}
	t.SetText(fmt.Sprintf("You clicked: %v ", bo.Board))
	d.DrawText(ctx, t)

	for {
		if !d.IsGameRunning() { // wait until escape character has been pressed
			return
		}
	}

	//client.StartGame()
}
