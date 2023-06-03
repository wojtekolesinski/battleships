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
	exitText  *gui.Text
	timer     *gui.Text
	statsInfo *gui.Text
}

func newUi() *ui {
	g := gui.NewGUI(true)
	board1 := gui.NewBoard(2, 6, nil)
	board2 := gui.NewBoard(60, 6, nil)
	exitText := gui.NewText(2, 2, "Press Ctrl+C to exit", nil)
	infoText := gui.NewText(2, 4, "", nil)
	timer := gui.NewText(50, 15, " 60s ", &gui.TextConfig{
		FgColor: gui.NewColor(10, 10, 10),
		BgColor: gui.NewColor(255, 0, 255),
	})
	statsInfo := gui.NewText(50, 20, "0.00%", &gui.TextConfig{FgColor: gui.White, BgColor: gui.Black})

	g.Draw(board1)
	g.Draw(board2)
	g.Draw(exitText)
	g.Draw(infoText)
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
	}
}

func (u *ui) renderDescriptions(playerDesc, oppDesc string) {
	log.Info("app [renderDescriptions]", "playerDesc", playerDesc, "oppDesc", oppDesc)
	fragments := strings.Split(wordwrap.WrapString(playerDesc, 40), "\n")
	for i, f := range fragments {
		u.gui.Draw(gui.NewText(2, 28+i, f, nil))
	}

	fragments = strings.Split(wordwrap.WrapString(oppDesc, 40), "\n")
	for i, f := range fragments {
		u.gui.Draw(gui.NewText(60, 28+i, f, nil))
	}
}

func (u *ui) setInfoText(text string) {
	u.infoText.SetText(text)
}

func (u *ui) setExitText(text string) {
	u.exitText.SetText(text)
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
