package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"hexdeskgo/auth"
	"hexdeskgo/database"
	"hexdeskgo/models"
)

func makeUsersTab(w fyne.Window, db *database.DB, session *auth.Session) fyne.CanvasObject {
	var data []models.User
	var selected *models.User

	table := widget.NewTable(
		func() (int, int) { return len(data) + 1, 6 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, o fyne.CanvasObject) {
			l := o.(*widget.Label)
			if id.Row == 0 {
				l.TextStyle = fyne.TextStyle{Bold: true}
				l.SetText([]string{"ID", "Username", "Full Name", "Role", "Active", "Last Login"}[id.Col])
				return
			}
			u := data[id.Row-1]
			active := "No"
			if u.Active {
				active = "Yes"
			}
			ll := ""
			if u.LastLogin != nil {
				ll = u.LastLogin.Format("2006-01-02 15:04")
			}
			l.TextStyle = fyne.TextStyle{}
			l.SetText([]string{itoa(u.ID), u.Username, u.FullName, u.Role, active, ll}[id.Col])
		},
	)
	for i, wdt := range []float32{60, 150, 200, 130, 80, 160} {
		table.SetColumnWidth(i, wdt)
	}

	refresh := func() {
		data = nil
		rows, err := db.Query("SELECT id, username, password_hash, full_name, role, active, created_at, last_login FROM users ORDER BY id")
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var u models.User
			var last *string
			_ = rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.FullName, &u.Role, &u.Active, &u.CreatedAt, &last)
			if last != nil {
				t, _ := parseTime(*last)
				u.LastLogin = &t
			}
			data = append(data, u)
		}
		table.Refresh()
	}

	table.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 || id.Row-1 >= len(data) {
			return
		}
		u := data[id.Row-1]
		selected = &u
	}

	newBtn := widget.NewButtonWithIcon("New User", theme.ContentAddIcon(), func() {
		showUserForm(w, db, session, nil, refresh)
	})
	editBtn := widget.NewButtonWithIcon("Edit", theme.DocumentCreateIcon(), func() {
		if selected == nil {
			dialog.ShowInformation("No selection", "Select a user first.", w)
			return
		}
		showUserForm(w, db, session, selected, refresh)
	})
	delBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
		if selected == nil {
			dialog.ShowInformation("No selection", "Select a user first.", w)
			return
		}
		if selected.ID == session.User.ID {
			dialog.ShowError(errSelfDelete, w)
			return
		}
		dialog.ShowConfirm("Confirm Delete", "Delete user "+selected.Username+"?", func(ok bool) {
			if !ok {
				return
			}
			_, _ = db.Exec("DELETE FROM users WHERE id = ?", selected.ID)
			db.LogAudit(session.User.ID, session.User.Username, "DELETE", "user", selected.ID, "Deleted user "+selected.Username)
			selected = nil
			refresh()
		}, w)
	})
	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), refresh)

	toolbar := container.NewHBox(newBtn, editBtn, delBtn, widget.NewSeparator(), refreshBtn)
	refresh()
	return container.NewBorder(container.NewPadded(toolbar), nil, nil, nil, table)
}

func showUserForm(w fyne.Window, db *database.DB, session *auth.Session, existing *models.User, onSaved func()) {
	username := widget.NewEntry()
	fullName := widget.NewEntry()
	role := widget.NewSelect(models.Roles, nil)
	password := widget.NewPasswordEntry()
	active := widget.NewCheck("Active", nil)

	title := "New User"
	if existing != nil {
		title = "Edit User"
		username.SetText(existing.Username)
		fullName.SetText(existing.FullName)
		role.SetSelected(existing.Role)
		active.SetChecked(existing.Active)
		password.SetPlaceHolder("(leave blank to keep unchanged)")
	} else {
		role.SetSelected("technician")
		active.SetChecked(true)
	}

	form := widget.NewForm(
		widget.NewFormItem("Username *", username),
		widget.NewFormItem("Full Name *", fullName),
		widget.NewFormItem("Role *", role),
		widget.NewFormItem("Password *", password),
		widget.NewFormItem("", active),
	)

	d := dialog.NewCustomConfirm(title, "Save", "Cancel", form, func(ok bool) {
		if !ok {
			return
		}
		if username.Text == "" || fullName.Text == "" || role.Selected == "" {
			dialog.ShowError(errRequired, w)
			return
		}
		if existing == nil {
			if password.Text == "" {
				dialog.ShowError(errRequired, w)
				return
			}
			hash, err := auth.HashPassword(password.Text)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			res, err := db.Exec(`INSERT INTO users (username, password_hash, full_name, role, active)
                VALUES (?, ?, ?, ?, ?)`,
				username.Text, hash, fullName.Text, role.Selected, active.Checked)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			id, _ := res.LastInsertId()
			db.LogAudit(session.User.ID, session.User.Username, "CREATE", "user", id,
				"Created user "+username.Text+" ("+role.Selected+")")
		} else {
			if password.Text != "" {
				hash, err := auth.HashPassword(password.Text)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				_, err = db.Exec(`UPDATE users SET username=?, full_name=?, role=?, active=?, password_hash=? WHERE id=?`,
					username.Text, fullName.Text, role.Selected, active.Checked, hash, existing.ID)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				db.LogAudit(session.User.ID, session.User.Username, "PASSWORD_RESET", "user", existing.ID,
					"Changed password for "+username.Text)
			} else {
				_, err := db.Exec(`UPDATE users SET username=?, full_name=?, role=?, active=? WHERE id=?`,
					username.Text, fullName.Text, role.Selected, active.Checked, existing.ID)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
			}
			db.LogAudit(session.User.ID, session.User.Username, "UPDATE", "user", existing.ID, "Updated user "+username.Text)
		}
		onSaved()
	}, w)
	d.Resize(fyne.NewSize(480, 420))
	d.Show()
}
