package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"hexdeskgo/auth"
	"hexdeskgo/database"
)

func ShowMain(w fyne.Window, db *database.DB, session *auth.Session, version string) {
	w.SetMainMenu(buildMenu(w, db, session, version))

	header := widget.NewLabelWithStyle(
		"HexDeskGo  •  Logged in as "+session.User.FullName+" ("+session.User.Role+")",
		fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Dashboard", theme.HomeIcon(), makeDashboard(w, db, session)),
		container.NewTabItemWithIcon("Repairs", theme.ContentCopyIcon(), makeRepairsTab(w, db, session)),
		container.NewTabItemWithIcon("Customers", theme.AccountIcon(), makeCustomersTab(w, db, session)),
	)
	if session.CanManageUsers() {
		tabs.Append(container.NewTabItemWithIcon("Users", theme.SettingsIcon(), makeUsersTab(w, db, session)))
	}
	if session.CanViewAudit() {
		tabs.Append(container.NewTabItemWithIcon("Audit Log", theme.DocumentIcon(), makeAuditTab(w, db, session)))
	}
	tabs.Append(container.NewTabItemWithIcon("About", theme.InfoIcon(), makeAboutTab(w, version)))

	content := container.NewBorder(container.NewPadded(header), nil, nil, nil, tabs)
	w.SetContent(content)
}

func buildMenu(w fyne.Window, db *database.DB, session *auth.Session, version string) *fyne.MainMenu {
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Backup Database...", func() {
			if !session.CanBackup() {
				dialog.ShowError(errPerm, w)
				return
			}
			doBackup(w, db, session)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Log Out", func() {
			db.LogAudit(session.User.ID, session.User.Username, "LOGOUT", "user", session.User.ID, "User logged out")
			ShowLogin(w, db, version)
		}),
		fyne.NewMenuItem("Exit", func() { w.Close() }),
	)

	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("About HexDeskGo", func() { ShowAbout(w, version) }),
	)

	return fyne.NewMainMenu(fileMenu, helpMenu)
}
