package models

import "time"

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	FullName     string
	Role         string
	Active       bool
	CreatedAt    time.Time
	LastLogin    *time.Time
}

type Customer struct {
	ID        int64
	Name      string
	Phone     string
	Email     string
	Address   string
	Notes     string
	CreatedAt time.Time
}

type Repair struct {
	ID         int64
	CustomerID int64
	Device     string
	Serial     string
	Issue      string
	Status     string
	Priority   string
	AssignedTo *int64
	Cost       float64
	Notes      string
	CreatedAt  time.Time
	UpdatedAt  time.Time

	CustomerName string
	AssignedName string
}

type AuditEntry struct {
	ID        int64
	UserID    *int64
	Username  string
	Action    string
	Entity    string
	EntityID  *int64
	Details   string
	Timestamp time.Time
}

var RepairStatuses = []string{"Intake", "Diagnosing", "Awaiting Parts", "In Repair", "Ready for Pickup", "Completed", "Cancelled"}
var RepairPriorities = []string{"Low", "Normal", "High", "Urgent"}
var Roles = []string{"admin", "technician", "receptionist"}
