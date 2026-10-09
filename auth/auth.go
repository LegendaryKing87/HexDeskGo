package auth

import (
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"hexdeskgo/database"
	"hexdeskgo/models"
)

type Session struct {
	User *models.User
}

func (s *Session) HasRole(roles ...string) bool {
	if s.User == nil {
		return false
	}
	for _, r := range roles {
		if s.User.Role == r {
			return true
		}
	}
	return false
}

// Permission helpers
func (s *Session) CanManageUsers() bool     { return s.HasRole("admin") }
func (s *Session) CanViewAudit() bool       { return s.HasRole("admin") }
func (s *Session) CanBackup() bool          { return s.HasRole("admin") }
func (s *Session) CanDeleteRepairs() bool   { return s.HasRole("admin") }
func (s *Session) CanDeleteCustomers() bool { return s.HasRole("admin") }
func (s *Session) CanEditUsers() bool       { return s.HasRole("admin") }
func (s *Session) CanEditCustomers() bool   { return s.HasRole("admin", "receptionist") }
func (s *Session) CanEditRepairs() bool     { return s.HasRole("admin", "technician", "receptionist") }
func (s *Session) CanAssignRepairs() bool   { return s.HasRole("admin", "technician") }

func Authenticate(db *database.DB, username, password string) (*models.User, error) {
	var u models.User
	var lastLogin sql.NullTime
	err := db.QueryRow(`SELECT id, username, password_hash, full_name, role, active, created_at, last_login
        FROM users WHERE username = ?`, username).
		Scan(&u.ID, &u.Username, &u.PasswordHash, &u.FullName, &u.Role, &u.Active, &u.CreatedAt, &lastLogin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("invalid username or password")
		}
		return nil, err
	}
	if !u.Active {
		return nil, errors.New("account is disabled")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid username or password")
	}
	now := time.Now().UTC()
	_, _ = db.Exec("UPDATE users SET last_login = ? WHERE id = ?", now, u.ID)
	if lastLogin.Valid {
		u.LastLogin = &lastLogin.Time
	}
	db.LogAudit(u.ID, u.Username, "LOGIN", "user", u.ID, "User logged in")
	return &u, nil
}

func HashPassword(pw string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(h), err
}
