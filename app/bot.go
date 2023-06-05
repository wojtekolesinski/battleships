package app

import (
	"github.com/charmbracelet/log"
	gui "github.com/grupawp/warships-gui/v2"
)

type Board [10][10]gui.State

var possibleShapes = map[int][][]point{
	4: {
		// lines
		{{0, 0}, {1, 0}, {2, 0}, {3, 0}},
		{{0, 0}, {0, 1}, {0, 2}, {0, 3}},
		// l-shape mirrored
		{{0, 0}, {1, 0}, {2, 0}, {2, 1}},
		{{0, 0}, {0, 1}, {0, 2}, {-1, 2}},
		{{0, 0}, {0, 1}, {1, 1}, {2, 1}},
		{{0, 0}, {1, 0}, {0, 1}, {0, 2}},
		// l-shape
		{{0, 0}, {1, 0}, {2, 0}, {0, 1}},
		{{0, 0}, {1, 0}, {1, 1}, {1, 2}},
		{{0, 0}, {1, 0}, {2, 0}, {2, -1}},
		{{0, 0}, {0, 1}, {0, 2}, {1, 2}},
		// t-shape
		{{0, 0}, {1, 0}, {2, 0}, {1, 1}},
		{{0, 0}, {0, 1}, {0, 2}, {-1, 1}},
		{{0, 0}, {1, 0}, {2, 0}, {1, -1}},
		{{0, 0}, {0, 1}, {0, 2}, {1, 1}},
		// square
		{{0, 0}, {0, 1}, {1, 0}, {1, 1}},
	},
	3: {
		// lines
		{{0, 0}, {1, 0}, {2, 0}},
		{{0, 0}, {0, 1}, {0, 2}},
		// corner-missing square
		{{0, 0}, {0, 1}, {1, 1}},
		{{0, 0}, {0, 1}, {1, 0}},
		{{0, 0}, {1, 0}, {1, 1}},
		{{0, 0}, {0, 1}, {-1, 1}},
	},
	2: {
		{{0, 0}, {1, 0}},
		{{0, 0}, {0, 1}},
	},
	1: {
		{{0, 0}},
	},
}

type bot struct {
	targets []point
	board   Board
	fleet   map[int]int
}

func (b *bot) getRecommendation(board Board, fleet map[int]int) point {
	var rec point
	if len(b.targets) > 0 {
		rec, b.targets = b.targets[0], b.targets[1:]
		return rec
	}
	return point{}
}

func (b *bot) hit(board Board, x, y int)  {}
func (b *bot) sunk(board Board, x, y int) {}

type genState struct {
	shipsLeft map[int]int
}

func NewGenState() genState {
	return genState{
		shipsLeft: map[int]int{4: 1, 3: 2, 2: 3, 1: 4},
	}
}

func (s genState) copy() genState {
	var state genState
	state.shipsLeft = make(map[int]int)
	for k, v := range s.shipsLeft {
		state.shipsLeft[k] = v
	}
	return state
}

func GenerateBoards(board Board, state genState) (boards []Board) {
	var currLength int
	for i := 4; i >= 1; i-- {
		if state.shipsLeft[i] > 0 {
			currLength = i
			break
		}
	}

	if currLength == 0 {
		return []Board{board}
	}

	log.Info("bot [GenerateBoards]", "currLength", currLength)
	for x := range board {
		for y := range board[x] {
			if board[x][y] != gui.Empty {
				continue
			}

			for _, ship := range possibleShapes[currLength] {
				if fits(ship, board, x, y) {
					s := state.copy()
					s.shipsLeft[currLength]--
					boards = append(boards, GenerateBoards(placeShip(board, ship, x, y), s)...)
				}
			}
		}
	}
	return boards
}

func GenerateBoards2(board Board, state genState) [10][10]int {
	var probs [10][10]int
	for i := range probs {
		probs[i] = [10]int{}
	}

	for length := 4; length >= 1; length-- {
		if state.shipsLeft[length] == 0 {
			continue
		}

		for x := range board {
			for y := range board[x] {
				if board[x][y] != gui.Empty {
					continue
				}
				for _, ship := range possibleShapes[length] {
					if fits(ship, board, x, y) {
						for _, p := range ship {
							probs[x+p.x][y+p.y]++
						}
					}
				}
			}
		}
	}
	return probs

}

func placeShip(board Board, ship []point, x int, y int) Board {
	for _, p := range ship {
		board[p.x+x][p.y+y] = gui.Ship
	}

	board = setImpossiblePositions(board, ship)
	return board
}

func fits(ship []point, board Board, x int, y int) bool {
	for _, p := range ship {
		n := point{p.x + x, p.y + y}
		if n.x < 0 || n.x >= 10 || n.y < 0 || n.y >= 10 {
			return false
		}
		if board[n.x][n.y] != gui.Empty {
			return false
		}
	}
	return true
}
