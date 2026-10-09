package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	*sql.DB
}

func Init(path string) (*DB, error) {
	conn, err := sql.Open("sqlite3", path+"?_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	db := &DB{conn}
	if err := db.migrate(); err != nil {
		return nil, err
	}
	if err := db.seed(); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DB) migrate() error {
	schema := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT UNIQUE NOT NULL,
        password_hash TEXT NOT NULL,
        full_name TEXT NOT NULL,
        role TEXT NOT NULL CHECK(role IN ('admin','technician','receptionist')),
        active INTEGER NOT NULL DEFAULT 1,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        last_login DATETIME
    );

    CREATE TABLE IF NOT EXISTS customers (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        phone TEXT,
        email TEXT,
        address TEXT,
        notes TEXT,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS repairs (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        customer_id INTEGER NOT NULL,
        device TEXT NOT NULL,
        serial TEXT,
        issue TEXT NOT NULL,
        status TEXT NOT NULL DEFAULT 'Intake',
        priority TEXT NOT NULL DEFAULT 'Normal',
        assigned_to INTEGER,
        cost REAL NOT NULL DEFAULT 0,
        notes TEXT,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE,
        FOREIGN KEY (assigned_to) REFERENCES users(id) ON DELETE SET NULL
    );

    CREATE TABLE IF NOT EXISTS audit_log (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER,
        username TEXT,
        action TEXT NOT NULL,
        entity TEXT,
        entity_id INTEGER,
        details TEXT,
        timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
    );

    CREATE INDEX IF NOT EXISTS idx_audit_ts ON audit_log(timestamp);
    CREATE INDEX IF NOT EXISTS idx_repairs_status ON repairs(status);
    `
	_, err := db.Exec(schema)
	return err
}

func (db *DB) seed() error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO users (username, password_hash, full_name, role) VALUES (?, ?, ?, ?)`,
		"admin", string(hash), "Administrator", "admin")
	if err != nil {
		return err
	}
	fmt.Println("Seeded default admin user: admin / admin123 - PLEASE CHANGE PASSWORD")
	return nil
}

func (db *DB) LogAudit(userID int64, username, action, entity string, entityID int64, details string) {
	_, _ = db.Exec(`INSERT INTO audit_log (user_id, username, action, entity, entity_id, details, timestamp)
        VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, username, action, entity, entityID, details, time.Now().UTC())
}

// append to database/database.go
func (db *DB) Path() string {
	var p string
	_ = db.QueryRow("PRAGMA database_list").Scan(new(int), new(string), &p)
	return p
}
