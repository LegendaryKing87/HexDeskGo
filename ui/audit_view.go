package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"hexdeskgo/audit"
	"hexdeskgo/auth"
	"hexdeskgo/database"
	"hexdeskgo/models"
)

func makeAuditTab(w fyne.Window, db *database.DB, session *auth.Session) fyne.CanvasObject {
	var data []models.AuditEntry

	table := widget.NewTable(
		func() (int, int) { return len(data) + 1, 7 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, o fyne.CanvasObject) {
			l := o.(*widget.Label)
			if id.Row == 0 {
				l.TextStyle = fyne.TextStyle{Bold: true}
				l.SetText([]string{"Time", "User", "Action", "Entity", "ID", "Details", ""}[id.Col])
				return
			}
			e := data[id.Row-1]
			eid := ""
			if e.EntityID != nil {
				eid = itoa(*e.EntityID)
			}
			l.TextStyle = fyne.TextStyle{Monospace: true}
			l.SetText([]string{
				e.Timestamp.Format("2006-01-02 15:04:05"),
				e.Username, e.Action, e.Entity, eid, e.Details, "",
			}[id.Col])
		},
	)
	for i, wdt := range []float32{170, 130, 130, 100, 60, 420, 20} {
		table.SetColumnWidth(i, wdt)
	}

	refresh := func() {
		data = nil
		rows, err := audit.List(db, 500)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		data = rows
		table.Refresh()
	}

	refreshBtn := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), refresh)
	purgeBtn := widget.NewButtonWithIcon("Purge older than 90 days", theme.DeleteIcon(), func() {
		dialog.ShowConfirm("Confirm Purge", "Delete audit entries older than 90 days?", func(ok bool) {
			if !ok {
				return
			}
			n, err := audit.Purge(db, 90)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			db.LogAudit(session.User.ID, session.User.Username, "PURGE", "audit_log", 0,
				fmt.Sprintf("Purged %d audit entries older than 90 days", n))
			refresh()
		}, w)
	})

	toolbar := container.NewHBox(refreshBtn, purgeBtn,
		widget.NewLabel("Showing last 500 entries"))

	refresh()
	return container.NewBorder(container.NewPadded(toolbar), nil, nil, nil, table)
}
