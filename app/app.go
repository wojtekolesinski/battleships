package app

import (
	"context"
	"fmt"
	gui "github.com/grupawp/warships-gui/v2"
	"github.com/mitchellh/go-wordwrap"
	"github.com/wojtekolesinski/battleships/client"
	"golang.org/x/exp/slog"
	"strconv"
	"strings"
	"time"
)

type App struct {
	client        *client.Client
	playerBoard   [10][10]gui.State
	opponentBoard [10][10]gui.State
	status        client.StatusData
	gui           struct {
		board1 *gui.Board
		board2 *gui.Board
		ui     *gui.GUI
		text   *gui.Text
	}
}

func New(c *client.Client) *App {
	return &App{
		client: c,
	}
}

func (a *App) updateStatus(data client.StatusData) {
	a.status.ShouldFire = data.ShouldFire
	a.status.GameStatus = data.GameStatus
	a.status.OppShots = data.OppShots
	a.status.LastGameStatus = data.LastGameStatus
	a.status.Timer = data.Timer
}

func (a *App) updateDescription(data client.StatusData) {
	a.status.Nick = data.Nick
	a.status.Desc = data.Desc
	a.status.Opponent = data.Opponent
	a.status.OppDesc = data.OppDesc
}

func (a *App) Run() error {
	err := a.client.InitGame("testtt", "Player description", "", true)
	if err != nil {
		return err
	}

	status, err := a.client.GetStatus()
	if err != nil {
		return err
	}

	for status.GameStatus != "game_in_progress" {
		time.Sleep(time.Second)
		status, err = a.client.GetStatus()
		if err != nil {
			return err
		}
		a.updateStatus(status)
	}

	status, err = a.client.GetDescription()
	if err != nil {
		return err
	}
	a.updateDescription(status)

	board, err := a.client.GetBoard()
	if err != nil {
		return err
	}

	err = a.parseBoard(board)
	if err != nil {
		return err
	}

	a.InitGUI()

	go func() {
		for a.status.GameStatus == "game_in_progress" {
			err = a.waitForYourTurn()
			if err != nil {
				return
			}
			slog.Info("app [Run]", slog.String("opp shots", strings.Join(a.status.OppShots, " ")), slog.Bool("shouldFire", a.status.ShouldFire))
			a.updateBoard()

			var answer client.FireAnswer

			for answer.Result != "miss" {
				coord := a.handleShot()

				answer, err = a.client.Fire(coord)
				if err != nil {
					return
				}

				x, y, _ := parseCoords(coord)
				switch answer.Result {
				case "hit":
					a.opponentBoard[x][y] = gui.Hit
					a.gui.text.SetText("HIT")
				case "miss":
					a.opponentBoard[x][y] = gui.Miss
					a.gui.text.SetText("MISS")
					break
				case "sunk":
					a.opponentBoard[x][y] = gui.Hit
					a.gui.text.SetText("SUNK")
				}
				slog.Info("app [Run] no break")
				a.updateBoard()
				time.Sleep(1 * time.Second)
			}

			status, err = a.client.GetStatus()
			a.updateStatus(status)

		}

		if a.status.LastGameStatus == "win" {
			a.gui.text.SetBgColor(gui.Green)
			a.gui.text.SetFgColor(gui.White)
			a.gui.text.SetText("You win")
		} else {
			a.gui.text.SetBgColor(gui.Red)
			a.gui.text.SetFgColor(gui.White)
			a.gui.text.SetText("You lose")
		}
	}()

	a.gui.ui.Start(nil)

	return nil
}

//func handleShot() {
//
//}

func (a *App) waitForYourTurn() error {
	for !a.status.ShouldFire {
		time.Sleep(time.Second)
		status, err := a.client.GetStatus()
		if err != nil {
			return err
		}
		a.updateStatus(status)
	}
	a.updateOppShots()
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

func (a *App) updateOppShots() {
	for _, coord := range a.status.OppShots {
		x, y, _ := parseCoords(coord)

		if a.playerBoard[x][y] == gui.Ship {
			a.playerBoard[x][y] = gui.Hit
		} else if a.playerBoard[x][y] == gui.Empty {
			a.playerBoard[x][y] = gui.Miss
		}
	}
}

func (a *App) parseBoard(b client.Board) error {
	a.playerBoard = [10][10]gui.State{}
	a.opponentBoard = [10][10]gui.State{}
	for i := range a.playerBoard {
		a.playerBoard[i] = [10]gui.State{}
		a.opponentBoard[i] = [10]gui.State{}

		for j := range a.playerBoard[i] {
			a.playerBoard[i][j] = gui.Empty
			a.opponentBoard[i][j] = gui.Empty
		}
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

func (a *App) InitGUI() {
	ui := gui.NewGUI(true)
	b1 := gui.NewBoard(2, 4, nil)
	b2 := gui.NewBoard(50, 4, nil)
	ui.Draw(b1)
	ui.Draw(b2)

	b1.SetStates(a.playerBoard)
	b2.SetStates(a.opponentBoard)

	exitInfo := gui.NewText(2, 2, "Press Ctrl+C to exit", nil) // initialize some text object
	ui.Draw(exitInfo)
	renderDescriptions(ui, a.status.Desc, a.status.OppDesc)
	a.gui.board1 = b1
	a.gui.board2 = b2
	a.gui.ui = ui
	a.gui.text = exitInfo
}

func (a *App) handleShot() string {
	a.gui.text.SetText("Choose your target:")
	for {
		coords := a.gui.board2.Listen(context.TODO())
		a.gui.text.SetText(fmt.Sprintf("Coordinate: %s", coords))
		a.gui.ui.Log("Coordinate: %s", coords) // logs are displayed after the game exits

		x, y, _ := parseCoords(coords)
		if a.opponentBoard[x][y] != gui.Empty {
			slog.Info("app [handleShot]", slog.Any("wrong_coord", coords), slog.Any("value", a.opponentBoard[x][y]))
			a.gui.text.SetText("Choose again!")
		} else {
			slog.Info("app [handleShot]", slog.Any("correct_coord", coords), slog.Any("value", a.opponentBoard[x][y]))
			return coords
		}
	}
}

func (a *App) updateBoard() {
	a.gui.board1.SetStates(a.playerBoard)
	a.gui.board2.SetStates(a.opponentBoard)
}

func renderDescriptions(g *gui.GUI, playerDesc, oppDesc string) {
	fragments := strings.Split(wordwrap.WrapString(playerDesc, 40), "\n")
	for i, f := range fragments {
		g.Draw(gui.NewText(2, 26+i, f, nil))
	}

	fragments = strings.Split(wordwrap.WrapString(oppDesc, 40), "\n")
	for i, f := range fragments {
		g.Draw(gui.NewText(50, 26+i, f, nil))
	}
}
