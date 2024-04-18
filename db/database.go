package db

import (
	"database/sql"
	"errors"
	"log/slog"
	"strconv"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var DB *wrapped

//go:embed install.sql
var installSql string

var currentVersion = 2

func Init(path string) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		slog.Error("open database", "err", err)
		panic(err)
	}

	DB = newWrapped(db)

	_, err = DB.Exec(installSql)
	if err != nil {
		slog.Error("create tables", "err", err)
		panic(err)
	}

	// migration
	migrate()

	// create admin account if not exists
	var count int
	row := DB.QueryRow("SELECT count(id) FROM users WHERE role = ? LIMIT 1", 0)
	err = row.Scan(&count)
	if err != nil {
		slog.Error("insert default values", "err", err)
		panic(err)
	}
	if count == 0 {
		username := "admin"
		password := "admin"
		hashed, err := bcrypt.GenerateFromPassword([]byte(password), 0)
		if err != nil {
			slog.Error("bcrypt hash", "err", err)
			panic(err)
		}
		_, err = DB.Exec("INSERT INTO users(username, email, password, email_verified, role) VALUES (?, ?, ?, ?, ?)", username, "admin@example.com", string(hashed), true, 0)
		if err != nil {
			slog.Error("insert default values", "err", err)
			panic(err)
		}
		slog.Warn("admin account created", "email", "admin@example.com", "password", password)
	}

	// create a local storage driver if no driver exists
	row = DB.QueryRow("SELECT count(id) FROM storages")
	err = row.Scan(&count)
	if err != nil {
		slog.Error("insert default values", "err", err)
		panic(err)
	}

	if count == 0 {
		_, err = StorageCreate("local storage", "local", "{\"path\": \"./uploads\"}", true, true)
		if err != nil {
			slog.Error("create default storage", "err", err)
			panic(err)
		}
	}

	slog.Info("database initalized")

}

func migrate() {
	slog.Debug("database migration start")

	// check installed version
	versionString := ""
	row := DB.QueryRow("SELECT value FROM settings WHERE key = 'VERSION'")
	if errors.Is(row.Scan(&versionString), sql.ErrNoRows) {
		versionString = "1"

		_, err := DB.Exec("INSERT OR IGNORE INTO settings(key, value) VALUES('VERSION', '1')")
		if err != nil {
			slog.Error("database migration", "err", err)
		}
	}

	installedVersion, err := strconv.Atoi(versionString)
	if err != nil {
		panic("invalid database version: " + versionString)
	}

	slog.Debug("database migration", "installed version", installedVersion)

	// do the migration and update database version
	doMigration := func(from int, to int, s string) {
		if installedVersion != from {
			slog.Debug("skip migration", "from", from, "to", to, "installed version", installedVersion)
			return
		}

		slog.Debug("do migration", "from", from, "to", to, "installed version", installedVersion)

		_, err := DB.Exec(s)
		if err != nil {
			slog.Error("database migration", "err", err)
		}

		slog.Debug("do migration finished", "from", from, "to", to)

		// update version
		installedVersion = to
		_, err = DB.Exec("UPDATE settings SET value = ? WHERE key = ?", strconv.Itoa(to), "VERSION")
		if err != nil {
			slog.Error("database migration", "err", err)
		}
	}

	// add internal file name
	doMigration(1, 2, `
		ALTER TABLE images ADD internal_name TEXT NOT NULL DEFAULT '';
		UPDATE images SET internal_name = file_name WHERE internal_name = '';
	`)

	// add user groups
	doMigration(2, 3, `
		CREATE TABLE IF NOT EXISTS groups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			allow_upload BOOLEAN NOT NULL,
			max_file_size INTEGER NOT NULL,
			upload_per_minute INTEGER NOT NULL,
			upload_per_hour INTEGER NOT NULL,
			upload_per_day INTEGER NOT NULL,
			upload_per_month INTEGER NOT NULL,
			total_uploads INTEGER NOT NULL,
			max_retention_seconds INTEGER NOT NULL
		);

		INSERT OR IGNORE INTO groups(id, name, max_file_size, upload_per_minute, upload_per_hour, upload_per_day, upload_per_month, total_uploads, allow_upload, max_retention_seconds) VALUES(0, 'Default user group', 16000000, 30, 100, 1000, 1000, 10000, TRUE, 0);

		ALTER TABLE users DROP space;
		ALTER TABLE users ADD user_group INTEGER NOT NULL DEFAULT 0 REFERENCES groups(id);
		ALTER TABLE users ADD user_group_expire INTEGER NOT NULL DEFAULT 0;

		DELETE FROM settings WHERE key = 'MAX_IMAGE_SIZE';
		DELETE FROM settings WHERE key = 'GUEST_UPLOAD';
		DELETE FROM settings WHERE key = 'GUEST_MAX_TIME';
		DELETE FROM settings WHERE key = 'USER_MAX_TIME';
	`)

	slog.Debug("database migration done")
}
