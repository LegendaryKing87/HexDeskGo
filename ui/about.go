package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const aboutText = `HexDeskGo
Computer Repair Shop Management System

A lightweight desktop application for managing
customers, repairs, technicians, and shop operations.

Created by Joel R. Frank (Github @LegendaryKing87)
PrinTech Solutions
`

func ShowAbout(w fyne.Window, version string) {
	title := canvas.NewText("HexDeskGo", color.NRGBA{R: 0x2E, G: 0x86, B: 0xAB, A: 0xFF})
	title.TextSize = 32
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}

	ver := widget.NewLabelWithStyle("Version "+version, fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	body := widget.NewLabelWithStyle(aboutText, fyne.TextAlignCenter, fyne.TextStyle{})

	content := container.NewVBox(
		container.NewPadded(title),
		ver,
		widget.NewSeparator(),
		body,
	)

	d := dialog.NewCustom("About HexDeskGo", "Close",
		container.New(layout.NewCenterLayout(), container.NewPadded(content)), w)
	d.Resize(fyne.NewSize(440, 320))
	d.Show()
}

func makeAboutTab(w fyne.Window, version string) fyne.CanvasObject {
	title := canvas.NewText("HexDeskGo", color.NRGBA{R: 0x2E, G: 0x86, B: 0xAB, A: 0xFF})
	title.TextSize = 40
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}

	ver := widget.NewLabelWithStyle("Version "+version, fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	credits := widget.NewLabelWithStyle(
		"Created by Joel R. Frank (Github @LegendaryKing87)\nPrinTech Solutions",
		fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	details := widget.NewLabelWithStyle(
		"A lightweight desktop application for managing customers, repairs,\n"+
			"technicians, and shop operations.\n\n"+
			"Built with Go, Fyne, and SQLite. \n\n"+
			"HexDeskGo is licensed under the GNU Affero General Public License v3.0. AGPL v3.0",
		fyne.TextAlignCenter, fyne.TextStyle{})

	btn := widget.NewButton("Show About Dialog", func() { ShowAbout(w, version) })

	return container.New(layout.NewCenterLayout(), container.NewVBox(
		title, ver, widget.NewSeparator(), credits, widget.NewSeparator(), details,
		container.NewCenter(btn),
	))
}
