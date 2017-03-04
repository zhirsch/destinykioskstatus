package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	db *sql.DB

	selectUserStmt *sql.Stmt
	insertUserStmt *sql.Stmt
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
    ID                TEXT PRIMARY KEY,
    Name              TEXT,
    TokenAccessToken  TEXT,
    TokenRefreshToken TEXT,
    TokenExpiry       DATETIME
);
`
	if _, err := db.db.Exec(sql); err != nil {
		return nil, err
	}

	sql = `
SELECT
    ID,
    Name,
    TokenAccessToken,
    TokenRefreshToken,
    TokenExpiry
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
    TokenAccessToken,
    TokenRefreshToken,
    TokenExpiry
) VALUES(?, ?, ?, ?, ?);
`
	if stmt, err := db.db.Prepare(sql); err != nil {
		return nil, err
	} else {
		db.insertUserStmt = stmt
	}

	return db, nil
}
