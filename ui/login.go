package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"hexdeskgo/auth"
	"hexdeskgo/database"
)

func ShowLogin(w fyne.Window, db *database.DB, version string) {
	title := canvas.NewText("HexDeskGo", color.NRGBA{R: 0x2E, G: 0x86, B: 0xAB, A: 0xFF})
	//title := widget.NewLabelWithStyle("HexDeskGo", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	title.TextSize = 42
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabelWithStyle("Computer Repair Shop Management",
		fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("Username")
	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder("Password")

	loginBtn := widget.NewButtonWithIcon("Sign In", theme.LoginIcon(), nil)

	doLogin := func() {
		u, err := auth.Authenticate(db, userEntry.Text, passEntry.Text)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		ShowMain(w, db, &auth.Session{User: u}, version)
	}
	loginBtn.OnTapped = doLogin
	passEntry.OnSubmitted = func(string) { doLogin() }

	form := container.NewVBox(
		container.NewPadded(userEntry),
		container.NewPadded(passEntry),
		container.NewPadded(loginBtn),
		widget.NewLabelWithStyle("Default: admin / admin123", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
	)

	card := container.NewVBox(
		title,
		subtitle,
		widget.NewSeparator(),
		form,
	)

	centered := container.New(layout.NewCenterLayout(), container.NewPadded(card))

	aboutBtn := widget.NewButtonWithIcon("About", theme.InfoIcon(), func() {
		ShowAbout(w, version)
	})
	footer := container.NewHBox(layout.NewSpacer(), aboutBtn)

	w.SetContent(container.NewBorder(nil, container.NewPadded(footer), nil, nil, centered))
}
