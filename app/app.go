package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/charmbracelet/log"
	gui "github.com/grupawp/warships-gui/v2"
	"github.com/wojtekolesinski/battleships/client"
	"github.com/wojtekolesinski/battleships/models"
	"strings"
	"time"
)

var ErrorGameEnded = fmt.Errorf("game ended")
var maxRequests = 3

type App struct {
	client        *client.Client
	playerBoard   [10][10]gui.State
	opponentBoard [10][10]gui.State
	status        models.StatusData
	ui            *ui
}

func New(c *client.Client) *App {
	return &App{
		client: c,
	}
}

func (a *App) updateStatus() (err error) {
	var status models.StatusData
	makeRequest(func() error {
		status, err = a.client.GetStatus()
		return err
	})
	if err != nil {
		return fmt.Errorf("client.GetStatus %w", err)
	}
	log.Info("app [updateStatus]", "status", status)
	a.status.ShouldFire = status.ShouldFire
	a.status.GameStatus = status.GameStatus
	a.status.OppShots = status.OppShots
	a.status.LastGameStatus = status.LastGameStatus
	a.status.Timer = status.Timer
	return
}

func (a *App) updateDescription() (err error) {
	var status models.StatusData
	makeRequest(func() error {
		status, err = a.client.GetDescription()
		return err
	})
	if err != nil {
		return fmt.Errorf("client.GetDescription: %w", err)
	}
	log.SetDefault(log.Default().With("nick", status.Nick))
	a.status.Nick = status.Nick
	a.status.Desc = status.Desc
	a.status.Opponent = status.Opponent
	a.status.OppDesc = status.OppDesc
	log.Info("app [updateDescription]", "status", a.status)
	return nil
}

func (a *App) Run() (err error) {
	name, desc := getNameAndDescription()
	targetNick, playWithBot, err := a.getOpponent()
	if err != nil {
		return fmt.Errorf("app.getOpponent: %w", err)
	}

	makeRequest(func() error {
		err = a.client.InitGame(name, desc, targetNick, playWithBot)
		return err
	})
	if err != nil {
		return fmt.Errorf("client.InitGame: %w", err)
	}

	refreshCtx, cancelRefresh := context.WithCancel(context.Background())
	defer cancelRefresh()
	go func() {
		for {
			time.Sleep(10 * time.Second)
			select {
			case <-refreshCtx.Done():
				return
			default:
				makeRequest(func() error {
					err = a.client.RefreshSession()
					return err
				})
				if err != nil {
					log.Error("client.RefreshSession", err)
				}
			}
		}
	}()

	err = a.updateStatus()
	if err != nil {
		return fmt.Errorf("app.updateStatus: %w", err)
	}

	log.Info("app [Run] - waiting for the game to start")
	for !a.gameInProgress() {
		time.Sleep(time.Second)
		err = a.updateStatus()
		if err != nil {
			return fmt.Errorf("app.updateStatus: %w", err)
		}
	}
	cancelRefresh()

	err = a.updateDescription()
	if err != nil {
		return fmt.Errorf("app.updateDescription: %w", err)
	}

	var board models.Board
	makeRequest(func() error {
		board, err = a.client.GetBoard()
		return err
	})
	if err != nil {
		return fmt.Errorf("client.GetBoard: %w", err)
	}

	log.Info("app [Run] - parsing board")
	err = a.parseBoard(board)
	if err != nil {
		return fmt.Errorf("parseBoard: %w", err)
	}
	log.Info("app [Run] - initializing gui")
	a.ui = newUi()
	a.ui.renderDescriptions(a.status.Desc, a.status.OppDesc)
	a.updateBoard()

	errChan := make(chan error, 0)

	uiCtx, stopUi := context.WithCancel(context.Background())
	go func() {
		log.Info("app [Run] - starting gameloop", "status", a.status)
		for a.gameInProgress() {
			err = a.waitForYourTurn()
			if err != nil {
				if errors.Is(err, ErrorGameEnded) {
					log.Info("app [Run] - game ended")
					break
				}
				errChan <- fmt.Errorf("app.waitForYourTurn: %w", err)
				return
			}

			err = a.shoot()
			if err != nil {
				errChan <- fmt.Errorf("app.shoot: %w", err)
				return
			}

			time.Sleep(1 * time.Second)
			err = a.updateStatus()
			if err != nil {
				errChan <- fmt.Errorf("app.updateStatus: %w", err)
				return
			}
		}
		log.Info("app [Run] - exited gameloop")
		a.updateOppShots()
		a.updateBoard()
		a.ui.renderGameResult(a.status.LastGameStatus)
		stopUi()
	}()

	go func() {
		for {
			select {
			case err = <-errChan:
				stopUi()
				errChan <- err
			default:
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()

	log.Info("app [Run] - Starting ui")
	a.ui.gui.Start(uiCtx, nil)

	select {
	case err = <-errChan:
		return err
	default:
		return nil
	}
}

func (a *App) gameInProgress() bool {
	return a.status.GameStatus == "game_in_progress"
}

func (a *App) waitForYourTurn() error {
	log.Info("app [waitForYourTurn] - starting to wait")
	a.ui.setInfoText("Opponent's turn")

	for !a.status.ShouldFire {
		time.Sleep(2 * time.Second)
		err := a.updateStatus()
		if err != nil {
			return fmt.Errorf("app.updateStatus: %w", err)
		}

		if !a.gameInProgress() {
			return ErrorGameEnded
		}
	}
	a.updateOppShots()
	a.updateBoard()
	log.Info("app [waitForYourTurn]", "opp shots", strings.Join(a.status.OppShots, " "), "shouldFire", a.status.ShouldFire)
	return nil
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

func (a *App) parseBoard(b models.Board) error {
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
			return fmt.Errorf("parseCoords: %w", err)
		}
		a.playerBoard[x][y] = gui.Ship
	}
	return nil
}

func (a *App) handleShot() (string, error) {
	//err := a.updateStatus()
	//if err != nil {
	//	return "", fmt.Errorf("app.updateStatus: %w", err)
	//}
	log.Info("app [handleShot]", "status", a.status)
	a.ui.updateTime(a.status.Timer)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Second)
				a.status.Timer--
				a.ui.updateTime(a.status.Timer)
			}
		}
	}()
	a.ui.setInfoText("Choose your target:")
	for {
		coords := a.ui.board2.Listen(context.TODO())
		a.ui.setInfoText(fmt.Sprintf("Coordinate: %s", coords))
		x, y, err := parseCoords(coords)
		if err != nil {
			return "", fmt.Errorf("parseCoords: %w", err)
		}

		if a.opponentBoard[x][y] == gui.Empty {
			log.Info("app [handleShot]", "correct_coord", coords, "value", a.opponentBoard[x][y])
			cancel()
			return coords, nil
		}

		log.Info("app [handleShot]", "wrong_coord", coords, "value", a.opponentBoard[x][y])
		a.ui.setInfoText("Choose again!")
	}
}

func (a *App) getOpponent() (targetNick string, playWithBot bool, err error) {
	playWithBot = promptPlayer("Do you want to play with a bot?")
	if playWithBot {
		return
	}

	var players models.ListData
	fmt.Println("Fetching list of active players")
	makeRequest(func() error {
		players, err = a.client.GetPlayersList()
		return err
	})
	if err != nil {
		err = fmt.Errorf("client.GetPlayersList: %w", err)
		return
	}

	if len(players) == 0 {
		fmt.Println("No active players starting to wait for an invitation")
		return
	}

	if promptPlayer("Do you want to join another player?") {
		targetNick = players[promptListOfPlayers(players)].Nick
		return
	}
	fmt.Println("Waiting for an invitation")
	return
}

func (a *App) updateBoard() {
	a.ui.board1.SetStates(a.playerBoard)
	a.ui.board2.SetStates(a.opponentBoard)
}

func (a *App) shoot() error {
	var answer models.FireAnswer
	for answer.Result != "miss" && a.gameInProgress() {
		log.Info("app[Run] - handle shot")
		coord, err := a.handleShot()
		if err != nil {
			return fmt.Errorf("handleShot: %w", err)
		}

		makeRequest(func() error {
			answer, err = a.client.Fire(coord)
			return err
		})
		if err != nil {
			return fmt.Errorf("client.Fire: %w", err)
		}

		x, y, err := parseCoords(coord)
		if err != nil {
			return fmt.Errorf("parseCoords: %w", err)

		}

		switch answer.Result {
		case "hit":
			a.opponentBoard[x][y] = gui.Hit
			a.ui.setInfoText("HIT")
		case "miss":
			a.opponentBoard[x][y] = gui.Miss
			a.ui.setInfoText("MISS")
		case "sunk":
			a.opponentBoard[x][y] = gui.Hit
			a.ui.setInfoText("SUNK")
		}

		a.updateBoard()
		err = a.updateStatus()
		if err != nil {
			return fmt.Errorf("app.updateStatus: %w", err)
		}
	}
	return nil
}
