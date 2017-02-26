package db

import (
	"fmt"
)

type User struct {
	ID           string
	Name         string
	AuthToken    Token
	RefreshToken Token
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

	user := &User{}
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
