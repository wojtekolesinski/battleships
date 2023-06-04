package app

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/wojtekolesinski/battleships/client"
	"github.com/wojtekolesinski/battleships/models"
	"strconv"
)

func (a *App) displayMenu() (models.GamePayload, error) {
	for {
		choices := []string{
			"Join a game",
			"Wait for an opponent",
			"Display top 10 stats",
			"Display your stats",
			"Modify your board",
		}

		choice := promptList(choices, 1, func(a string) string { return a })
		log.Debug("app [displayMenu]", "choice", choice)

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
		case 5:
			err := a.editBoard()
			if err != nil {
				return models.GamePayload{}, fmt.Errorf("app.editBoard: %w", err)
			}
		}
	}

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
	fmt.Printf("| %4s | %-20s | %-5s | %4s | %6s |\n", "RANK", "NICK", "GAMES", "WINS", "POINTS")
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
	fmt.Printf("| %4s | %-20s | %-5s | %4s | %6s |\n", "RANK", "NICK", "GAMES", "WINS", "POINTS")
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
