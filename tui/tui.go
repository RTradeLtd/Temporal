package tui

import (
	"log"

	"github.com/rivo/tview"
)

// Initializes the Terminal User Interface
func InitializeBox() {
	box := tview.NewBox().SetBorder(true).SetTitle("Temporal Administrative Console")
	if err := tview.NewApplication().SetRoot(box, true).Run(); err != nil {
		log.Fatal(err)
	}
}

func InitializeApplication() {
	// create the tview app
	app := tview.NewApplication()
	// list of possible commands
	commandList := tview.NewList().
		AddItem("Run API", "Run the Temporal API software", 'a', nil).
		AddItem("Run DPA Queue", "Run the Datapase Pin Add Queue", 'b', nil).
		AddItem("Run box", "Run the box function", 'c', box).
		AddItem("Quit", "Press to exit", 'q', func() {
			app.Stop()
		})
	if err := app.SetRoot(commandList, true).SetFocus(commandList).Run(); err != nil {
		log.Fatal(err)
	}
}

func box() {
	box := tview.NewBox().SetBorder(true).SetTitle("Temporal Administrative Console")
	if err := tview.NewApplication().SetRoot(box, true).Run(); err != nil {
		log.Fatal(err)
	}
}
