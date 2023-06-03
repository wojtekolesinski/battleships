package models

type GamePayload struct {
	Coords     []string `json:"coords,omitempty"`
	Desc       string   `json:"desc"`
	Nick       string   `json:"nick"`
	TargetNick string   `json:"target_nick,omitempty"`
	Wpbot      bool     `json:"wpbot"`
}

type StatusData struct {
	Desc           string   `json:"desc"`
	GameStatus     string   `json:"game_status"`
	LastGameStatus string   `json:"last_game_status"`
	Nick           string   `json:"nick"`
	OppDesc        string   `json:"opp_desc"`
	OppShots       []string `json:"opp_shots"`
	Opponent       string   `json:"opponent"`
	ShouldFire     bool     `json:"should_fire"`
	Timer          int      `json:"timer"`
}

type Board struct {
	Board []string `json:"board"`
}

type FirePayload struct {
	Coord string `json:"coord"`
}

type FireAnswer struct {
	Result string `json:"result"`
}

type ListData struct {
	GameStatus string `json:"game_status"`
	Nick       string `json:"nick"`
}

type StatsData struct {
	Nick   string `json:"nick"`
	Games  int    `json:"games"`
	Wins   int    `json:"wins"`
	Rank   int    `json:"rank"`
	Points int    `json:"points"`
}

type StatsList struct {
	Stats []StatsData `json:"stats"`
}

type StatsNick struct {
	Stats StatsData `json:"stats"`
}
