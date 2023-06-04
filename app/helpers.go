package app

import (
	"fmt"
	"github.com/charmbracelet/log"
	gui "github.com/grupawp/warships-gui/v2"
	"github.com/wojtekolesinski/battleships/models"
	"strconv"
)

func promptList[T any](list []T, start int, mapper func(T) string) int {
	for i, el := range list {
		fmt.Printf("(%d)\t%s\n", start+i, mapper(el))
	}

	var res string
	var choice int
	for {
		fmt.Print("Your choice: ")
		_, err := fmt.Scanf("%s", &res)
		if err != nil {
			fmt.Printf("Try again 1 %s\n", err)
			continue
		}
		choice, err = strconv.Atoi(res)
		if err != nil {
			fmt.Printf("Try again 2 %s\n", err)
			continue
		}

		if choice >= start && choice < len(list)+start {
			return choice
		}
	}
}

func makeRequest(target func() error) {
	for i := 0; i < maxRequests; i++ {
		err := target()
		if err == nil {
			return
		}
		log.Error("app [makeRequest]", "err", err)
	}
}

func promptPlayer(prompt string) bool {
	var res string
	for {
		fmt.Print(fmt.Sprintf("%s (y/n): ", prompt))
		_, err := fmt.Scanln(&res)
		if err == nil {
			if res == "y" {
				return true
			} else if res == "n" {
				return false
			}
		} else {
			log.Error("app [promptPlayWithBot]", "err", err, "res", res)
		}
	}

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

func (a *App) gameInProgress() bool {
	return a.status.GameStatus == "game_in_progress"
}

func (a *App) getAccuracy() float32 {
	if a.totalShots == 0 {
		return 0
	}
	return 100 * float32(a.hits) / float32(a.totalShots)
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

func (a *App) getGamePayload(targetNick string) models.GamePayload {
	log.Debug("app [getGamePayload]", "targetNick", targetNick)
	payload := models.GamePayload{
		Nick: a.status.Nick,
		Desc: a.status.Desc,
	}

	if targetNick == "wp_bot" {
		payload.Wpbot = true
	} else {
		payload.TargetNick = targetNick
	}

	if a.customBoard {
		log.Debug("app [getGamePayload] - adding custom board")
		payload.Coords = getCoordsFromBoard(a.playerBoard)
	}

	return payload
}

func getCoordsFromBoard(board [10][10]gui.State) []string {
	var coords []string

	for x := range board {
		for y := range board[x] {
			if board[x][y] == gui.Ship {
				coords = append(coords, fmt.Sprintf("%c%d", x+'A', y+1))
			}
		}
	}

	return coords
}

func (a *App) updateBoard() {
	log.Debug("app [updateBoard]")
	a.ui.board1.SetStates(a.playerBoard)
	a.ui.board2.SetStates(a.opponentBoard)
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
	log.Debug("app [updateDescription]", "status", a.status)
	return nil
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

func (a *App) updateStatus() (err error) {
	var status models.StatusData
	makeRequest(func() error {
		status, err = a.client.GetStatus()
		return err
	})
	if err != nil {
		return fmt.Errorf("client.GetStatus %w", err)
	}
	log.Debug("app [updateStatus]", "status", status)
	a.status.ShouldFire = status.ShouldFire
	a.status.GameStatus = status.GameStatus
	a.status.OppShots = status.OppShots
	a.status.LastGameStatus = status.LastGameStatus
	a.status.Timer = status.Timer
	return
}
