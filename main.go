package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"

	"hexdeskgo/database"
	"hexdeskgo/ui"
)

const AppVersion = "1.0.0"

func main() {
	a := app.NewWithID("com.printech.hexdeskgo")
	a.Settings().SetTheme(theme.DarkTheme())

	db, err := database.Init("hexdesk.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	w := a.NewWindow("HexDeskGo - Computer Repair Shop Management")
	w.Resize(fyne.NewSize(1200, 750))
	w.CenterOnScreen()

	ui.ShowLogin(w, db, AppVersion)
	w.ShowAndRun()
}
