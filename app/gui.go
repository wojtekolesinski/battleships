package app

import (
	"fmt"
	"github.com/charmbracelet/log"
	gui "github.com/grupawp/warships-gui/v2"
	"github.com/mitchellh/go-wordwrap"
	"strings"
)

type ui struct {
	gui       *gui.GUI
	board1    *gui.Board
	board2    *gui.Board
	infoText  *gui.Text
	errorText *gui.Text
	exitText  *gui.Text
	timer     *gui.Text
	statsInfo *gui.Text
	fleetInfo []*gui.Text
}

var (
	modelFleet  = map[int]int{4: 1, 3: 2, 2: 3, 1: 4}
	boardConfig = &gui.BoardConfig{
		RulerColor: gui.White,
		TextColor:  gui.Black,
		EmptyColor: gui.NewColor(99, 161, 184),
		HitColor:   gui.NewColor(230, 30, 22),
		MissColor:  gui.Grey,
		ShipColor:  gui.NewColor(91, 181, 22),
		EmptyChar:  ' ',
		HitChar:    ' ',
		MissChar:   ' ',
		ShipChar:   ' ',
	}
	oppBoardConfig = &gui.BoardConfig{
		RulerColor: gui.White,
		TextColor:  gui.Black,
		EmptyColor: gui.NewColor(99, 161, 184),
		HitColor:   gui.NewColor(230, 30, 22),
		MissColor:  gui.Grey,
		ShipColor:  gui.Green,
		EmptyChar:  ' ',
		HitChar:    ' ',
		MissChar:   ' ',
		ShipChar:   ' ',
	}
	textConfig = &gui.TextConfig{
		FgColor: gui.White,
		BgColor: gui.Black,
	}
	errorConfig = &gui.TextConfig{
		FgColor: gui.NewColor(240, 0, 0),
		BgColor: gui.Black,
	}
)

func newGameUi() *ui {
	g := gui.NewGUI(false)
	board1 := gui.NewBoard(2, 6, boardConfig)
	board2 := gui.NewBoard(60, 6, oppBoardConfig)
	exitText := gui.NewText(2, 2, "Press Ctrl+C to exit", textConfig)
	infoText := gui.NewText(2, 4, "", textConfig)
	errorText := gui.NewText(60, 2, "", errorConfig)
	timer := gui.NewText(50, 15, " 60s ", &gui.TextConfig{
		FgColor: gui.NewColor(10, 10, 10),
		BgColor: gui.NewColor(255, 0, 255),
	})
	statsInfo := gui.NewText(50, 20, "0.00%", textConfig)

	g.Draw(gui.NewText(2, 40, "Legend:", textConfig))
	g.Draw(gui.NewText(2, 42, "   ", &gui.TextConfig{BgColor: boardConfig.ShipColor}))
	g.Draw(gui.NewText(6, 42, "ship", textConfig))
	g.Draw(gui.NewText(2, 43, "   ", &gui.TextConfig{BgColor: boardConfig.MissColor}))
	g.Draw(gui.NewText(6, 43, "miss (no ship)", textConfig))
	g.Draw(gui.NewText(2, 44, "   ", &gui.TextConfig{BgColor: boardConfig.HitColor}))
	g.Draw(gui.NewText(6, 44, "hit", textConfig))
	g.Draw(gui.NewText(2, 45, "   ", &gui.TextConfig{BgColor: boardConfig.EmptyColor}))
	g.Draw(gui.NewText(6, 45, "empty", textConfig))

	var fleetInfo []*gui.Text
	fleetInfo = append(fleetInfo, gui.NewText(60, 40, "Opponent's ships:", textConfig))
	g.Draw(fleetInfo[0])
	for i := 0; i < 4; i++ {
		info := gui.NewText(60, 41+i,
			fmt.Sprintf("%d masted: (%d/%d)", 4-i, modelFleet[4-i], modelFleet[4-i]),
			textConfig)
		fleetInfo = append(fleetInfo, info)
		g.Draw(info)
	}

	g.Draw(board1)
	g.Draw(board2)
	g.Draw(exitText)
	g.Draw(infoText)
	g.Draw(errorText)
	g.Draw(timer)
	g.Draw(statsInfo)
	g.Draw(gui.NewText(48, 19, "Accuracy:", nil))

	return &ui{
		gui:       g,
		board1:    board1,
		board2:    board2,
		infoText:  infoText,
		exitText:  exitText,
		timer:     timer,
		statsInfo: statsInfo,
		fleetInfo: fleetInfo,
		errorText: errorText,
	}
}

func newFleetUi() *ui {
	g := gui.NewGUI(false)
	board1 := gui.NewBoard(2, 6, boardConfig)
	exitText := gui.NewText(2, 2, "Press Ctrl+C to exit without saving", textConfig)
	infoText := gui.NewText(2, 4, "", textConfig)
	errorText := gui.NewText(2, 28, "", errorConfig)

	g.Draw(board1)
	g.Draw(exitText)
	g.Draw(infoText)
	g.Draw(errorText)

	return &ui{
		gui:       g,
		board1:    board1,
		infoText:  infoText,
		exitText:  exitText,
		errorText: errorText,
	}
}

func (u *ui) renderNicks(playerNick, oppNick string) {
	log.Debug("app [renderNicks]", "playerNick", playerNick, "oppNick", oppNick)
	u.gui.Draw(gui.NewText(2, 28, playerNick, textConfig))
	u.gui.Draw(gui.NewText(60, 28, oppNick, textConfig))
}

func (u *ui) renderDescriptions(playerDesc, oppDesc string) {
	log.Debug("app [renderDescriptions]", "playerDesc", playerDesc, "oppDesc", oppDesc)
	fragments := strings.Split(wordwrap.WrapString(playerDesc, 40), "\n")
	for i, f := range fragments {
		u.gui.Draw(gui.NewText(2, 30+i, f, textConfig))
	}

	fragments = strings.Split(wordwrap.WrapString(oppDesc, 40), "\n")
	for i, f := range fragments {
		u.gui.Draw(gui.NewText(60, 30+i, f, textConfig))
	}
}

func (u *ui) setFleetInfo(fleet map[int]int) {
	for i := 0; i < 4; i++ {
		u.fleetInfo[i+1].SetText(fmt.Sprintf("%d masted: (%d/%d)", 4-i, fleet[4-i], modelFleet[4-i]))
	}
}

func (u *ui) setInfoText(text string) {
	u.infoText.SetText(text)
}

func (u *ui) setExitText(text string) {
	u.exitText.SetText(text)
}

func (u *ui) setErrorText(text string) {
	u.errorText.SetText(text)
}

func (u *ui) resetErrorText() {
	u.errorText.SetText("")
}

func (u *ui) renderGameResult(result string) {
	if result == "win" {
		u.infoText.SetBgColor(gui.Green)
		u.infoText.SetFgColor(gui.White)
		u.setInfoText("You win")
	} else if result == "lose" {
		u.infoText.SetBgColor(gui.Red)
		u.infoText.SetFgColor(gui.White)
		u.setInfoText("You lose")
	} else {
		// TODO
	}
}

func (u *ui) updateTime(time int) {
	u.timer.SetText(fmt.Sprintf(" %ds ", time))
	if time <= 5 {
		u.timer.SetBgColor(gui.NewColor(250, 0, 0))
	} else if time > 55 {
		u.timer.SetBgColor(gui.NewColor(255, 0, 255))
	} else {
		// TODO
	}
}

func (u *ui) updateAccuracy(accuracy float32) {
	u.statsInfo.SetText(fmt.Sprintf("%.2f%%", accuracy))
}

func (u *ui) addAssistantInfo() {
	u.gui.Draw(gui.NewText(2, 46, "   ", &gui.TextConfig{BgColor: oppBoardConfig.ShipColor}))
	u.gui.Draw(gui.NewText(6, 46, "assistant's pick", textConfig))
}
