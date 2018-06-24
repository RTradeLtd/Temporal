package gui

import (
	"log"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const ApplicationID = "com.rtradetechnologies"

func Run() {

	// This is used to create our root GTK application
	application, err := gtk.ApplicationNew(ApplicationID, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Fatal(err)
	}

	application.Connect("activate", func() {
		appWindow, err := gtk.ApplicationWindowNew(application)
		if err != nil {
			log.Fatal(err)
		}
		appWindow.SetTitle("Basic Application")
		appWindow.Add(generateAndSetupGrid())
		appWindow.SetDefaultSize(400, 400)
		appWindow.ShowAll()

	})
	application.Run(nil)
}

func generateAndSetupGrid() *gtk.Grid {
	grid, err := gtk.GridNew()
	if err != nil {
		log.Fatal(err)
	}

	buttonAbout, err := gtk.ButtonNewWithLabel("About")
	if err != nil {
		log.Fatal(err)
	}
	grid.Add(buttonAbout)
	return grid
}
