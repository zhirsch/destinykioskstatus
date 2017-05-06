package db

import (
	"time"

	"github.com/zhirsch/oauth2"
)

type (
	BungieMembershipID    string
	DestinyMembershipType int64
	DestinyMembershipID   string
	DestinyCharacterID    string
)

type DestinyCharacter struct {
	CharacterID DestinyCharacterID
	ClassName   string
}

type DestinyUser struct {
	MembershipType    DestinyMembershipType
	MembershipID      DestinyMembershipID
	DisplayName       string
	DestinyCharacters []*DestinyCharacter
}

type BungieUser struct {
	MembershipID BungieMembershipID
	DisplayName  string
	Token        *oauth2.Token
	DestinyUsers []*DestinyUser
}

func (db *DB) SelectBungieUser(membershipID BungieMembershipID) (*BungieUser, error) {
	// Select the BungieUser.
	var displayName, accessToken, refreshToken string
	var expiry time.Time
	if err := db.tables[tableBungieUsers].stmts[stmtSelect].QueryRow(string(membershipID)).Scan(&displayName, &accessToken, &refreshToken, &expiry); err != nil {
		return nil, err
	}
	bungieUser := &BungieUser{
		MembershipID: membershipID,
		DisplayName:  displayName,
		Token: &oauth2.Token{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			Expiry:       expiry,
		},
	}

	// Select the DestinyUsers.
	if destinyUsers, err := db.SelectDestinyUsers(bungieUser); err != nil {
		return nil, err
	} else {
		bungieUser.DestinyUsers = destinyUsers
	}

	return bungieUser, nil
}

func (db *DB) SelectDestinyUsers(bungieUser *BungieUser) ([]*DestinyUser, error) {
	var destinyUsers []*DestinyUser

	// Select the DestinyUsers.
	rows, err := db.tables[tableDestinyUsers].stmts[stmtSelect].Query(string(bungieUser.MembershipID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var membershipType int64
		var membershipID, displayName string
		if err := rows.Scan(&membershipType, &membershipID, &displayName); err != nil {
			return nil, err
		}
		destinyUsers = append(destinyUsers, &DestinyUser{
			MembershipType: DestinyMembershipType(membershipType),
			MembershipID:   DestinyMembershipID(membershipID),
			DisplayName:    displayName,
		})
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}

	// Select the DestinyCharacters.
	for _, destinyUser := range destinyUsers {
		if destinyCharacters, err := db.SelectDestinyCharacters(destinyUser); err != nil {
			return nil, err
		} else {
			destinyUser.DestinyCharacters = destinyCharacters
		}
	}

	return destinyUsers, nil
}

func (db *DB) SelectDestinyCharacters(destinyUser *DestinyUser) ([]*DestinyCharacter, error) {
	var destinyCharacters []*DestinyCharacter

	// Select the DestinyCharacters.
	rows, err := db.tables[tableDestinyCharacters].stmts[stmtSelect].Query(int64(destinyUser.MembershipType), string(destinyUser.MembershipID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var characterID, className string
		if err := rows.Scan(&characterID, &className); err != nil {
			return nil, err
		}
		destinyCharacters = append(destinyCharacters, &DestinyCharacter{
			CharacterID: DestinyCharacterID(characterID),
			ClassName:   className,
		})
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}

	return destinyCharacters, nil
}

func (db *DB) InsertBungieUser(bungieUser *BungieUser) error {
	_, err := db.tables[tableBungieUsers].stmts[stmtInsert].Exec(
		string(bungieUser.MembershipID),
		bungieUser.DisplayName,
		bungieUser.Token.AccessToken,
		bungieUser.Token.RefreshToken,
		bungieUser.Token.Expiry,
	)
	if err != nil {
		return err
	}
	for _, destinyUser := range bungieUser.DestinyUsers {
		if err := db.InsertDestinyUser(bungieUser.MembershipID, destinyUser); err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) InsertDestinyUser(bungieMembershipID BungieMembershipID, destinyUser *DestinyUser) error {
	_, err := db.tables[tableDestinyUsers].stmts[stmtInsert].Exec(
		int64(destinyUser.MembershipType),
		string(destinyUser.MembershipID),
		destinyUser.DisplayName,
		string(bungieMembershipID),
	)
	if err != nil {
		return err
	}
	for _, destinyCharacter := range destinyUser.DestinyCharacters {
		if err := db.InsertDestinyCharacter(destinyUser.MembershipType, destinyUser.MembershipID, destinyCharacter); err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) InsertDestinyCharacter(destinyMembershipType DestinyMembershipType, destinyMembershipID DestinyMembershipID, destinyCharacter *DestinyCharacter) error {
	_, err := db.tables[tableDestinyCharacters].stmts[stmtInsert].Exec(
		string(destinyCharacter.CharacterID),
		destinyCharacter.ClassName,
		int64(destinyMembershipType),
		string(destinyMembershipID),
	)
	return err
}
