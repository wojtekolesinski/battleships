package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/charmbracelet/log"
	gui "github.com/grupawp/warships-gui/v2"
	"github.com/wojtekolesinski/battleships/client"
	"github.com/wojtekolesinski/battleships/models"
	"strconv"
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

		a.updateOppShots()
		a.updateBoard()

		if !a.gameInProgress() {
			return ErrorGameEnded
		}
	}
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

func (a *App) getAccuracy() float32 {
	if a.totalShots == 0 {
		return 0
	}
	return 100 * float32(a.hits) / float32(a.totalShots)
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

func (a *App) getOpponent() (targetNick string, err error) {
	var players []models.ListData
	fmt.Println("Fetching list of active players")
	makeRequest(func() error {
		players, err = a.client.GetPlayersList()
		return err
	})
	if err != nil {
		err = fmt.Errorf("client.GetPlayersList: %w", err)
		return
	}

	players = append([]models.ListData{{Nick: "wp_bot"}}, players...)

	choice := promptList(players, 0, func(a models.ListData) string { return a.Nick })

	return players[choice].Nick, nil
}

func (a *App) getNameAndDescription() {
	var name, desc string
	for {
		fmt.Print("Insert your name (leave blank to get one assigned): ")
		_, err := fmt.Scanln(&name)
		if err == nil || err.Error() == "unexpected newline" {
			break
		} else {
			log.Error("app [getNameAndDescripiton]", "err", err, "name", name)
		}

	}

	for {
		fmt.Print("Insert your description (leave blank to get one assigned): ")
		_, err := fmt.Scanln(&desc)
		if err == nil || err.Error() == "unexpected newline" {
			break
		} else {
			log.Error("app [getNameAndDescripiton]", "err", err, "desc", desc)
		}

	}
	a.status.Nick = name
	a.status.Desc = desc
}

func (a *App) updateBoard() {
	a.ui.board1.SetStates(a.playerBoard)
	a.ui.board2.SetStates(a.opponentBoard)
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

func (a *App) displayMenu() (models.GamePayload, error) {
	choices := []string{
		"Join a game",
		"Wait for an opponent",
		"Display top 10 stats",
		"Display your stats",
	}

	choice := promptList(choices, 1, func(a string) string { return a })
	log.Info("app [displayMenu]", "choice", choice)

	switch choice {
	case 1:
		targetNick, err := a.getOpponent()
		if err != nil {
			return models.GamePayload{}, fmt.Errorf("app.getOpponent: %w", err)
		}
		return a.getGamePayload(targetNick), nil
	case 2:
		fmt.Println("Waiting for an invitation...")
		return a.getGamePayload(""), nil
	case 3:
		err := a.displayTop10Stats()
		if err != nil {
			return models.GamePayload{}, fmt.Errorf("app.displayTop10Stats: %w", err)
		}
	case 4:
		err := a.displayPlayerStats()
		if err != nil {
			return models.GamePayload{}, fmt.Errorf("app.displayPlayerStats: %w", err)
		}
	}
	return a.displayMenu()
}

func (a *App) playGame() {

}

func (a *App) displayTop10Stats() error {
	var stats models.StatsList
	var err error
	makeRequest(func() error {
		stats, err = a.client.GetStats()
		return err
	})
	if err != nil {
		return fmt.Errorf("client.GetStats: %w", err)
	}

	fmt.Println()
	fmt.Printf("| %4s | %-20s | %-5s | %4s | %6s |\n", "RANK", "NICK", "Games", "Wins", "Points")
	for _, s := range stats.Stats {
		fmt.Printf("| %4s | %-20s | %5s | %4s | %6s |\n",
			strconv.Itoa(s.Rank),
			s.Nick,
			strconv.Itoa(s.Games),
			strconv.Itoa(s.Wins),
			strconv.Itoa(s.Points),
		)
	}
	fmt.Println()
	return nil
}

func (a *App) displayPlayerStats() error {
	var stats models.StatsNick
	var err error
	makeRequest(func() error {
		stats, err = a.client.GetPlayerStats(a.status.Nick)
		return err
	})
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			fmt.Println("\nNo stats for player\n")
			return nil
		}
		return fmt.Errorf("client.GetPlayerStats: %w", err)
	}

	fmt.Println()
	fmt.Printf("| %4s | %-20s | %-5s | %4s | %6s |\n", "RANK", "NICK", "Games", "Wins", "Points")
	s := stats.Stats
	fmt.Printf("| %4s | %-20s | %5s | %4s | %6s |\n",
		strconv.Itoa(s.Rank),
		s.Nick,
		strconv.Itoa(s.Games),
		strconv.Itoa(s.Wins),
		strconv.Itoa(s.Points),
	)

	fmt.Println()
	return nil
}

func (a *App) getGamePayload(targetNick string) models.GamePayload {
	log.Info("app [getGamePayload]", "targetNick", targetNick)
	payload := models.GamePayload{
		Nick: a.status.Nick,
		Desc: a.status.Desc,
	}

	if targetNick == "wp_bot" {
		payload.Wpbot = true
	} else {
		payload.TargetNick = targetNick
	}
	return payload
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
	a.ui = newUi()
	a.ui.renderDescriptions(a.status.Desc, a.status.OppDesc)
	a.updateBoard()
	return nil
}
