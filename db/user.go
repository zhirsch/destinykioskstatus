package db

import (
	"fmt"

	"github.com/zhirsch/oauth2"
)

type User struct {
	ID    string
	Name  string
	Token *oauth2.Token
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
	token := &oauth2.Token{}
	err = rows.Scan(
		&user.ID,
		&user.Name,
		&token.AccessToken,
		&token.RefreshToken,
		&token.Expiry,
	)
	if err != nil {
		return nil, err
	}
	user.Token = token

	if rows.Next() {
		return nil, fmt.Errorf("too many rows for %v", id)
	}

	return user, nil
}

func (db *DB) InsertUser(user *User) error {
	_, err := db.insertUserStmt.Exec(
		user.ID,
		user.Name,
		user.Token.AccessToken,
		user.Token.RefreshToken,
		user.Token.Expiry,
	)
	return err
}
