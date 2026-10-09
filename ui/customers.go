package ui

import (
	"database/sql"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"hexdeskgo/auth"
	"hexdeskgo/database"
	"hexdeskgo/models"
)

func makeCustomersTab(w fyne.Window, db *database.DB, session *auth.Session) fyne.CanvasObject {
	var data []models.Customer
	var selected *models.Customer

	search := widget.NewEntry()
	search.SetPlaceHolder("Search by name, phone, or email...")

	table := widget.NewTable(
		func() (int, int) { return len(data) + 1, 6 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, o fyne.CanvasObject) {
			l := o.(*widget.Label)
			if id.Row == 0 {
				l.TextStyle = fyne.TextStyle{Bold: true}
				l.SetText([]string{"ID", "Name", "Phone", "Email", "Address", "Created"}[id.Col])
				return
			}
			l.TextStyle = fyne.TextStyle{}
			c := data[id.Row-1]
			vals := []string{
				itoa(c.ID), c.Name, c.Phone, c.Email, c.Address,
				c.CreatedAt.Format("2006-01-02"),
			}
			l.SetText(vals[id.Col])
		},
	)
	table.SetColumnWidth(0, 60)
	table.SetColumnWidth(1, 180)
	table.SetColumnWidth(2, 140)
	table.SetColumnWidth(3, 220)
	table.SetColumnWidth(4, 260)
	table.SetColumnWidth(5, 130)

	refresh := func() {
		data = nil
		q := "SELECT id, name, phone, email, address, notes, created_at FROM customers"
		args := []interface{}{}
		if s := search.Text; s != "" {
			q += " WHERE name LIKE ? OR phone LIKE ? OR email LIKE ?"
			like := "%" + s + "%"
			args = append(args, like, like, like)
		}
		q += " ORDER BY id DESC"
		rows, err := db.Query(q, args...)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var c models.Customer
			_ = rows.Scan(&c.ID, &c.Name, &c.Phone, &c.Email, &c.Address, &c.Notes, &c.CreatedAt)
			data = append(data, c)
		}
		table.Refresh()
	}

	table.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 || id.Row-1 >= len(data) {
			return
		}
		c := data[id.Row-1]
		selected = &c
	}

	search.OnChanged = func(string) { refresh() }

	newBtn := widget.NewButtonWithIcon("New", theme.ContentAddIcon(), func() {
		if !session.CanEditCustomers() {
			dialog.ShowError(errPerm, w)
			return
		}
		showCustomerForm(w, db, session, nil, refresh)
	})
	editBtn := widget.NewButtonWithIcon("Edit", theme.DocumentCreateIcon(), func() {
		if selected == nil {
			dialog.ShowInformation("No selection", "Select a customer first.", w)
			return
		}
		if !session.CanEditCustomers() {
			dialog.ShowError(errPerm, w)
			return
		}
		showCustomerForm(w, db, session, selected, refresh)
	})
	delBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
		if selected == nil {
			dialog.ShowInformation("No selection", "Select a customer first.", w)
			return
		}
		if !session.CanDeleteCustomers() {
			dialog.ShowError(errPerm, w)
			return
		}
		dialog.ShowConfirm("Confirm Delete",
			"Delete customer "+selected.Name+" and all their repairs?",
			func(ok bool) {
				if !ok {
					return
				}
				if _, err := db.Exec("DELETE FROM customers WHERE id = ?", selected.ID); err != nil {
					dialog.ShowError(err, w)
					return
				}
				db.LogAudit(session.User.ID, session.User.Username, "DELETE", "customer", selected.ID, "Deleted customer "+selected.Name)
				selected = nil
				refresh()
			}, w)
	})
	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), refresh)

	toolbar := container.NewHBox(newBtn, editBtn, delBtn, widget.NewSeparator(), refreshBtn, search)

	refresh()
	return container.NewBorder(container.NewPadded(toolbar), nil, nil, nil, table)
}

func showCustomerForm(w fyne.Window, db *database.DB, session *auth.Session, existing *models.Customer, onSaved func()) {
	name := widget.NewEntry()
	phone := widget.NewEntry()
	email := widget.NewEntry()
	address := widget.NewEntry()
	notes := widget.NewMultiLineEntry()
	notes.SetMinRowsVisible(4)

	title := "New Customer"
	if existing != nil {
		title = "Edit Customer"
		name.SetText(existing.Name)
		phone.SetText(existing.Phone)
		email.SetText(existing.Email)
		address.SetText(existing.Address)
		notes.SetText(existing.Notes)
	}

	form := widget.NewForm(
		widget.NewFormItem("Name *", name),
		widget.NewFormItem("Phone", phone),
		widget.NewFormItem("Email", email),
		widget.NewFormItem("Address", address),
		widget.NewFormItem("Notes", notes),
	)

	d := dialog.NewCustomConfirm(title, "Save", "Cancel", form, func(ok bool) {
		if !ok {
			return
		}
		if name.Text == "" {
			dialog.ShowError(errRequired, w)
			return
		}
		if existing == nil {
			res, err := db.Exec(`INSERT INTO customers (name, phone, email, address, notes)
                VALUES (?, ?, ?, ?, ?)`,
				name.Text, phone.Text, email.Text, address.Text, notes.Text)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			id, _ := res.LastInsertId()
			db.LogAudit(session.User.ID, session.User.Username, "CREATE", "customer", id, "Created customer "+name.Text)
		} else {
			_, err := db.Exec(`UPDATE customers SET name=?, phone=?, email=?, address=?, notes=? WHERE id=?`,
				name.Text, phone.Text, email.Text, address.Text, notes.Text, existing.ID)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			db.LogAudit(session.User.ID, session.User.Username, "UPDATE", "customer", existing.ID, "Updated customer "+name.Text)
		}
		onSaved()
	}, w)
	d.Resize(fyne.NewSize(520, 460))
	d.Show()
}

// small utility used by scan targets
var _ = sql.ErrNoRows
