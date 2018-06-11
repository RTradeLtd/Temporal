package tui

import (
	"log"

	"github.com/rivo/tview"
)

var app *tview.Application
var pages *tview.Pages
var title = "Temporal Administrative Console"

// Initializes the Terminal User Interface
func InitializeBox() {
	box := tview.NewBox().SetBorder(true).SetTitle(title)
	if err := tview.NewApplication().SetRoot(box, true).Run(); err != nil {
		log.Fatal(err)
	}
}

func InitializeApplication() {
	// create the tview app
	app = tview.NewApplication()
	pages = tview.NewPages()
	// list of possible commands
	commandList := tview.NewList().
		AddItem("Temporal", "Access the Temporal server commands", 'a', nil).
		AddItem("Client", "Access the Temporal text based user interface", 'c', client).
		AddItem("Run box", "Run the box function", 'b', box).
		AddItem("Quit", "Press to exit", 'q', func() {
			app.Stop()
		})

	pages.AddPage("Command List", commandList, true, true).SetBorder(true).SetTitle(title)
	if err := app.SetRoot(pages, true).SetFocus(pages).Run(); err != nil {
		log.Fatal(err)
	}
}

func client() {
	blockchain := tview.NewList().ShowSecondaryText(false)
	blockchain.SetBorder(true).SetTitle("Blockchain Client Commands")
	blockchain.AddItem("Return", "Return to main menu", 'r', func() {
		// this allows you to switch back to the main command listing
		blockchain.Clear()
		app.SetRoot(pages, true).SetFocus(pages.ShowPage("Command List"))
	})
	blockchain.AddItem("Quit", "Exit client", 'q', func() {
		app.Stop()
	})

	database := tview.NewList().ShowSecondaryText(false)
	database.SetBorder(true).SetTitle("Database Client Commands")
	database.AddItem("Return", "Return to main menu", 'r', func() {
		database.Clear()
		app.SetRoot(pages, true).SetFocus(pages.ShowPage("Command List"))
	})
	flex := tview.NewFlex().
		AddItem(blockchain, 0, 1, true).
		AddItem(database, 0, 1, false)
	app.SetRoot(flex, true)
}

func box() {
	box := tview.NewBox().SetBorder(true).SetTitle("Temporal Administrative Console")
	if err := app.SetRoot(box, true).Run(); err != nil {
		log.Fatal(err)
	}
}
