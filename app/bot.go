package app

import (
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
}

func newBot() *bot {
	return &bot{targets: []point{}}
}

func (b *bot) getRecommendation(board Board, fleet map[int]int) point {
	if len(b.targets) > 0 {
		for i := range b.targets {
			rec := b.targets[i]
			if board[rec.x][rec.y] == gui.Empty {
				return rec
			}
		}
	}

	probs := generateProbs(board, fleet)
	var max, x, y int

	for i := range probs {
		for j := range probs[i] {
			if probs[i][j] > max {
				max = probs[i][j]
				x = i
				y = j
			}
		}
	}
	return point{x, y}
}

func (b *bot) hit(board Board, x, y int) {
	neighbours := []point{
		{0, 1},
		{1, 0},
		{0, -1},
		{-1, 0},
	}

	for _, offset := range neighbours {
		n := point{x + offset.x, y + offset.y}
		if n.x < 0 || n.x >= 10 || n.y < 0 || n.y >= 10 {
			continue
		}

		if board[n.x][n.y] == gui.Empty {
			b.targets = append(b.targets, n)
		}
	}

}
func (b *bot) sunk() {
	b.targets = []point{}
}

func generateProbs(board Board, fleet map[int]int) [10][10]int {
	var probs [10][10]int
	for i := range probs {
		probs[i] = [10]int{}
	}

	for length := 4; length >= 1; length-- {
		if fleet[length] == 0 {
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
