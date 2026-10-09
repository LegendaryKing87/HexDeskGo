package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"hexdeskgo/auth"
	"hexdeskgo/database"
	"hexdeskgo/models"
)

func makeRepairsTab(w fyne.Window, db *database.DB, session *auth.Session) fyne.CanvasObject {
	var data []models.Repair
	var selected *models.Repair

	search := widget.NewEntry()
	search.SetPlaceHolder("Search device, issue, or customer...")

	statusFilter := widget.NewSelect(append([]string{"All"}, models.RepairStatuses...), nil)
	statusFilter.SetSelected("All")

	table := widget.NewTable(
		func() (int, int) { return len(data) + 1, 9 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, o fyne.CanvasObject) {
			l := o.(*widget.Label)
			if id.Row == 0 {
				l.TextStyle = fyne.TextStyle{Bold: true}
				l.SetText([]string{"ID", "Customer", "Device", "Issue", "Status", "Priority", "Assigned", "Cost", "Updated"}[id.Col])
				return
			}
			r := data[id.Row-1]
			vals := []string{
				itoa(r.ID), r.CustomerName, r.Device, r.Issue, r.Status, r.Priority,
				r.AssignedName, fmt.Sprintf("$%.2f", r.Cost), r.UpdatedAt.Format("2006-01-02 15:04"),
			}
			l.TextStyle = fyne.TextStyle{}
			l.SetText(vals[id.Col])
		},
	)
	for i, wdt := range []float32{60, 160, 160, 280, 130, 90, 130, 90, 150} {
		table.SetColumnWidth(i, wdt)
	}

	refresh := func() {
		data = nil
		q := `SELECT r.id, r.customer_id, r.device, COALESCE(r.serial,''), r.issue, r.status, r.priority,
                     r.assigned_to, r.cost, COALESCE(r.notes,''), r.created_at, r.updated_at,
                     c.name, COALESCE(u.full_name,'')
              FROM repairs r
              JOIN customers c ON c.id = r.customer_id
              LEFT JOIN users u ON u.id = r.assigned_to
              WHERE 1=1`
		args := []interface{}{}
		if s := search.Text; s != "" {
			q += " AND (r.device LIKE ? OR r.issue LIKE ? OR c.name LIKE ?)"
			like := "%" + s + "%"
			args = append(args, like, like, like)
		}
		if statusFilter.Selected != "" && statusFilter.Selected != "All" {
			q += " AND r.status = ?"
			args = append(args, statusFilter.Selected)
		}
		q += " ORDER BY r.id DESC"

		rows, err := db.Query(q, args...)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var r models.Repair
			var aid *int64
			_ = rows.Scan(&r.ID, &r.CustomerID, &r.Device, &r.Serial, &r.Issue, &r.Status, &r.Priority,
				&aid, &r.Cost, &r.Notes, &r.CreatedAt, &r.UpdatedAt, &r.CustomerName, &r.AssignedName)
			r.AssignedTo = aid
			data = append(data, r)
		}
		table.Refresh()
	}

	table.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 || id.Row-1 >= len(data) {
			return
		}
		r := data[id.Row-1]
		selected = &r
	}

	search.OnChanged = func(string) { refresh() }
	statusFilter.OnChanged = func(string) { refresh() }

	newBtn := widget.NewButtonWithIcon("New Repair", theme.ContentAddIcon(), func() {
		if !session.CanEditRepairs() {
			dialog.ShowError(errPerm, w)
			return
		}
		showRepairForm(w, db, session, nil, refresh)
	})
	editBtn := widget.NewButtonWithIcon("Edit", theme.DocumentCreateIcon(), func() {
		if selected == nil {
			dialog.ShowInformation("No selection", "Select a repair first.", w)
			return
		}
		if !session.CanEditRepairs() {
			dialog.ShowError(errPerm, w)
			return
		}
		showRepairForm(w, db, session, selected, refresh)
	})
	delBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
		if selected == nil {
			dialog.ShowInformation("No selection", "Select a repair first.", w)
			return
		}
		if !session.CanDeleteRepairs() {
			dialog.ShowError(errPerm, w)
			return
		}
		dialog.ShowConfirm("Confirm Delete", "Delete repair #"+itoa(selected.ID)+"?", func(ok bool) {
			if !ok {
				return
			}
			_, _ = db.Exec("DELETE FROM repairs WHERE id = ?", selected.ID)
			db.LogAudit(session.User.ID, session.User.Username, "DELETE", "repair", selected.ID,
				"Deleted repair for "+selected.CustomerName)
			selected = nil
			refresh()
		}, w)
	})
	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), refresh)

	toolbar := container.NewHBox(newBtn, editBtn, delBtn, widget.NewSeparator(),
		widget.NewLabel("Status:"), statusFilter, refreshBtn, search)

	refresh()
	return container.NewBorder(container.NewPadded(toolbar), nil, nil, nil, table)
}

func showRepairForm(w fyne.Window, db *database.DB, session *auth.Session, existing *models.Repair, onSaved func()) {
	// Load customers
	var custNames []string
	var custIDs []int64
	rows, err := db.Query("SELECT id, name FROM customers ORDER BY name")
	if err != nil {
		dialog.ShowError(err, w)
		return
	}
	for rows.Next() {
		var id int64
		var name string
		_ = rows.Scan(&id, &name)
		custIDs = append(custIDs, id)
		custNames = append(custNames, name)
	}
	rows.Close()
	if len(custIDs) == 0 {
		dialog.ShowInformation("No customers", "Add a customer first before creating a repair.", w)
		return
	}

	// Load technicians/admins for assignment
	techNames := []string{"(unassigned)"}
	techIDs := []int64{0}
	if session.CanAssignRepairs() {
		tRows, _ := db.Query("SELECT id, full_name FROM users WHERE active = 1 AND role IN ('admin','technician') ORDER BY full_name")
		if tRows != nil {
			for tRows.Next() {
				var id int64
				var n string
				_ = tRows.Scan(&id, &n)
				techIDs = append(techIDs, id)
				techNames = append(techNames, n)
			}
			tRows.Close()
		}
	}

	customerSel := widget.NewSelect(custNames, nil)
	device := widget.NewEntry()
	serial := widget.NewEntry()
	issue := widget.NewMultiLineEntry()
	issue.SetMinRowsVisible(3)
	statusSel := widget.NewSelect(models.RepairStatuses, nil)
	prioritySel := widget.NewSelect(models.RepairPriorities, nil)
	assignSel := widget.NewSelect(techNames, nil)
	cost := widget.NewEntry()
	notes := widget.NewMultiLineEntry()
	notes.SetMinRowsVisible(3)

	title := "New Repair"
	if existing != nil {
		title = "Edit Repair #" + itoa(existing.ID)
		for i, id := range custIDs {
			if id == existing.CustomerID {
				customerSel.SetSelected(custNames[i])
				break
			}
		}
		device.SetText(existing.Device)
		serial.SetText(existing.Serial)
		issue.SetText(existing.Issue)
		statusSel.SetSelected(existing.Status)
		prioritySel.SetSelected(existing.Priority)
		for i, id := range techIDs {
			if existing.AssignedTo != nil && id == *existing.AssignedTo {
				assignSel.SetSelected(techNames[i])
				break
			}
		}
		cost.SetText(fmt.Sprintf("%.2f", existing.Cost))
		notes.SetText(existing.Notes)
	} else {
		customerSel.SetSelected(custNames[0])
		statusSel.SetSelected("Intake")
		prioritySel.SetSelected("Normal")
		assignSel.SetSelected(techNames[0])
		cost.SetText("0.00")
	}

	form := widget.NewForm(
		widget.NewFormItem("Customer *", customerSel),
		widget.NewFormItem("Device *", device),
		widget.NewFormItem("Serial", serial),
		widget.NewFormItem("Issue *", issue),
		widget.NewFormItem("Status", statusSel),
		widget.NewFormItem("Priority", prioritySel),
		widget.NewFormItem("Assigned To", assignSel),
		widget.NewFormItem("Cost", cost),
		widget.NewFormItem("Notes", notes),
	)

	d := dialog.NewCustomConfirm(title, "Save", "Cancel", container.NewVScroll(form), func(ok bool) {
		if !ok {
			return
		}
		if device.Text == "" || issue.Text == "" || customerSel.Selected == "" {
			dialog.ShowError(errRequired, w)
			return
		}
		var custID int64
		for i, n := range custNames {
			if n == customerSel.Selected {
				custID = custIDs[i]
				break
			}
		}
		var assignID *int64
		for i, n := range techNames {
			if n == assignSel.Selected {
				if techIDs[i] != 0 {
					v := techIDs[i]
					assignID = &v
				}
				break
			}
		}
		var costVal float64
		fmt.Sscanf(cost.Text, "%f", &costVal)

		if existing == nil {
			res, err := db.Exec(`INSERT INTO repairs
                (customer_id, device, serial, issue, status, priority, assigned_to, cost, notes)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				custID, device.Text, serial.Text, issue.Text, statusSel.Selected,
				prioritySel.Selected, assignID, costVal, notes.Text)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			id, _ := res.LastInsertId()
			db.LogAudit(session.User.ID, session.User.Username, "CREATE", "repair", id,
				"Created repair for "+customerSel.Selected+" device "+device.Text)
		} else {
			_, err := db.Exec(`UPDATE repairs SET customer_id=?, device=?, serial=?, issue=?, status=?,
                priority=?, assigned_to=?, cost=?, notes=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
				custID, device.Text, serial.Text, issue.Text, statusSel.Selected,
				prioritySel.Selected, assignID, costVal, notes.Text, existing.ID)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			db.LogAudit(session.User.ID, session.User.Username, "UPDATE", "repair", existing.ID,
				"Updated repair for "+customerSel.Selected)
		}
		onSaved()
	}, w)
	d.Resize(fyne.NewSize(600, 650))
	d.Show()
}
