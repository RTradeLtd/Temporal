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
		AddItem("Temporal", "Access the Temporal server commands", 'a', nil).
		AddItem("Client", "Access the Temporal text based user interface", 'c', client).
		AddItem("Run box", "Run the box function", 'b', box).
		AddItem("Quit", "Press to exit", 'q', func() {
			app.Stop()
		})
	if err := app.SetRoot(commandList, true).SetFocus(commandList).Run(); err != nil {
		log.Fatal(err)
	}
}

func client() {
	app := tview.NewApplication()
	blockchain := tview.NewList().ShowSecondaryText(false)
	blockchain.SetBorder(true).SetTitle("Blockchain Client Commands")
	blockchain.AddItem("Quit", "Exit client", 'q', func() {
		app.Stop()
	})
	database := tview.NewList().ShowSecondaryText(false)
	database.SetBorder(true).SetTitle("Database Client Commands")
	flex := tview.NewFlex().
		AddItem(blockchain, 0, 1, true).
		AddItem(database, 0, 1, false)
	if err := app.SetRoot(flex, true).SetFocus(flex).Run(); err != nil {
		log.Fatal(err)
	}
}

func box() {
	box := tview.NewBox().SetBorder(true).SetTitle("Temporal Administrative Console")
	if err := tview.NewApplication().SetRoot(box, true).Run(); err != nil {
		log.Fatal(err)
	}
}
