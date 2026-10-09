# HexDeskGo

A lightweight desktop application for managing a computer repair shop —
customers, repairs, technicians, users/roles, audit log, and database backup.

**Created by Joel R. Frank (Github @LegendaryKing87) — PrinTech Solutions**
## License
HexDeskGo is licensed under the GNU Affero General Public License v3.0. You are free to use, modify, and distribute it, provided that derivative works remain under the same license — including when offered as a network service.
## Features

- **User roles & permissions**: `admin`, `technician`, `receptionist`
- **Customers management** — add, edit, search, delete (admin only)
- **Repairs management** — full lifecycle with status/priority, technician assignment, cost tracking
- **Audit log** — every login, create, update, delete, backup, and purge is recorded
- **Backup** — one-click consistent SQLite backups (admin only) via `VACUUM INTO`
- **Cross-platform** desktop UI with Fyne

## Roles at a glance

| Action                  | Admin | Technician | Receptionist |
|-------------------------|:-----:|:----------:|:------------:|
| View dashboard/repairs  |   ✓   |     ✓      |      ✓       |
| Create/edit repairs     |   ✓   |     ✓      |      ✓       |
| Assign repairs          |   ✓   |     ✓      |              |
| Delete repairs          |   ✓   |            |              |
| Manage customers        |   ✓   |            |      ✓       |
| Delete customers        |   ✓   |            |              |
| Manage users            |   ✓   |            |              |
| View audit log          |   ✓   |            |              |
| Backup database         |   ✓   |            |              |

## Default login
Username: admin
Password: admin123


**Change this immediately after first login (Users tab → Edit → set new password).**

## Build & Run

Requirements: Go 1.21+, a C compiler (for `go-sqlite3`), and platform Fyne prerequisites
(see https://fyne.io/develop/).

```bash
go mod tidy
go run .
```
## Backups
Backups are created via the File → Backup Database… menu item and saved
as hexdesk_backup_YYYYMMDD_HHMMSS.db in the folder you choose.
The backup uses SQLite's VACUUM INTO, which is safe while the application is running.

## Data
The SQLite database (hexdesk.db) is created next to the executable on first run.
Audit entries older than 90 days can be purged from the Audit Log tab.


## Notes on the implementation

1. **User roles** — enforced in `auth.Session` via `Can*()` methods; every destructive action is guarded at the UI layer and the action is recorded in `audit_log`.
2. **Permissions** — role-based (`admin`, `technician`, `receptionist`), with guards for manage users, view/purge audit, backup, delete operations, and repair assignment.
3. **Audit log** — every login/logout, create/update/delete, backup, and purge writes a row with user, action, entity, entity id, and details.
4. **Backup** — uses `VACUUM INTO` which produces a consistent snapshot while the DB is open, with a file-copy fallback.
5. **About dialog** — appears from the menu, the About tab, and the login screen; contains the required credit line: *Created by Joel R. Frank (Github @LegendaryKing87) - PrinTech Solutions*.
6. **Default admin** — seeded as `admin` / `admin123`; change immediately.

### Setup

```bash
mkdir hexdeskgo && cd hexdeskgo
# create the files above in their subdirectories
go mod tidy
go run .
```

Depending on your OS you may need CGO enabled (CGO_ENABLED=1) and a C toolchain for go-sqlite3. If you prefer pure-Go SQLite, swap github.com/mattn/go-sqlite3 for modernc.org/sqlite and change the driver name from sqlite3 to sqlite in database/database.go