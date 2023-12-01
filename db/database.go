package db

import (
	"database/sql"
	"log/slog"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var DB *wrapped

//go:embed install.sql
var installSql string

func init() {
	db, err := sql.Open("sqlite3", "./db.sqlite")
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
