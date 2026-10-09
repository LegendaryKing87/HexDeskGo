package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"hexdeskgo/database"
)

// Backup creates a consistent copy of the SQLite database file using
// SQLite's VACUUM INTO which is safe while the DB is open.
func Backup(db *database.DB, destDir string) (string, error) {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", err
	}
	name := fmt.Sprintf("hexdesk_backup_%s.db", time.Now().Format("20060102_150405"))
	dest := filepath.Join(destDir, name)

	// VACUUM INTO produces a compact, consistent backup.
	if _, err := db.Exec("VACUUM INTO ?", dest); err != nil {
		// Fallback: plain copy (requires no writer active). Try anyway.
		if cerr := copyFile(db.Path(), dest); cerr != nil {
			return "", fmt.Errorf("vacuum failed: %v; copy failed: %v", err, cerr)
		}
	}
	return dest, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
