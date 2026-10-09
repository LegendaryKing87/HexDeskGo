package audit

import (
	"hexdeskgo/database"
	"hexdeskgo/models"
)

func List(db *database.DB, limit int) ([]models.AuditEntry, error) {
	if limit <= 0 {
		limit = 500
	}
	rows, err := db.Query(`SELECT id, user_id, username, action, entity, entity_id, details, timestamp
        FROM audit_log ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.AuditEntry
	for rows.Next() {
		var e models.AuditEntry
		var uid, eid *int64
		if err := rows.Scan(&e.ID, &uid, &e.Username, &e.Action, &e.Entity, &eid, &e.Details, &e.Timestamp); err != nil {
			return nil, err
		}
		e.UserID = uid
		e.EntityID = eid
		out = append(out, e)
	}
	return out, rows.Err()
}

func Purge(db *database.DB, days int) (int64, error) {
	res, err := db.Exec("DELETE FROM audit_log WHERE timestamp < datetime('now', ?)", "-"+itoa(days)+" days")
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var b [20]byte
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
