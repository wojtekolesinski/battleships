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
	totalShots    int
	hits          int
	ui            *ui
	customBoard   bool
}

func New(c *client.Client) *App {
	return &App{
		client: c,
	}
}

func (a *App) Run() error {
	a.getNameAndDescription()

	for {
		gamePayload, err := a.displayMenu()
		if err != nil {
			return fmt.Errorf("app.displayMenu: %w", err)
		}

		err = a.initGame(gamePayload)
		if err != nil {
			return fmt.Errorf("app.initGame: %w", err)
		}

		errChan := make(chan error, 0)
		ctx, cancelFunc := context.WithCancel(context.Background())
		go a.loop(ctx, errChan, cancelFunc)
		go func() {
			for {
				select {
				case err = <-errChan:
					cancelFunc()
					errChan <- err
				default:
					time.Sleep(500 * time.Millisecond)
				}
			}
		}()

		log.Info("app [Run] - Starting ui")
		a.ui.gui.Start(ctx, nil)
		log.Info("app [Run] - abandoning game")
		err = a.client.AbandonGame()
		if err != nil {
			return fmt.Errorf("client.AbandonGame: %w", err)
		}

		select {
		case err = <-errChan:
			return err
		default:
			continue
		}
	}
}

func (a *App) loop(ctx context.Context, errChan chan error, cancelFunc context.CancelFunc) {
	log.Info("app [Run] - starting gameloop", "status", a.status)
	defer cancelFunc()
	for a.gameInProgress() {
		err := a.waitForYourTurn()
		if err != nil {
			if errors.Is(err, ErrorGameEnded) {
				log.Info("app [Run] - game ended")
				break
			}
			errChan <- fmt.Errorf("app.waitForYourTurn: %w", err)
			return
		}

		err = a.shoot(ctx)
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
	for i := 5; i > 0; i-- {
		a.ui.setExitText(fmt.Sprintf("Exiting in %ds", i))
		time.Sleep(1 * time.Second)
	}
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

		a.updateOppShots()
		a.updateBoard()

		if !a.gameInProgress() {
			return ErrorGameEnded
		}
	}
	log.Info("app [waitForYourTurn]", "opp shots", strings.Join(a.status.OppShots, " "), "shouldFire", a.status.ShouldFire)
	return nil
}

func (a *App) handleShot(ctx context.Context) (string, error) {
	log.Info("app [handleShot]", "status", a.status)
	a.ui.updateTime(a.status.Timer)
	ctx, cancel := context.WithCancel(ctx)
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

func (a *App) shoot(ctx context.Context) error {
	var answer models.FireAnswer
	for answer.Result != "miss" && a.gameInProgress() {
		log.Info("app[Run] - handle shot")
		coord, err := a.handleShot(ctx)
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

		a.totalShots++
		switch answer.Result {
		case "hit":
			a.opponentBoard[x][y] = gui.Hit
			a.ui.setInfoText("HIT")
			a.hits++
		case "miss":
			a.opponentBoard[x][y] = gui.Miss
			a.ui.setInfoText("MISS")
		case "sunk":
			a.opponentBoard[x][y] = gui.Hit
			a.ui.setInfoText("SUNK")
			a.hits++
		}

		a.updateBoard()
		a.ui.updateAccuracy(a.getAccuracy())
		err = a.updateStatus()
		if err != nil {
			return fmt.Errorf("app.updateStatus: %w", err)
		}
	}
	return nil
}

func (a *App) initGame(payload models.GamePayload) error {
	var err error
	makeRequest(func() error {
		err = a.client.InitGame(payload)
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

	log.Info("app [initGame] - waiting for the game to start")
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

	log.Info("app [initGame] - parsing board")
	err = a.parseBoard(board)
	if err != nil {
		return fmt.Errorf("parseBoard: %w", err)
	}
	log.Info("app [initGame] - initializing gui")
	a.ui = newGameUi()
	a.ui.renderDescriptions(a.status.Desc, a.status.OppDesc)
	a.updateBoard()
	return nil
}

type point struct {
	x, y int
}

func (a *App) editBoard() error {
	fleet := map[int]int{4: 1, 3: 2, 2: 3, 1: 4}

	ui := newFleetUi()
	board := [10][10]gui.State{}
	for i := range board {
		board[i] = [10]gui.State{}
		for j := range board[i] {
			board[i][j] = gui.Hit
		}
	}
	ui.board1.SetStates(board)
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	go func() {
		for length := 4; length >= 1; length-- {
			count := fleet[length]
			for s := 0; s < count; s++ {
				ship := []point{}
				ui.setInfoText(fmt.Sprintf("Placing ship with length: %d (%d/%d)", length, s+1, count))

				for i := range board {
					for j := range board[i] {
						if board[i][j] == gui.Empty {
							board[i][j] = gui.Hit
						}
					}
				}
				ui.board1.SetStates(board)

				for j := 0; j < length; j++ {
					ui.resetErrorText()
					for {
						coords := ui.board1.Listen(context.TODO())
						x, y, err := parseCoords(coords)
						if err != nil {
							log.Error(fmt.Errorf("parseCoords: %w", err))
						}

						if board[x][y] == gui.Hit {
							ship = append(ship, point{x, y})
							board[x][y] = gui.Ship
							setPossiblePositions(&board, ship)
							log.Info("app [editBoard]", "board", board)
							ui.board1.SetStates(board)
							break
						}
						ui.setErrorText("Invalid choice")
					}
				}

				setImpossiblePositions(&board, ship)
				ui.board1.SetStates(board)
			}
		}

		for i := range board {
			for j := range board[i] {
				if board[i][j] != gui.Ship {
					board[i][j] = gui.Empty
				}
			}
		}
		a.customBoard = true
		a.playerBoard = board
		log.Debug("app [editBoard]", "coords", getCoordsFromBoard(board))
	}()

	ui.gui.Start(ctx, nil)

	return nil
}

func setPossiblePositions(board *[10][10]gui.State, ship []point) {
	for i := range board {
		for j := range board[i] {
			if board[i][j] == gui.Hit {
				board[i][j] = gui.Empty
			}
		}
	}

	neighbours := []point{
		{0, 1},
		{1, 0},
		{0, -1},
		{-1, 0},
	}

	for _, p := range ship {
		log.Info("app [setPossiblePositions]", "ship", p)
		for _, offset := range neighbours {
			n := point{p.x + offset.x, p.y + offset.y}
			if n.x < 0 || n.x >= 10 || n.y < 0 || n.y >= 10 {
				continue
			}

			log.Info("app [setPossiblePositions]", "neighbour", n)

			if board[n.x][n.y] == gui.Empty {
				board[n.x][n.y] = gui.Hit
			}
		}
	}

	log.Info("app [setPossiblePositions]", "board", board)
}

func setImpossiblePositions(board *[10][10]gui.State, ship []point) {
	for i := range board {
		for j := range board[i] {
			if board[i][j] == gui.Hit {
				board[i][j] = gui.Empty
			}
		}
	}

	neighbours := []point{
		{0, 1},
		{1, 1},
		{1, 0},
		{1, -1},
		{0, -1},
		{-1, -1},
		{-1, 0},
		{-1, 1},
	}

	for _, p := range ship {
		log.Info("app [setImpossiblePositions]", "ship", p)
		for _, offset := range neighbours {
			n := point{p.x + offset.x, p.y + offset.y}
			if n.x < 0 || n.x >= 10 || n.y < 0 || n.y >= 10 {
				continue
			}

			log.Info("app [setImpossiblePositions]", "neighbour", n)

			if board[n.x][n.y] == gui.Empty {
				board[n.x][n.y] = gui.Miss
			}
		}
	}

	log.Info("app [setImpossiblePositions]", "board", board)
}
