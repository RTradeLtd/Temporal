package tui

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/RTradeLtd/Temporal/api"
	"github.com/RTradeLtd/Temporal/config"
	"github.com/rivo/tview"
)

var app *tview.Application
var pages *tview.Pages
var title = "Temporal Administrative Console"
var tCfg *config.TemporalConfig

// Initializes the Terminal User Interface
func InitializeBox() {
	box := tview.NewBox().SetBorder(true).SetTitle(title)
	if err := tview.NewApplication().SetRoot(box, true).Run(); err != nil {
		log.Fatal(err)
	}
}

func InitializeApplication(tCfg *config.TemporalConfig) {

	// create the tview app
	app = tview.NewApplication()
	pages = tview.NewPages()
	// list of possible commands
	commandList := tview.NewList().
		AddItem("Temporal", "Access the Temporal server commands", 'a', temporal).
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

func temporal() {
	configDag := os.Getenv("CONFIG_DAG")
	if configDag == "" {
		log.Fatal(errors.New("CONFIG_DAG env not present"))
	}
	tCfg := config.LoadConfig(configDag)
	// jwtKey, rollbarToken, mqConnectionURL, dbPass, dbURL, ethKey, ethPass, listenAddress, dbUser string
	temporalCMDList := tview.NewList().ShowSecondaryText(true)
	temporalCMDList.AddItem("Start API", "Start the Temporal API", 'a', func() {
		//https://stackoverflow.com/questions/35333302/how-to-write-the-output-of-this-statement-into-a-file-in-golang
		file, err := os.Create("/tmp/tui.log")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		w := bufio.NewWriter(file)
		// jwtKey, rollbarToken, mqConnectionURL, dbPass, dbURL, ethKey, ethPass, listenAddress, dbUser string
		router := api.Setup(tCfg.API.JwtKey, tCfg.API.RollbarToken, tCfg.RabbitMQ.URL, tCfg.Database.Password, tCfg.Database.URL,
			tCfg.Ethereum.Account.KeyFile, tCfg.Ethereum.Account.KeyPass, tCfg.API.Connection.ListenAddress, tCfg.Database.Username)
		fmt.Fprint(w, router.RunTLS(fmt.Sprintf("%s:6767", tCfg.API.Connection.ListenAddress),
			tCfg.API.Connection.Certificates.CertPath, tCfg.API.Connection.Certificates.KeyPath))
		/*
			temporalCMDList.Clear()
			app.SetRoot(pages, true).SetFocus(pages.ShowPage("Command List"))*/
	})

	flex := tview.NewFlex().
		AddItem(temporalCMDList, 0, 1, true)
	flex.SetBorder(true)
	flex.SetTitle(title)

	app.SetRoot(flex, true)
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
	box := tview.NewBox().SetBorder(true).SetTitle(title)
	if err := app.SetRoot(box, true).Run(); err != nil {
		log.Fatal(err)
	}
}
