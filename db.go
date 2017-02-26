package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zhirsch/destinykioskstatus/api"
)

type DB struct {
	db *sql.DB

	selectUserStmt *sql.Stmt
	insertUserStmt *sql.Stmt
}

type User struct {
	ID           string
	Name         string
	AuthToken    *api.Token
	RefreshToken *api.Token
}

func NewDB(path string) (*DB, error) {
	sqldb, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	} else if sqldb == nil {
		return nil, fmt.Errorf("db is nil")
	}
	db := &DB{db: sqldb}

	sql := `
CREATE TABLE IF NOT EXISTS Users(
    ID                  TEXT PRIMARY KEY,
    Name                TEXT,
    AuthTokenValue      TEXT,
    AuthTokenReady      DATETIME,
    AuthTokenExpires    DATETIME,
    RefreshTokenValue   TEXT,
    RefreshTokenReady   DATETIME,
    RefreshTokenExpires DATETIME
);
`
	if _, err := db.db.Exec(sql); err != nil {
		return nil, err
	}

	sql = `
SELECT
    ID,
    Name,
    AuthTokenValue,
    AuthTokenReady,
    AuthTokenExpires,
    RefreshTokenValue,
    RefreshTokenReady,
    RefreshTokenExpires
FROM
    Users
WHERE
    ID = ?;
`
	if stmt, err := db.db.Prepare(sql); err != nil {
		return nil, err
	} else {
		db.selectUserStmt = stmt
	}

	sql = `
INSERT OR REPLACE INTO Users(
    ID,
    Name,
    AuthTokenValue,
    AuthTokenReady,
    AuthTokenExpires,
    RefreshTokenValue,
    RefreshTokenReady,
    RefreshTokenExpires
) VALUES(?, ?, ?, ?, ?, ?, ?, ?);
`
	if stmt, err := db.db.Prepare(sql); err != nil {
		return nil, err
	} else {
		db.insertUserStmt = stmt
	}

	return db, nil
}

func (db *DB) SelectUser(id string) (*User, error) {
	rows, err := db.selectUserStmt.Query(id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("no rows for %v", id)
	}

	user := &User{
		AuthToken:    &api.Token{},
		RefreshToken: &api.Token{},
	}
	err = rows.Scan(
		&user.ID,
		&user.Name,
		&user.AuthToken.Value,
		&user.AuthToken.Ready,
		&user.AuthToken.Expires,
		&user.RefreshToken.Value,
		&user.RefreshToken.Ready,
		&user.RefreshToken.Expires,
	)
	if err != nil {
		return nil, err
	}

	if rows.Next() {
		return nil, fmt.Errorf("too many rows for %v", id)
	}

	return user, nil
}

func (db *DB) InsertUser(user *User) error {
	_, err := db.insertUserStmt.Exec(
		user.ID,
		user.Name,
		user.AuthToken.Value,
		user.AuthToken.Ready,
		user.AuthToken.Expires,
		user.RefreshToken.Value,
		user.RefreshToken.Ready,
		user.RefreshToken.Expires,
	)
	return err
}
