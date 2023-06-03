package app

import (
	"fmt"
	"github.com/charmbracelet/log"
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
