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
		// create an application window attached to our main application
		appWindow, err := gtk.ApplicationWindowNew(application)
		if err != nil {
			log.Fatal(err)
		}
		// Set the title
		appWindow.SetTitle("Temporal GUI")
		// Set the size
		appWindow.SetDefaultSize(400, 400)
		grid, err := gtk.GridNew()
		if err != nil {
			log.Fatal(err)
		}
		// Lets create a button
		btnTest, err := gtk.ButtonNewWithLabel("test")
		if err != nil {
			log.Fatal(err)
		}
		btnAbout, err := gtk.ButtonNewWithLabel("about")
		if err != nil {
			log.Fatal(err)
		}
		// Add the button
		grid.Add(btnTest)
		// Attach the button can be used to specify where in teh grid
		// to attach to
		grid.Attach(btnAbout, 1, 1, 1, 1)
		appWindow.Add(grid)
		// Show all widgets
		appWindow.ShowAll()
	})
	application.Run(nil)
}
