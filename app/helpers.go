package app

import (
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/wojtekolesinski/battleships/models"
	"strconv"
)

func getNameAndDescription() (string, string) {
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

	return name, desc

}

func promptListOfPlayers(data models.ListData) int {
	for i, player := range data {
		fmt.Printf("(%d)\t%s\t%s\n", i, player.Nick, player.GameStatus)
	}
	var res int
	fmt.Print("Choose an opponent: ")
	fmt.Scanf("%d", &res)
	fmt.Printf("Chosen: %d\n", res)
	return res

}

func makeRequest(target func() error) {
	for i := 0; i < maxRequests; i++ {
		err := target()
		if err == nil {
			return
		}
		log.Error("app [makeRequest]", err)
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
