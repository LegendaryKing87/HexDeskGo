package ui

import (
	"errors"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"hexdeskgo/auth"
	"hexdeskgo/backup"
	"hexdeskgo/database"
)

var (
	errPerm       = errors.New("you do not have permission to perform this action")
	errRequired   = errors.New("required fields missing")
	errSelfDelete = errors.New("you cannot delete your own account")
)

func itoa(i int64) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var b [24]byte
	pos := len(b)
	for i > 0 {
		pos--
		b[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		b[pos] = '-'
	}
	return string(b[pos:])
}

func parseTime(s string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("cannot parse time: " + s)
}

func doBackup(w fyne.Window, db *database.DB, session *auth.Session) {
	dirDlg := dialog.NewFolderOpen(func(lu fyne.ListableURI, err error) {
		if err != nil || lu == nil {
			return
		}
		dest, err := backup.Backup(db, lu.Path())
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		db.LogAudit(session.User.ID, session.User.Username, "BACKUP", "database", 0, "Backup created at "+dest)
		dialog.ShowInformation("Backup Complete", "Backup saved to:\n"+dest, w)
	}, w)
	dirDlg.Show()
}

func makeDashboard(w fyne.Window, db *database.DB, session *auth.Session) fyne.CanvasObject {
	stat := func(label, query string) fyne.CanvasObject {
		var n int
		_ = db.QueryRow(query).Scan(&n)
		l := widget.NewLabelWithStyle(itoa(int64(n)), fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Monospace: true})
		return container.NewVBox(
			widget.NewLabelWithStyle(label, fyne.TextAlignCenter, fyne.TextStyle{}),
			l,
		)
	}

	cards := container.NewGridWithColumns(4,
		wrapCard(stat("Total Customers", "SELECT COUNT(*) FROM customers")),
		wrapCard(stat("Open Repairs", "SELECT COUNT(*) FROM repairs WHERE status NOT IN ('Completed','Cancelled')")),
		wrapCard(stat("Completed", "SELECT COUNT(*) FROM repairs WHERE status = 'Completed'")),
		wrapCard(stat("Active Users", "SELECT COUNT(*) FROM users WHERE active = 1")),
	)

	welcome := widget.NewLabelWithStyle("Welcome, "+session.User.FullName+"!",
		fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	quick := widget.NewLabelWithStyle(
		"Use the tabs above to manage repairs and customers.\n"+
			"Administrators have access to Users, Audit Log, and Backup.",
		fyne.TextAlignCenter, fyne.TextStyle{})

	return container.New(layout.NewCenterLayout(),
		container.NewVBox(
			welcome,
			widget.NewSeparator(),
			cards,
			widget.NewSeparator(),
			quick,
		))
}

func wrapCard(inner fyne.CanvasObject) fyne.CanvasObject {
	return container.NewPadded(container.NewVBox(
		widget.NewCard("", "", inner),
	))
}

// keep imports used when building variants
var _ = theme.HomeIcon
